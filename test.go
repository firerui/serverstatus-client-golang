package main

import "C"
import (
	"encoding/json"
	"fmt"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	gopsnet "github.com/shirou/gopsutil/v3/net"
	"log"
	"math"
	"net"
	"os/exec"
	"strings"
	"time"
	// "github.com/shirou/gopsutil/mem"  // to use v2
)

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
	load15        float64
	//OnLine4       bool
	//OnLine6       bool
}

type IPInfo struct {
	IPQuery      string
	IPSource     string
	IPContinent  string
	IPCountry    string
	IPRegionName string
	IPCity       string
	IPIsp        string
	IPOrg        string
	IPAs         string
	IPASName     string
	IPLat        string
	IPLon        string
	IPTimeZone   string
}

type MemoryUsage struct {
	Total uint64
	Used  uint64
	Swap  SwapUsage
}

// SwapUsage 交换分区使用情况
type SwapUsage struct {
	Total uint64
	Used  uint64
}

type HDDUsage struct {
	Size uint64
	Used uint64
}

type NetworkStatus struct {
	OnIPv4     bool
	OnIPv6     bool
	PingCT     uint32
	PingCU     uint32
	PingCM     uint32
	TimeCT     uint32
	TimeCU     uint32
	TimeCM     uint32
	NetworkRX  uint32
	NetworkTX  uint32
	NetworkIN  uint32
	NetworkOUT uint32
	TCPNum     uint32
	UDPNum     uint32
}

// StatData 结构体存储统计数据
type StatData struct {
	Frame       string  `json:"frame"`
	Version     string  `json:"version"`
	Gid         string  `json:"gid"`
	Alias       string  `json:"alias"`
	ConnectName string  `json:"name"`
	Weight      uint8   `json:"weight"`
	Vnstat      bool    `json:"vnstat"`
	Notify      bool    `json:"notify"`
	OnLine4     bool    `json:"onLine4"`
	OnLine6     bool    `json:"onLine6"`
	Uptime      uint64  `json:"uptime"`
	Load1       float64 `json:"load_1"`
	Load5       float64 `json:"load_5"`
	Load15      float64 `json:"load_15"`
	MemoryTotal uint64  `json:"memory_total"`
	MemoryUsed  uint64  `json:"memory_used"`
	SwapTotal   uint64  `json:"swap_total"`
	SwapUsed    uint64  `json:"swap_used"`
	TotalHDD    uint64  `json:"hdd_total"`
	UsedHDD     uint64  `json:"hdd_used"`
	CPUNum      uint8   `json:"cpu"`
	NetworkRx   uint64  `json:"network_rx"`
	NetworkTx   uint64  `json:"network_tx"`
	PingCT      uint32  `json:"ping_189"`
	PingCU      uint32  `json:"ping_10010"`
	PingCM      uint32  `json:"ping_10086"`
	TimeCT      uint32  `json:"time_189"`
	TimeCU      uint32  `json:"time_10010"`
	TimeCM      uint32  `json:"time_10086"`
	NetworkIN   uint32  `json:"network_in"`
	NetworkOUT  uint32  `json:"network_out"`
	TCPNum      uint32  `json:"tcp"`
	UDPNum      uint32  `json:"udp"`
	Process     uint64  `json:"process"`
	Thread      uint64  `json:"thread"`
	IPInfo      IPInfo  `json:"ip_info"`
	SysInfo     SystemInfo
	//Memory  MemoryUsage
	//HDD     HDDUsage
	//Network NetworkStatus
}

func main() {
	//// get osFamily
	////osF, _ := host.Info()
	//fmt.Printf("HostName: %v\n", h.Hostname)
	//// 获取 kernelVersion
	//kv, _ := host.KernelVersion()
	//fmt.Printf(kv)

	hostInfo, _ := host.Info()
	// 从这一行开始构建 py 版本 sample 函数中所需的数据
	// 1. 获取 指定时间内 cpu 使用间隔
	usageDiff := calculateCPUUsageChange(1)
	fmt.Printf("CPU usage change over 1 second : % .2f%%\n", usageDiff)

	// 2. 获取uptime
	myUptime := hostInfo.Uptime
	fmt.Printf("Uptime: %v\n", myUptime)

	// 3. 获取 avg, 总共三个指标
	myAVG, _ := load.Avg()
	fmt.Printf("load1: %v, load5: %v, load15: %v\n", myAVG.Load1, myAVG.Load5, myAVG.Load15)

	// 4. 获取内存总数，以及 swap 的总数，以及使用数字
	myMem, _ := mem.VirtualMemory()
	totalMB := float64(myMem.Total) / math.Pow(1024, 2)
	usedMB := float64(myMem.Used) / math.Pow(1024, 2)
	totalSwapMB := float64(myMem.SwapTotal) / math.Pow(1024, 2)
	//freeSwapMB := float64(freeSwapBytes) / math.Pow(1024, 2)
	usedSwapMB := float64(myMem.SwapTotal-myMem.SwapFree) / math.Pow(1024, 2)
	fmt.Printf("T: %v, U: %v, TSwap: %v, USwap: %v\n", totalMB, usedMB, totalSwapMB, usedSwapMB)

	// 5. 获取所有硬盘的大小和使用数据
	//diskList, _ := disk.Partitions(true)
	//fmt.Printf("diskList? %v", diskList)
	hddu := getHDD()
	fmt.Printf("hdd: %v\n", hddu)

	a := getSysInfo()
	fmt.Printf("SystemInfo:%v\n", a)

	if getNetwork(4) {
		fmt.Printf("ipv4 is ok!\n")
	}

	if getNetwork(6) {
		fmt.Printf("ipv6 is ok!\n")
	}

	//  计算网卡的方式计算内容
	sysTrafficIn, sysTrafficOut := getSysTraffic()
	fmt.Printf("Traffic in: %v\nTraffic out: %v\n", sysTrafficIn, sysTrafficOut)
}

