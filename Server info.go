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

func getIPAddress() string {
	addrs, _ := net.InterfaceAddrs()

	var ipAddress string
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			ipAddress = ipnet.IP.String()
			break
		}
	}

	return ipAddress
}

func getHostname() string {
	hostname, _ := os.Hostname()
	return hostname
}

func getCPUUsage() []float64 {
	percentages, _ := cpu.Percent(time.Second, false)
	return percentages
}

func getRAMUsage() float64 {
	vm, _ := mem.VirtualMemory()
	return float64(vm.Used) / float64(vm.Total) * 100.0
}

func getStorageInfo() []disk.UsageStat {
	partitions, _ := disk.Partitions(false)

	var storageInfo []disk.UsageStat
	for _, partition := range partitions {
		if partition.Fstype != "" && partition.Mountpoint != "" {
			usage, _ := disk.Usage(partition.Mountpoint)
			storageInfo = append(storageInfo, *usage)
		}
	}

	return storageInfo
}

func getUptime() time.Duration {
	hostInfo, _ := host.Info()
	return time.Duration(hostInfo.Uptime) * time.Second
}

func main() {

	fmt.Println("System Information:")
	fmt.Println("Hostname:", getHostname())
	fmt.Println("IP Address:", getIPAddress())

	fmt.Printf("\nCPU Usage:\n")
	for i, usage := range getCPUUsage() {
		fmt.Printf("  Core %d: %.2f%%\n", i, usage)
	}

	fmt.Printf("\nRAM Usage: %.2f%%\n", getRAMUsage())

	fmt.Println("\nStorage Info:")
	for _, info := range getStorageInfo() {
		fmt.Printf("  Mountpoint: %v\n", info.Path)
		fmt.Printf("    Total: %v GB\n    Used: %v GB\n    Free: %v GB\n", info.Total/1024/1024/1024, info.Used/1024/1024/1024, info.Free/1024/1024/1024)
	}

	fmt.Println("\nUptime:", getUptime())
}
