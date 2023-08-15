package main

import (
	"fmt"
	"net"
)

func CheckNetPorts(ports []int) bool {
	for _, port := range ports {
		if !CheckNet(port) {
			return false
		}
	}
	return true
}

func CheckNet(port int) bool {
	destAddr := fmt.Sprintf("127.0.0.1:%d", port)
	conn, err := net.Dial("tcp", destAddr)
	defer func() {
		if err == nil {
			conn.Close()
		}
	}()
	if err != nil {
		fmt.Println("network check ", destAddr, " error:", err)
		return false
	}
	return true
}
