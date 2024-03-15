package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"serverstatus-client/client"
)

var (
	ADDRESS  = flag.String("a", "http://127.0.0.1:8080/report", "HTTP/TCP address")
	USER     = flag.String("u", "h1", "auth user")
	PASS     = flag.String("p", "p1", "auth pass")
	INTERVAL = flag.Float64("interval", 2.0, "report interval")
	EXTRA    = flag.Bool("disable-extra", false, "disable extra info report")
	PING     = flag.Bool("disable-ping", false, "disable ping")
	TUPD     = flag.Bool("disable-tupd", false, "disable t/u/p/d")
	CM       = flag.String("cm", "cm.tz.cloudcpp.com:80", "China Mobile probe addr")
	CT       = flag.String("ct", "ct.tz.cloudcpp.com:80", "China Telecom probe addr")
	CU       = flag.String("cu", "ct.tz.cloudcpp.com:80", "China Unicom probe addr")
	W        = flag.Int64("w", 0, "weight for rank")
	NOTIFY   = flag.Bool("disable-notify", false, "disable notify")
	T        = flag.String("t", "", "host type")
	LOCATION = flag.String("location", "", "location")
	I        = flag.String("iface", "", "iface list, eg: eth0,eth1")
	//interval = flag.Int64("interval",1,"report interval")
	// 其他命令行选项...
)

type StatData struct {
	Frame          string        `json:"frame"`
	Version        string        `json:"version"`
	Gid            string        `json:"gid"`
	Alias          string        `json:"alias"`
	ConnectName    string        `json:"name"`
	Weight         uint8         `json:"weight"`
	VnStat         bool          `json:"vnstat"`
	Notify         bool          `json:"notify"`
	OnLine4        bool          `json:"onLine4"`
	OnLine6        bool          `json:"onLine6"`
	Uptime         uint64        `json:"uptime"`
	Load1          float64       `json:"load_1"`
	Load5          float64       `json:"load_5"`
	Load15         float64       `json:"load_15"`
	MemoryTotal    uint64        `json:"memory_total"`
	MemoryUsed     uint64        `json:"memory_used"`
	SwapTotal      uint64        `json:"swap_total"`
	SwapUsed       uint64        `json:"swap_used"`
	TotalHDD       uint64        `json:"hdd_total"`
	UsedHDD        uint64        `json:"hdd_used"`
	CPU            uint8         `json:"cpu"`
	NetworkRx      uint64        `json:"network_rx"`
	NetworkTx      uint64        `json:"network_tx"`
	PingCT         uint32        `json:"ping_189"`
	PingCU         uint32        `json:"ping_10010"`
	PingCM         uint32        `json:"ping_10086"`
	TimeCT         uint32        `json:"time_189"`
	TimeCU         uint32        `json:"time_10010"`
	TimeCM         uint32        `json:"time_10086"`
	NetworkIN      uint64        `json:"network_in"`
	NetworkOUT     uint64        `json:"network_out"`
	LastNetworkIn  uint64        `json:"last_network_in"`
	LastNetworkOut uint64        `json:"last_network_out"`
	TCPNum         uint32        `json:"tcp"`
	UDPNum         uint32        `json:"udp"`
	Process        uint64        `json:"process"`
	Thread         uint64        `json:"thread"`
	IPInfo         client.IPInfo `json:"ip_info"`
	//SysInfo        client.SystemInfo `json:"sys_info"`
	SysInfo struct {
		Name          string `json:"name"`
		Version       string `json:"version"`
		OSName        string `json:"os_name"`
		OSArch        string `json:"os_arch"`
		OSFamily      string `json:"os_family"`
		OSRelease     string `json:"os_release"`
		KernelVersion string `json:"kernel_version"`
		CPUNum        int    `json:"cpu_num"`
		CPUBrand      string `json:"cpu_brand"`
		CPUVenderID   string `json:"cpu_vender_id"`
		HostName      string `json:"host_name"`
	} `json:"sys_info"`
	//Memory  MemoryUsage
	//HDD     HDDUsage
	//Network NetworkStatus
}

func main() {
	flag.Parse()
	if *ADDRESS == "" || *USER == "" || *PASS == "" {
		log.Fatal("ADDRESS, USER, PASS must need.")
	}
	//item := getClientData()
	item := StatData{}
	getClientData(&item)
	item.ConnectName = *USER
	//if *INTERVAL == 0{
	//	INTERVAL = time.Second
	//}
	item.VnStat = true
	data, _ := json.Marshal(item)
	strData := string(data)
	fmt.Printf("data: %v\n", strData)

	// todo 先做一个让它打印出需要收集信息的东西
	//a := getSysInfo()
	//a := client.GetSysInfo()
	//b := client.GetHDDUsage()
	//fmt.Printf("systeminfo: %v\n", a)
	//fmt.Printf("diskinfo: %v\n", b)
	////client.GetIOCounters()
	////client.GetTraffic(2)
	//client.GetVnstatTrafiic(2)
	//client.GetConns()
	//c := client.GetIPInfo()
	//fmt.Printf("ipinfo: %v\n", c)
}

func getClientData(item *StatData) {
	//var item StatData
	item.Frame = "data"
	item.Version = "beta-v0.1"
	item.OnLine4 = client.CheckNetwork(4, 2.0)
	item.OnLine6 = client.CheckNetwork(6, 2.0)
	if item.VnStat {
		item.NetworkIN, item.NetworkOUT, item.LastNetworkIn, item.LastNetworkOut = client.GetVnStatTraffic()
		_, _, item.NetworkRx, item.NetworkTx = client.GetSysTraffic(*INTERVAL)
	} else {
		item.NetworkIN, item.NetworkOUT, item.NetworkRx, item.NetworkTx = client.GetSysTraffic(*INTERVAL)
	}
	item.IPInfo = client.GetIPInfo()
	sInfo := client.GetSysInfo()
	item.SysInfo.Name = *USER
	item.SysInfo.Version = "client.go"
	item.SysInfo.OSName = sInfo.OSName
	item.SysInfo.OSArch = sInfo.OSArch
	item.SysInfo.OSFamily = sInfo.OSFamily
	item.SysInfo.OSRelease = sInfo.OSRelease
	item.SysInfo.KernelVersion = sInfo.KernelVersion
	item.SysInfo.CPUNum = sInfo.CPUNum
	item.SysInfo.CPUBrand = sInfo.CPUBrand
	item.SysInfo.CPUVenderID = sInfo.CPUVenderID
	item.SysInfo.HostName = sInfo.HostName
	item.Load1 = sInfo.Load1
	item.Load5 = sInfo.Load5
	item.Load15 = sInfo.Load15
	item.TCPNum, item.UDPNum = client.GetConns()
	item.TotalHDD, item.UsedHDD = client.GetHDDUsage()
	//return item
}
