package test

import (
	"archive/tar"
	"bytes"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	docker "github.com/fsouza/go-dockerclient"
)

type DockerHelper struct {
	client     *docker.Client
	containers []*docker.Container
	networks   []*docker.Network
}

func NewDockerHelper() (*DockerHelper, error) {
	client, err := docker.NewClientFromEnv()
	if err != nil {
		return nil, err
	}
	return &DockerHelper{
		client: client,
	}, nil
}

func (helper *DockerHelper) PullImage(repoTag string) error {
	return helper.client.PullImage(docker.PullImageOptions{
		Repository:   repoTag,
		OutputStream: os.Stdout,
	}, docker.AuthConfiguration{})
}

func (helper *DockerHelper) BuildImage(imageName string, dockerBuildContextFs fs.FS) error {
	inputBuf := bytes.NewBuffer(nil)
	tr := tar.NewWriter(inputBuf)

	err := createTar(tr, dockerBuildContextFs, nil)
	if err != nil {
		return err
	}

	err = tr.Close()
	if err != nil {
		return err
	}

	err = helper.client.BuildImage(docker.BuildImageOptions{
		Name:         imageName,
		InputStream:  inputBuf,
		OutputStream: os.Stdout,
	})
	if err != nil {
		return err
	}

	return nil
}

func (helper *DockerHelper) CreateNetwork(networkId string) (*docker.Network, error) {
	network, err := helper.client.CreateNetwork(docker.CreateNetworkOptions{
		Name: networkId,
	})
	if err != nil {
		return nil, err
	}

	helper.networks = append(helper.networks, network)

	return network, nil
}

func (helper *DockerHelper) ConnectNetworks(container *docker.Container, networks ...*docker.Network) error {
	for _, network := range networks {
		err := helper.client.ConnectNetwork(network.ID, docker.NetworkConnectionOptions{
			Container: container.ID,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (helper *DockerHelper) CreateContainer(options docker.CreateContainerOptions) (*docker.Container, error) {
	c, err := helper.client.CreateContainer(options)
	if err == nil {
		helper.containers = append(helper.containers, c)
	}
	return c, err
}

func (helper *DockerHelper) StartContainer(id string, hostConfig *docker.HostConfig) error {
	return helper.client.StartContainer(id, hostConfig)
}

func (helper *DockerHelper) Close() error {
	for _, container := range helper.containers {
		err := helper.client.StopContainer(container.ID, 15)
		if err != nil {
			return err
		}

		err = helper.client.RemoveContainer(docker.RemoveContainerOptions{
			ID: container.ID,
		})
		if err != nil {
			return err
		}
	}

	for _, network := range helper.networks {
		err := helper.client.RemoveNetwork(network.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func createTar(tr *tar.Writer, filesystem fs.FS, vars any) error {
	return fs.WalkDir(filesystem, ".", func(path string, entry fs.DirEntry, err error) error {
		if path == "." {
			return nil
		}
		fi, err := entry.Info()
		if err != nil {
			return err
		}
		if entry.IsDir() {
			t := time.Now()
			err = tr.WriteHeader(&tar.Header{
				Name:       path,
				Size:       0,
				Mode:       int64(fi.Mode()),
				ModTime:    t,
				AccessTime: t,
				ChangeTime: t,
			})
			if err != nil {
				return err
			}
		} else if strings.HasSuffix(path, ".tmpl") {
			tmpl, err := template.New(filepath.Base(path)).ParseFS(filesystem, path)
			if err != nil {
				return err
			}
			outputBuf := bytes.NewBuffer(nil)
			err = tmpl.Execute(outputBuf, vars)
			if err != nil {
				return err
			}
			actualPath := path[0 : len(path)-5]
			contentBytes := outputBuf.Bytes()
			t := time.Now()
			err = tr.WriteHeader(&tar.Header{
				Name:       actualPath,
				Size:       int64(len(contentBytes)),
				Mode:       int64(fi.Mode()),
				ModTime:    t,
				AccessTime: t,
				ChangeTime: t,
			})
			if err != nil {
				return err
			}
			_, err = tr.Write(contentBytes)
			if err != nil {
				return err
			}
		} else {
			f, err := filesystem.Open(path)
			if err != nil {
				return err
			}
			defer func(f fs.File) {
				_ = f.Close()
			}(f)
			contentBytes, err := io.ReadAll(f)
			if err != nil {
				return err
			}
			t := time.Now()
			err = tr.WriteHeader(&tar.Header{
				Name:       path,
				Size:       int64(len(contentBytes)),
				Mode:       int64(fi.Mode()),
				ModTime:    t,
				AccessTime: t,
				ChangeTime: t,
			})
			if err != nil {
				return err
			}
			_, err = tr.Write(contentBytes)
			if err != nil {
				return err
			}
		}
		return nil
	})
}
