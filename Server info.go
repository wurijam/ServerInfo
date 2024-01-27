package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
)

func main() {

	hostname, err := os.Hostname()
	if err != nil {
		fmt.Println("Error getting hostname:", err)
		return
	}

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println("Error getting IP address:", err)
		return
	}

	var ipAddress string
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			ipAddress = ipnet.IP.String()
			break
		}
	}

	cpu, err := cpu.Percent(time.Second, false)
	if err != nil {
		fmt.Println("Error getting CPU usage:", err)
		return
	}

	vm, err := mem.VirtualMemory()
	if err != nil {
		fmt.Println("Error getting RAM usage:", err)
		return
	}
	ramUsage := float64(vm.Used) / float64(vm.Total) * 100.0

	partitions, err := disk.Partitions(false)
	if err != nil {
		fmt.Println("Error getting storage info:", err)
		return
	}

	fmt.Println("System Information:")
	fmt.Println("Hostname:", hostname)
	fmt.Println("IP Address:", ipAddress)

	fmt.Printf("\nCPU Usage:\n")
	for i, usage := range cpu {
		fmt.Printf("  Core %d: %.2f%%\n", i, usage)
	}

	fmt.Printf("\nRAM Usage: %.2f%%\n", ramUsage)

	fmt.Println("\nStorage Info:")

	for _, partition := range partitions {
		if partition.Fstype != "" && partition.Mountpoint != "" {
			usage, err := disk.Usage(partition.Mountpoint)
			if err != nil {
				fmt.Printf("Error getting usage for %s: %s\n", partition.Mountpoint, err)
				continue
			}
			fmt.Printf("  Mountpoint: %v\n", usage.Path)
			fmt.Printf("    Total: %v GB\n    Used: %v GB\n    Free: %v GB\n", usage.Total/1024/1024/1024, usage.Used/1024/1024/1024, usage.Free/1024/1024/1024)
		}
	}

	hostInfo, err := host.Info()
	if err != nil {
		fmt.Println("Error getting uptime:", err)
		return
	}
	uptimeDuration := time.Duration(hostInfo.Uptime) * time.Second
	fmt.Println("\nUptime:", uptimeDuration)
}
