package test

import (
	"fmt"
	"os"
	"os/signal"
	"testing"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/google/uuid"
)

type Env struct {
	helper  *DockerHelper
	network *docker.Network
}

func NewEnv(t *testing.T) (*Env, error) {
	helper, err := NewDockerHelper()
	if err != nil {
		return nil, err
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			t.Log("Interrupted by user")
			_ = helper.Close()
			os.Exit(1)
		}
	}()

	id, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}

	network, err := helper.CreateNetwork(fmt.Sprintf("network_%s", id))
	if err != nil {
		return nil, err
	}

	return &Env{
		helper:  helper,
		network: network,
	}, nil
}

func (env *Env) Helper() *DockerHelper {
	return env.helper
}

func (env *Env) Network() *docker.Network {
	return env.network
}

func (env *Env) Close() error {
	if env.helper != nil {
		err := env.helper.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
