package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	_ "go.uber.org/automaxprocs"
)

const (
	ENV_REPORT_SERVER_ADDR = "report_server_addr" // 上报服务器地址
	ENV_SERVER_PORT        = "server_port"        // 服务端口
	ENV_PRE_IP_FORMAT      = "pre_ip_format"      // ip前缀
	ENV_CONFIG_FILE        = "config_file"
	ENV_CHECK_PORT         = "check_ports"
	ENV_SERVER_PATH        = "server_path"
	ENV_TARGET_CONTAINER   = "target_container_name"
)

func main() {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGTERM)
	go func() {
		<-signalChannel
		log.Println("Received SIGTERM. Cleaning up...")
		// 这里可以执行一些清理操作，然后退出程序
		os.Exit(0)
	}()

	preIpFormat := os.Getenv(ENV_PRE_IP_FORMAT)
	if preIpFormat == "" {
		preIpFormat = "172.31"
	}

	configFile := os.Getenv(ENV_CONFIG_FILE)
	if configFile == "" {
		configFile = "/config.yaml"
	}

	checkPortList := os.Getenv(ENV_CHECK_PORT)
	if checkPortList == "" {
		checkPortList = "10001,10002,10003"
	}
	strcheckPorts := strings.Split(checkPortList, ",")
	checkPorts := make([]int, 0, len(strcheckPorts)) // 创建一个整数切片
	for _, str := range strcheckPorts {
		if isNumber(str) {
			num, err := strconv.Atoi(str)
			if err != nil {
				fmt.Printf("Error converting %s to int: %v\n", str, err)
			} else {
				checkPorts = append(checkPorts, num)
			}
		}
	}

	log.Println("check current network...")
	log.Println("preIpFormat:", preIpFormat)
	if GetCurrentIp(preIpFormat) == "" {
		log.Println("Not in the right network")
		os.Exit(0)
	}

	log.Println("check current config file...")
	if ReadAndChangeConfig(configFile, GetCurrentIp(preIpFormat)) {
		RestartDockerContainer(os.Getenv(ENV_TARGET_CONTAINER))
	}

	log.Println("wait 30 seconds and check the port...")
	time.Sleep(30 * time.Second)

	tryCount := 0
	checkCount := 0
	log.Println("begin check logic, loop with 3 second...")
	for {
		if CheckNetPorts(checkPorts) {
			reportServerAddr := os.Getenv(ENV_REPORT_SERVER_ADDR)
			currentAddr := fmt.Sprintf("%s:%s,%d", GetCurrentIp(preIpFormat), os.Getenv(ENV_SERVER_PORT), runtime.GOMAXPROCS(0))
			remoteServerPaths := os.Getenv(ENV_SERVER_PATH)
			remoteServerPath := strings.Split(remoteServerPaths, ",")
			for _, v := range remoteServerPath {
				if v == "" {
					continue
				}
				ReportOnlineServer(currentAddr, v, reportServerAddr)
			}
			if checkCount%20 == 0 {
				log.Println("The current check is correct", checkPortList)
			}
			checkCount++
		} else {
			checkCount = 0
			tryCount++
			log.Println("check failed... try count:", tryCount)

			if tryCount > 5 {
				log.Println("network check failed,try restart container and kill current task...")
				RestartDockerContainer(os.Getenv(ENV_TARGET_CONTAINER))
				os.Exit(0)
			}
		}

		time.Sleep(3 * time.Second)
	}
}
