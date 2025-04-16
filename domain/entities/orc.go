package entities

import "github.com/docker/go-connections/nat"

type OrcConfig struct {
	Name          string
	AttachStdin   bool
	AttachStdout  bool
	AttachStderr  bool
	ExposedPorts  nat.PortSet
	Cmd           []string
	Image         string
	CPU           float64
	Memory        int64
	Disk          int64
	Env           []string
	RestartPolicy string
}

func NewOrcConfig(t *Task) OrcConfig {
	return OrcConfig{
		Name:          t.Name,
		AttachStdin:   false,
		AttachStdout:  false,
		AttachStderr:  false,
		ExposedPorts:  t.ExposedPorts,
		Cmd:           nil,
		Image:         t.Image,
		CPU:           0,
		Memory:        0,
		Disk:          0,
		Env:           nil,
		RestartPolicy: "",
	}
}
