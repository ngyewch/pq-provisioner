package test

import (
	"fmt"
	"os"
	"path/filepath"

	docker "github.com/fsouza/go-dockerclient"
)

const (
	customOpenSshServerRepoTag = "pq-provisioner/openssh-server:latest"
)

type OpenSSHServerFactory struct {
	helper *DockerHelper
}

func NewOpenSSHServerFactory(helper *DockerHelper) (*OpenSSHServerFactory, error) {
	factory := &OpenSSHServerFactory{
		helper: helper,
	}

	err := helper.BuildImage(customOpenSshServerRepoTag,
		os.DirFS(filepath.Join("resources", "dockerBuildContexts", "openssh-server")))
	if err != nil {
		return nil, err
	}

	return factory, nil
}

func (factory *OpenSSHServerFactory) StartSshHost(hostname string, username string, publicKey string, hostPort int) (*docker.Container, error) {
	createContainerOptions := docker.CreateContainerOptions{
		Config: &docker.Config{
			Image:        customOpenSshServerRepoTag,
			AttachStdout: true,
			AttachStderr: true,
			PortSpecs: []string{
				"2222/tcp",
			},
			Env: []string{
				fmt.Sprintf("PUBLIC_KEY=%s", publicKey),
				fmt.Sprintf("USER_NAME=%s", username),
			},
			Hostname: hostname,
		},
	}
	if hostPort > 0 {
		createContainerOptions.Config.PortSpecs = []string{
			"2222/tcp",
		}
		createContainerOptions.HostConfig = &docker.HostConfig{
			PortBindings: map[docker.Port][]docker.PortBinding{
				"2222/tcp": {
					{
						HostIP:   "127.0.0.1",
						HostPort: fmt.Sprintf("%d/tcp", hostPort),
					},
				},
			},
		}
	}
	container, err := factory.helper.CreateContainer(createContainerOptions)
	if err != nil {
		return nil, err
	}
	err = factory.helper.StartContainer(container.ID, &docker.HostConfig{
		AutoRemove: true,
	})
	if err != nil {
		return nil, err
	}

	return container, nil
}
