package docker

import (
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"orc/domain/entities"
)

type Docker struct {
	Client *client.Client
	Config entities.OrcConfig
}

type Result struct {
	Error       error
	Action      string
	ContainerID string
	Result      string
}

type InspectResponse struct {
	Error     error
	Container *types.ContainerJSON
}

func NewDocker(config entities.OrcConfig) (*Docker, error) {
	dc, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}
	return &Docker{
		Client: dc,
		Config: config,
	}, nil
}
