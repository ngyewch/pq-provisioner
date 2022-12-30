package ssh_tunnel

import (
	"fmt"
	slog "github.com/go-eden/slf4go"
	"github.com/mitchellh/go-homedir"
	"golang.org/x/crypto/ssh"
	"io"
	"net"
	"os"
)

var (
	log = slog.GetLogger()
)

func GetSignerFromFile(path string) (ssh.Signer, error) {
	expandedPath, err := homedir.Expand(path)
	if err != nil {
		return nil, err
	}
	privateKeyBytes, err := os.ReadFile(expandedPath)
	if err != nil {
		return nil, err
	}
	signer, err := ssh.ParsePrivateKey(privateKeyBytes)
	if err != nil {
		return nil, err
	}
	return signer, nil
}

func NewSshClient(host string, port int, user string, identityFile string) (*ssh.Client, error) {
	signer, err := GetSignerFromFile(identityFile)
	if err != nil {
		return nil, err
	}
	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	if port == 0 {
		port = 22
	}
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", host, port), sshConfig)
	if err != nil {
		return nil, err
	}
	return client, nil
}

type LocalPortForwarder struct {
	sshClient  *ssh.Client
	remoteAddr string
	listener   net.Listener
}

type Session struct {
	remoteAddr string
	local      net.Conn
	remote     net.Conn
}

func NewLocalPortForwarder(sshClient *ssh.Client, localAddr string, remoteAddr string) (*LocalPortForwarder, error) {
	listener, err := net.Listen("tcp", localAddr)
	if err != nil {
		return nil, err
	}

	return &LocalPortForwarder{
		sshClient:  sshClient,
		remoteAddr: remoteAddr,
		listener:   listener,
	}, nil
}

func (lpf *LocalPortForwarder) LocalAddr() *net.TCPAddr {
	return lpf.listener.Addr().(*net.TCPAddr)
}

func (lpf *LocalPortForwarder) getDescriptor() string {
	return fmt.Sprintf("%d -> %s", lpf.LocalAddr().Port, lpf.remoteAddr)
}

func (lpf *LocalPortForwarder) Start() {
	go func() {
		for {
			local, err := lpf.listener.Accept()
			if err != nil {
				opErr, ok := err.(*net.OpError)
				if ok {
					if opErr.Err.Error() != "use of closed network connection" {
						log.Errorf("[%s] %s", lpf.getDescriptor(), err)
					}
				}
				break
			}
			remote, err := lpf.sshClient.Dial("tcp", lpf.remoteAddr)
			if err != nil {
				log.Errorf("[%s] %s", lpf.getDescriptor(), err)
			} else {
				session := newSession(lpf.remoteAddr, local, remote)
				go func() {
					session.Start()
				}()
			}
		}
	}()
}

func (lpf *LocalPortForwarder) Close() error {
	err := lpf.listener.Close()
	if err != nil {
		return err
	}
	return nil
}

func newSession(remoteAddr string, local net.Conn, remote net.Conn) *Session {
	return &Session{
		remoteAddr: remoteAddr,
		local:      local,
		remote:     remote,
	}
}

func (s *Session) Start() {
	descriptor := fmt.Sprintf("%s -> %s", s.local.RemoteAddr(), s.remoteAddr)

	log.Debugf("[%s] started", descriptor)
	defer s.local.Close()
	defer s.remote.Close()

	chDone := make(chan bool)
	go func() {
		_, err := io.Copy(s.local, s.remote)
		if err != nil {
			log.Errorf("[%s / outgoing] %s", descriptor, err)
		}
		chDone <- true
	}()
	go func() {
		_, err := io.Copy(s.remote, s.local)
		if err != nil {
			log.Errorf("[%s / incoming] %s", descriptor, err)
		}
		chDone <- true
	}()
	<-chDone
	log.Debugf("[%s] ended", descriptor)
}
