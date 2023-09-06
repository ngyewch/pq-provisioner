package test

import (
	"fmt"
	docker "github.com/fsouza/go-dockerclient"
	"os"
	"path/filepath"
)

const (
	customPostgresRepoTag = "pq-provisioner/postgres:latest"
)

type PostgresFactory struct {
	helper *DockerHelper
}

func NewPostgresFactory(helper *DockerHelper) (*PostgresFactory, error) {
	factory := &PostgresFactory{
		helper: helper,
	}

	err := helper.BuildImage(customPostgresRepoTag,
		os.DirFS(filepath.Join("resources", "dockerBuildContexts", "postgres")))
	if err != nil {
		return nil, err
	}

	return factory, nil
}

func (factory *PostgresFactory) Start(hostname string, postgresPassword string, hostPort int) (*docker.Container, error) {
	createContainerOptions := docker.CreateContainerOptions{
		Config: &docker.Config{
			Image:        customPostgresRepoTag,
			AttachStdout: true,
			AttachStderr: true,
			PortSpecs: []string{
				"5432/tcp",
			},
			Env: []string{
				fmt.Sprintf("POSTGRES_PASSWORD=%s", postgresPassword),
			},
			Hostname: hostname,
		},
	}
	if hostPort > 0 {
		createContainerOptions.Config.PortSpecs = []string{
			"5432/tcp",
		}
		createContainerOptions.HostConfig = &docker.HostConfig{
			PortBindings: map[docker.Port][]docker.PortBinding{
				"5432/tcp": {
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
