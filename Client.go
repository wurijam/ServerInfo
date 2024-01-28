package main

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type SystemInfo struct {
	Hostname  string
	IPAddress string
	CPUUsage  []float64
	RAMUsage  float64
	Storage   []StorageInfo
	Uptime    time.Duration
}

type StorageInfo struct {
	Mountpoint string
	Total      uint64
	Used       uint64
	Free       uint64
}

func main() {
	serverAddresses := []string{"wurijam.se:9999", "test"}

	fmt.Println("Available servers:")
	for i, address := range serverAddresses {
		fmt.Printf("%d. %s\n", i+1, address)
	}
	fmt.Println("0. Request all servers")

	selectedServers := selectServers(len(serverAddresses))

	infoChannel := make(chan SystemInfo)

	for _, index := range selectedServers {
		if index == 0 {
			for _, address := range serverAddresses {
				go requestSystemInfo(address, infoChannel)
			}
		} else {
			address := serverAddresses[index-1]
			go requestSystemInfo(address, infoChannel)
		}
	}

	for range selectedServers {
		systemInfo := <-infoChannel
		printSystemInfo(systemInfo)
	}
}

func selectServers(numServers int) []int {
	var selectedServers []int

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("Enter the number of the server you want to request to: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input == "0" {
			for i := 0; i <= numServers; i++ {
				selectedServers = append(selectedServers, i)
			}
			break
		}

		indices := strings.Split(input, ",")
		for _, indexStr := range indices {
			indexStr = strings.TrimSpace(indexStr)
			index, err := strconv.Atoi(indexStr)
			if err != nil || index < 0 || index > numServers {
				fmt.Println("Invalid input")
				selectedServers = nil
				break
			}
			selectedServers = append(selectedServers, index)
		}

		if len(selectedServers) > 0 {
			break
		}
	}

	return selectedServers
}

func requestSystemInfo(address string, infoChannel chan<- SystemInfo) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Printf("Error requsting server %s: %s\n", address, err)
		infoChannel <- SystemInfo{}
		return
	}
	defer conn.Close()

	var systemInfo SystemInfo
	decoder := gob.NewDecoder(conn)
	err = decoder.Decode(&systemInfo)
	if err != nil {
		fmt.Printf("Error receiving information from %s: %s\n", address, err)
		infoChannel <- SystemInfo{}
		return
	}

	infoChannel <- systemInfo
}

func printSystemInfo(systemInfo SystemInfo) {
	if systemInfo.Hostname == "" {
		fmt.Println("Failed to retrieve system information")
		return
	}

	fmt.Println("System Information:")
	fmt.Println("Hostname:", systemInfo.Hostname)
	fmt.Println("IP Address:", systemInfo.IPAddress)

	fmt.Printf("\nCPU Usage:\n")
	for i, usage := range systemInfo.CPUUsage {
		fmt.Printf("  Core %d: %.2f%%\n", i, usage)
	}

	fmt.Printf("\nRAM Usage: %.2f%%\n", systemInfo.RAMUsage)

	fmt.Println("\nStorage Info:")
	for _, storage := range systemInfo.Storage {
		fmt.Printf("  Mountpoint: %v\n", storage.Mountpoint)
		fmt.Printf("    Total: %v GB\n    Used: %v GB\n    Free: %v GB\n", storage.Total/1024/1024/1024, storage.Used/1024/1024/1024, storage.Free/1024/1024/1024)

	}

	fmt.Println("\nUptime:", systemInfo.Uptime)
}
