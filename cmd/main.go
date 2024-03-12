package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/shirou/gopsutil/v3/disk"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/partition"
)

// SystemInfo 结构体存储系统信息
type SystemInfo struct {
	HostName      string
	OSName        string
	OSArch        string
	OSFamily      string
	OSRelease     string
	KernelVersion string
	CPUBrand      string
	CPUNum        int64
	CPUVenderID   string
	Uptime        int64
}

// StatData 结构体存储统计数据
type StatData struct {
	SysInfo SystemInfo
	Memory  MemoryUsage
	HDD     HDDUsage
	Network NetworkStatus
	// 其他统计数据...
}

// MemoryUsage 内存使用情况
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

// HDDUsage 硬盘使用情况
type HDDUsage struct {
	Size uint64
	Used uint64
}

// NetworkStatus 网络状态
type NetworkStatus struct {
	// ...
}

var (
	address       = flag.String("a", "http://127.0.0.1:8080/report", "HTTP/TCP address")
	user          = flag.String("u", "", "auth user")
	pass          = flag.String("p", "", "auth pass")
	interval      = flag.Duration("interval", time.Second, "report interval")
	disableExtra  = flag.Bool("disable-extra", false, "disable extra info report")
	disablePing   = flag.Bool("disable-ping", false, "disable ping")
	disableTupd   = flag.Bool("disable-tupd", false, "disable t/u/p/d")
	cm            = flag.String("cm", "cm.tz.cloudcpp.com:80", "China Mobile probe addr")
	ct            = flag.String("ct", "ct.tz.cloudcpp.com:80", "China Telecom probe addr")
	cu            = flag.String("cu", "ct.tz.cloudcpp.com:80", "China Unicom probe addr")
	w             = flag.Int64("w", 0, "weight for rank")
	disableNotify = flag.Bool("disable-notify", false, "disable notify")
	t             = flag.String("t", "", "host type")
	location      = flag.String("location", "", "location")
	i             = flag.String("iface", "", "iface list, eg: eth0,eth1")
	//interval = flag.Int64("interval",1,"report interval")
	// 其他命令行选项...
)

func main() {
	flag.Parse()

	// 获取系统基本信息
	sysInfo := getSysInfo()

	// 创建定时器，定期收集和上报数据
	ticker := time.NewTicker(*interval)
	defer ticker.Stop()

	for range ticker.C {
		statData := collectStats(sysInfo)

		// 序列化为JSON
		jsonData, err := json.Marshal(statData)
		if err != nil {
			fmt.Println("Error marshalling JSON:", err)
			continue
		}

		// 构建HTTP请求
		request, err := http.NewRequest("POST", *address, bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Println("Error creating request:", err)
			continue
		}

		// 设置基础认证信息
		request.SetBasicAuth(*user, *pass)

		// 发送请求
		client := &http.Client{}
		resp, err := client.Do(request)
		if err != nil {
			fmt.Println("Error sending request:", err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := ioutil.ReadAll(resp.Body)
			fmt.Printf("Unexpected response status: %d; Body: %s\n", resp.StatusCode, bodyBytes)
		}
	}
}

func getSysInfo() SystemInfo {
	info, err := host.Info()
	if err != nil {
		panic(err)
	}

	cpuInfo, err := cpu.Info()
	if err != nil || len(cpuInfo) == 0 {
		panic(err)
	}
	cpuBrand := cpuInfo[0].ModelName // 假设只有一颗CPU

	uptime := time.Since(time.Unix(info.Uptime, 0)).Seconds()

	return SystemInfo{
		HostName:      info.Hostname,
		OSName:        info.OS,
		OSArch:        info.Platform,
		OSFamily:      "",
		OSRelease:     info.KernelVersion,
		KernelVersion: info.KernelVersion,
		CPUBrand:      cpuBrand,
		Uptime:        int64(uptime),
	}
}

func collectStats(sysInfo SystemInfo) StatData {
	memUsage := mem.VirtualMemory()
	swapUsage := mem.SwapMemory()
	hddUsage := getHDDUsage()
	networkStatus := getNetworkStatus()

	// 注意：在实际项目中，根据需要添加其他数据收集功能

	return StatData{
		SysInfo: sysInfo,
		Memory: MemoryUsage{
			Total: memUsage.Total,
			Used:  memUsage.Used,
			Swap: SwapUsage{
				Total: swapUsage.Total,
				Used:  swapUsage.Used,
			},
		},
		HDD:     hddUsage,
		Network: networkStatus,
	}
}

func getHDDUsage() HDDUsage {
	usageList, err := partition.Partitions(true)
	if err != nil {
		panic(err)
	}

	var total, used uint64
	for _, usage := range usageList {
		if isValidFileSystem(usage.Fstype) {
			diskUsage, err := disk.Usage(usage.Mountpoint)
			if err != nil {
				continue
			}
			total += diskUsage.Total
			used += diskUsage.Used
		}
	}

	return HDDUsage{Size: total / 1024 / 1024, Used: used / 1024 / 1024}
}

func isValidFileSystem(fsType string) bool {
	validFS := []string{"ext4", "ext3", "ext2", "reiserfs", "jfs", "btrfs", "fuseblk", "zfs", "simfs", "ntfs", "fat32", "exfat", "xfs"}
	return contains(validFS, fsType)
}

// ... 其他辅助函数实现 ...

func getNetworkStatus() NetworkStatus {
	// 这里实现网络连接检测和流量统计等功能
	// 可能需要用到net.IOCountersByDiskio()等方法
	return NetworkStatus{} // 示例返回空结构体，需填充实际数据
}

// ... 其他未列出的辅助函数如contains()等...

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