// 使用 vnstat 来获取流量
func getVnstatNetwork(bool) (netIn uint64, netOut uint64, mNetIn uint64, mNetOut uint64) {
	now := time.Now()
	cmd := exec.Command("/usr/bin/vnstat", "--json")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error running vnstat command:", err)
		return 0, 0, 0, 0
	}
	err = json.Unmarshal(output, &vnstatRes)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return 0, 0, 0, 0
	}
	return
}

// 使用 go psutil net 从各个网卡中读取数据合并计算
func getSysTraffic() (netIn uint64, netOut uint64) {
	//var netIn, netOut uint64
	interfaces, err := gopsnet.IOCounters(true)
	if err != nil {
		log.Fatal("Error getSysTraffic...")
	}
	for _, info := range interfaces {
		netIn += info.BytesRecv
		netOut += info.BytesSent
	}
	return netIn, netOut
}

// 确定 IPv4 / IPv6 是否能够访问互联网
func getNetwork(IPVersion int) bool {
	var domain string
	switch IPVersion {
	case 4:
		domain = "ipv4.ip.sb"
	case 6:
		domain = "ipv6.ip.sb"
	default:
		return false
	}
	conn, err := net.DialTimeout("tcp", domain+":80", 2*time.Second)
	if err != nil {
		return false
	}
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Fatal("Error to Close getNetwork Conn...")
		}
	}(conn)
	return true
}

// 获取系统信息
func getSysInfo() SystemInfo {
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
		load15:        avg.Load15,
	}
}

// 获取 硬盘 信息
// 该方法在 mac 上存在比较严重的问题，不能获取到容量，增加 apfs 格式后会获取到重复容量
func getHDD() HDDUsage {
	validFs := []string{"ext4", "ext3", "ext2", "reiserfs", "jfs", "btrfs",
		"fuseblk", "zfs", "simfs", "ntfs", "fat32", "exfat", "xfs"}

	partitions, err := disk.Partitions(true)
	//fmt.Printf("par:%v\n", partitions)
	if err != nil {
		return HDDUsage{
			Size: 0,
			Used: 0,
		}
	}
	totalSizeMB := disk.UsageStat{}.Total
	usedSizeMB := disk.UsageStat{}.Total
	for _, part := range partitions {
		if contains(validFs, strings.ToLower(part.Fstype)) {
			usage, err := disk.Usage(part.Mountpoint)
			if err != nil {
				continue
			}
			totalSizeMB += usage.Total / 1024 / 1024
			usedSizeMB += usage.Used / 1024 / 1024
		}
	}
	return HDDUsage{
		Size: totalSizeMB,
		Used: usedSizeMB,
	}
}

// 自定义一个检查切片是否包含某元素的函数
func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// 计算cpu指定时间内的使用率百分比
func calculateCPUUsageChange(intervalTime uint64) float64 {
	before, _ := cpu.Times(false)
	time.Sleep(time.Duration(intervalTime) * time.Second)
	after, _ := cpu.Times(false)
	// Idle时间差
	idleDelta := after[0].Idle - before[0].Idle
	// 总时间差
	totalDelta := after[0].Total() - before[0].Total()

	// 计算CPU使用率百分比
	cpuUsage := (totalDelta - idleDelta) / totalDelta * 100

	return cpuUsage
}

//networkInfo, _ := gopsnet.IOCounters(true)
//type NetTraffic struct {
//	Name       string
//	RxBytes    uint64
//	TxBytes    uint64
//	RxPackets  uint64
//	TxPackets  uint64
//	RxErrors   uint64
//	TxErrors   uint64
//	RxDropped  uint64
//	TxDropped  uint64
//	RxOverruns uint64
//	TxCarrier  uint64
//}

//netTrafficList := []NetTraffic{}
//for _, info := range networkInfo {
//nt := NetTraffic{
//Name:      info.Name,
//RxBytes:   info.BytesRecv,
//TxBytes:   info.BytesSent,
//RxPackets: info.PacketsRecv,
//TxPackets: info.PacketsSent,
//RxErrors:  info.Errin,
//TxErrors:  info.Errout,
//RxDropped: info.Dropin,
//TxDropped: info.Dropout,
////RxOverruns: info.OversizePkts,
////TxCarrier:  info.CarrierLoss,
//}
//netTrafficList = append(netTrafficList, nt)
//}
//fmt.Printf("netTrafficList: %v\n", netTrafficList)
