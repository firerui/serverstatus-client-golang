package client

import "log"
import "github.com/shirou/gopsutil/v3/cpu"
import "github.com/shirou/gopsutil/v3/host"
import "github.com/shirou/gopsutil/v3/load"

type SystemInfo struct {
	OSName        string
	OSArch        string
	OSFamily      string
	OSRelease     string
	KernelVersion string
	CPUNum        int
	CPUBrand      string
	CPUVenderID   string
	HostName      string
	Uptime        uint64
	Load1         float64
	Load5         float64
	Load15        float64
}

func GetSysInfo() SystemInfo {
	// 获取 host info 信息
	hostInfo, err := host.Info()
	if err != nil {
		log.Fatal("Error getting Host.Info:", err)
	}
	// 获取 cpu info 信息
	cpuInfo, err := cpu.Info()
	if err != nil {
		log.Fatal("Error getting cpu.Info:", err)
	}
	// 获取 cpu 核心数
	cpuCounts, err := cpu.Counts(true)
	if err != nil {
		log.Fatal("Error getting cpu.Counts:", err)
	}
	// 获取 uptime
	uptime, _ := host.Uptime()
	// 获取 avg
	avg, err := load.Avg()
	if err != nil {
		log.Fatal("Error getting load.Avg:", err)
	}

	return SystemInfo{
		OSName:        hostInfo.OS,
		OSArch:        hostInfo.KernelArch,
		OSFamily:      hostInfo.PlatformFamily,
		OSRelease:     hostInfo.PlatformVersion,
		KernelVersion: hostInfo.KernelVersion,
		CPUNum:        cpuCounts,
		CPUBrand:      cpuInfo[0].ModelName,
		CPUVenderID:   cpuInfo[0].VendorID,
		HostName:      hostInfo.Hostname,
		Uptime:        uptime,
		Load1:         avg.Load1,
		Load5:         avg.Load5,
		Load15:        avg.Load15,
	}
}
