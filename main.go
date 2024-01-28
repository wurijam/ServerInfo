package main

import (
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
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

func collectSystemInfo() (*SystemInfo, error) {
	hostname, err := os.Hostname()
	if err != nil {
		fmt.Println("Failed to get hostname: %w", err)
	}

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println("Failed to get IP address: %w", err)
	}

	var ipAddress string
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			ipAddress = ipnet.IP.String()
			break
		}
	}

	cpuUsage, err := cpu.Percent(time.Second, false)
	if err != nil {
		fmt.Println("Failed to get CPU usage: %w", err)
	}

	vm, err := mem.VirtualMemory()
	if err != nil {
		fmt.Println("Failed to get RAM usage: %w", err)
	}
	ramUsage := float64(vm.Used) / float64(vm.Total) * 100.0

	partitions, err := disk.Partitions(false)
	if err != nil {
		fmt.Println("Failed to get storage info: %w", err)
	}

	storageInfo := make([]StorageInfo, 0)
	for _, partition := range partitions {
		if partition.Fstype != "" && partition.Mountpoint != "" {
			usage, err := disk.Usage(partition.Mountpoint)
			if err != nil {
				return nil, fmt.Errorf("failed to get usage for %s: %w", partition.Mountpoint, err)
			}
			storageInfo = append(storageInfo, StorageInfo{
				Mountpoint: partition.Mountpoint,
				Total:      usage.Total,
				Used:       usage.Used,
				Free:       usage.Free,
			})
		}
	}

	hostInfo, err := host.Info()
	if err != nil {
		fmt.Println("Failed to get uptime: %w", err)
	}
	uptimeDuration := time.Duration(hostInfo.Uptime) * time.Second

	return &SystemInfo{
		Hostname:  hostname,
		IPAddress: ipAddress,
		CPUUsage:  cpuUsage,
		RAMUsage:  ramUsage,
		Storage:   storageInfo,
		Uptime:    uptimeDuration,
	}, nil
}

func main() {
	listener, err := net.Listen("tcp", "wurijam.se:9999")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Server listening on wurijam.se:9999")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting request:", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	systemInfo, err := collectSystemInfo()
	if err != nil {
		fmt.Println("Error collecting system information:", err)
		return
	}

	encoder := gob.NewEncoder(conn)
	err = encoder.Encode(systemInfo)
	if err != nil {
		fmt.Println("Error encoding and sending system information:", err)
	}
}
