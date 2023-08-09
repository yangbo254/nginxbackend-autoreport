package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	_ "go.uber.org/automaxprocs"
)

func ReportOnlineServer(currentAddr, remoteServerPath, reportServerAddr string) []string {
	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s/getonlinereportserver", reportServerAddr), nil)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	req.Header.Set("Remote-Server-Addr", currentAddr)
	req.Header.Set("Server-Path", remoteServerPath)
	var t int64 = 1000
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(t)*time.Millisecond)
	defer cancel()
	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	return nil
}

func GetCurrentIp(prefix string) string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Println("Error:", err)
		return ""
	}

	for _, addr := range addrs {
		// 检查地址是否为 IP 地址，并且不是回环地址
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				if strings.HasPrefix(ipNet.IP.String(), prefix) {
					return ipNet.IP.String()
				}
			}
		}
	}
	return ""
}

func ReadAndChangeConfig(filePath, currentIp string) bool {
	filename := filePath // 替换成你要操作的文件名

	// 读取文件内容
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Println("Error reading file:", err)
		return false
	}

	// 将内容转换为字符串
	fileContent := string(content)

	// 检查内容是否有特定字符串
	if !strings.Contains(fileContent, "172.31.10.7") {
		log.Println("config file is ok")
		return false
	}

	// 替换内容
	newContent := strings.Replace(fileContent, "172.31.10.7", currentIp, -1)

	// 将新内容写入文件
	err = ioutil.WriteFile(filename, []byte(newContent), 0644)
	if err != nil {
		log.Println("Error writing file:", err)
		return false
	}

	log.Println("File content replaced and saved successfully.")
	return true
}

func main() {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGTERM)
	go func() {
		<-signalChannel
		log.Println("Received SIGTERM. Cleaning up...")
		// 这里可以执行一些清理操作，然后退出程序
		os.Exit(0)
	}()

	log.Println("check current network...")
	if GetCurrentIp("172.31") == "" {
		log.Println("Not in the right network")
		os.Exit(0)
	}

	log.Println("check current config file...")
	if ReadAndChangeConfig("config/config.yaml", GetCurrentIp("172.31")) {
		RestartDockerContainer(os.Getenv("target_container_name"))
	}

	log.Println("wait 30 seconds and check the port...")
	time.Sleep(30 * time.Second)

	tryCount := 0
	checkCount := 0
	log.Println("begin check logic, loop with 3 second...")
	for {
		if CheckNet(10001) && CheckNet(10002) && CheckNet(10003) {
			reportServerAddr := os.Getenv("report_server_addr")
			currentAddr := fmt.Sprintf("%s:%s,%d", GetCurrentIp("172.31"), os.Getenv("server_port"), runtime.GOMAXPROCS(0))
			remoteServerPath := os.Getenv("server_path")
			ReportOnlineServer(currentAddr, remoteServerPath, reportServerAddr)
			if checkCount%20 == 0 {
				log.Println("The current check is correct")
			}
			checkCount++
		} else {
			checkCount = 0
			tryCount++
			log.Println("check failed... try count:", tryCount)

			if tryCount > 5 {
				log.Println("network check failed,try restart container and kill current task...")
				RestartDockerContainer(os.Getenv("target_container_name"))
				os.Exit(0)
			}
		}

		time.Sleep(3 * time.Second)
	}
}
