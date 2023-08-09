package main

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func RestartDockerContainer(containerName string) {
	if containerName == "" {
		log.Println("containerName is empty")
		os.Exit(1)
	}
	os.Setenv("DOCKER_API_VERSION", "1.41")
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)

	}

	// 获取正在运行的容器列表
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{All: true})
	if err != nil {
		panic(err)
	}

	// 打印容器信息
	log.Println("Running Containers:")
	for _, containerNode := range containers {
		log.Printf("%s\t%s\t%s\n", containerNode.ID[:10], containerNode.State, containerNode.Names[0])
		if strings.Contains(containerNode.Names[0], containerName) && strings.ToLower(containerNode.State) == "running" {
			// 重启容器
			if err := cli.ContainerRestart(context.Background(), containerNode.ID, container.StopOptions{}); err != nil {
				panic(err)
			}
		}
	}
}
