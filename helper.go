package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func isNumber(str string) bool {
	_, err := strconv.ParseFloat(str, 64)
	return err == nil
}

func ReportOnlineServer(currentAddr, remoteServerPath, reportServerAddr string) []string {
	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s/autoreport", reportServerAddr), nil)
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
				log.Println("IP:", ipNet.IP.String())
				if strings.HasPrefix(ipNet.IP.String(), prefix) {
					return ipNet.IP.String()
				}
			}
		}
	}
	return ""
}

func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}

func ReadAndChangeConfig(filePath, currentIp string) bool {
	filename := filePath // 替换成你要操作的文件名

	// 读取文件内容
	if !fileExists(filename) {
		log.Println("file is not found:", filename)
		return false
	}
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
