package client

import (
	"fmt"
	"github.com/shirou/gopsutil/v3/mem"
	"log"
	"math"
	"time"
)
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
	MemoryTotal   uint64
	MemoryUsed    uint64
	SwapTotal     uint64
	SwapUsed      uint64
}

//var sysInfo SystemInfo

func GetSysInfo() SystemInfo {
	// 获取 host info 信息
	var sysInfo SystemInfo
	hostInfo, err := host.Info()
	if err != nil {
		log.Fatal("Error getting Host.Info:", err)
	} else {
		sysInfo.OSName = hostInfo.OS
		sysInfo.OSArch = hostInfo.KernelArch
		sysInfo.OSFamily = hostInfo.PlatformFamily
		sysInfo.OSRelease = hostInfo.KernelVersion
		sysInfo.HostName = hostInfo.Hostname
	}
	// 获取 cpu info 信息
	cpuInfo, err := cpu.Info()

	//fmt.Printf("cpu info: %v\n", cpuInfo)
	if err != nil {
		log.Fatal("Error getting cpu.Info:", err)
	} else {
		sysInfo.CPUBrand = cpuInfo[0].ModelName
		sysInfo.CPUVenderID = cpuInfo[0].VendorID
	}
	// 获取 cpu 核心数
	cpuCounts, err := cpu.Counts(true)
	if err != nil {
		log.Fatal("Error getting cpu.Counts:", err)
	} else {
		sysInfo.CPUNum = cpuCounts
	}
	// 获取 uptime
	uptime, _ := host.Uptime()
	sysInfo.Uptime = uptime
	// 获取 avg
	avg, err := load.Avg()
	if err != nil {
		log.Fatal("Error getting load.Avg:", err)
	} else {
		sysInfo.Load1 = avg.Load1
		sysInfo.Load5 = avg.Load5
		sysInfo.Load15 = avg.Load15
	}
	VirMemory, err := mem.VirtualMemory()
	if err != nil {
		log.Fatal("Error getting Memory:", err)
	} else {
		sysInfo.MemoryTotal = VirMemory.Total / 1024
		sysInfo.MemoryUsed = VirMemory.Used / 1024
		sysInfo.SwapTotal = VirMemory.SwapTotal
		sysInfo.SwapUsed = sysInfo.SwapTotal - VirMemory.SwapFree
	}
	return sysInfo
	//return SystemInfo{
	//	OSName:        hostInfo.OS,
	//	OSArch:        hostInfo.KernelArch,
	//	OSFamily:      hostInfo.PlatformFamily,
	//	OSRelease:     hostInfo.PlatformVersion,
	//	KernelVersion: hostInfo.KernelVersion,
	//	CPUNum:        cpuCounts,
	//	CPUBrand:      cpuInfo[0].ModelName,
	//	CPUVenderID:   cpuInfo[0].VendorID,
	//	HostName:      hostInfo.Hostname,
	//	Uptime:        uptime,
	//	Load1:         avg.Load1,
	//	Load5:         avg.Load5,
	//	Load15:        avg.Load15,
	//	MemoryTotal:   VirMemory.Total / 1024,
	//	MemoryUsed:    VirMemory.Used / 1024,
	//	SwapTotal:     VirMemory.SwapTotal,
	//	SwapUsed:      (VirMemory.SwapTotal - VirMemory.SwapFree) / 1024,
	//}
}

func CalculateCPUUsage(INTERVAL float64) float64 {
	//before, _ := cpu.Times(false)
	//time.Sleep(time.Duration(INTERVAL) * time.Second)
	//after, _ := cpu.Times(false)
	//// Idle时间差
	//idleDelta := after[0].Idle - before[0].Idle
	//// 总时间差
	//totalDelta := after[0].Total() - before[0].Total()
	//
	//// 计算CPU使用率百分比
	//cpuUsage := (totalDelta - idleDelta) / totalDelta * 100
	//fmt.Printf("cpu usage: %v\n", cpuUsage)
	percent, err := cpu.Percent(time.Duration(INTERVAL)*time.Second, false)
	if err != nil {
		log.Fatalf("Error to run cpu.Percent, %v", err)
		return 0
	}
	fmt.Printf("cpu usage: %v\n", percent[0])
	return math.Round(percent[0]*100) / 100
}
