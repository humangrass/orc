package docker

import (
	"context"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"io"
	"log"
	"math"
	"os"
)

func (d *Docker) Run() Result {
	ctx := context.Background()
	reader, err := d.Client.ImagePull(ctx, d.Config.Image, image.PullOptions{})
	if err != nil {
		log.Printf("Error pulling image %s: %v\n", d.Config.Image, err)
		return Result{Error: err}
	}
	_, err = io.Copy(os.Stdout, reader)
	if err != nil {
		return Result{Error: err}
	}

	restartPolicy := container.RestartPolicy{
		Name: container.RestartPolicyMode(d.Config.RestartPolicy),
	}
	resources := container.Resources{
		Memory:   d.Config.Memory,
		NanoCPUs: int64(d.Config.CPU * math.Pow(10, 9)),
	}

	cc := container.Config{
		Image:        d.Config.Image,
		Tty:          false,
		Env:          d.Config.Env,
		ExposedPorts: d.Config.ExposedPorts,
	}

	hc := container.HostConfig{
		RestartPolicy:   restartPolicy,
		Resources:       resources,
		PublishAllPorts: true,
	}

	resp, err := d.Client.ContainerCreate(ctx, &cc, &hc, nil, nil, d.Config.Name)
	if err != nil {
		log.Printf("Error creating container %s: %v\n", d.Config.Image, err)
		return Result{Error: err}
	}

	err = d.Client.ContainerStart(ctx, resp.ID, container.StartOptions{})
	if err != nil {
		log.Printf("Error starting container %s - %s: %v\n", d.Config.Image, resp.ID, err)
		return Result{Error: err}
	}

	result := Result{
		Action:      "start",
		Error:       nil,
		ContainerID: resp.ID,
	}

	out, err := d.Client.ContainerLogs(
		ctx,
		resp.ID,
		container.LogsOptions{ShowStdout: true, ShowStderr: true},
	)
	if err != nil {
		log.Printf("Error getting container logs %s: %v\n", resp.ID, err)
		result.Error = err
		return result
	}

	_, err = stdcopy.StdCopy(os.Stdout, os.Stderr, out)
	if err != nil {
		result.Error = err
		return result
	}

	result.Result = "success"
	return result
}

func (d *Docker) Stop(id string) Result {
	log.Printf("Attempting to stop container %s\n", id)
	ctx := context.Background()
	err := d.Client.ContainerStop(ctx, id, container.StopOptions{})
	if err != nil {
		log.Printf("Error stopping container %s: %v\n", id, err)
		return Result{Error: err}
	}

	err = d.Client.ContainerRemove(ctx, id, container.RemoveOptions{
		RemoveVolumes: true,
		RemoveLinks:   false,
		Force:         false,
	})
	if err != nil {
		log.Printf("Error removing container %s: %v\n", id, err)
		return Result{Error: err}
	}

	return Result{
		Error:       nil,
		Action:      "stop",
		ContainerID: id,
		Result:      "success",
	}
}

func (d *Docker) Inspect(containerID string) InspectResponse {
	dc, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return InspectResponse{Error: err}
	}
	ctx := context.Background()
	resp, err := dc.ContainerInspect(ctx, containerID)
	if err != nil {
		return InspectResponse{Error: err}
	}
	return InspectResponse{Container: &resp}
}
