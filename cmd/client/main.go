package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"serverstatus-client/client"
	"time"
)

var (
	ADDRESS  = flag.String("a", "http://127.0.0.1:8080", "HTTP/TCP address")
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
	CPU            float64       `json:"cpu"`
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
	*ADDRESS = "http://10.0.0.62:8080/report"
	*USER = "h2"
	*PASS = "p2"
	headers := map[string]string{
		"ssr-auth": "single",
	}
	for {
		connect(*ADDRESS, *USER, *PASS, headers, data)
		//connect(data)
		time.Sleep(2 * time.Second)
	}
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

func connect(address string, username string, password string, headers map[string]string, data []byte) {
	log.Println("尝试连接...")

	// 构建HTTP请求，并设置基础认证信息
	req, err := http.NewRequest("POST", address, bytes.NewBuffer(data))
	if err != nil {
		log.Println("创建请求失败：", err.Error())
		return
	}

	// 设置基础认证
	req.SetBasicAuth(username, password)

	// 设置请求头
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// 设置超时
	netClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	// 发送请求
	resp, err := netClient.Do(req)
	if err != nil {
		log.Println("发送请求失败：", err.Error())
		return
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		log.Println("收到非成功响应状态码：", resp.StatusCode)
		return
	}

	// 处理响应体（如果需要的话）
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("读取响应体失败：", err.Error())
		return
	}
	log.Printf("发送成功，响应长度：%d Byte\n", len(bodyBytes))
}

//func connect(data []byte) {
//	log.Println("尝试连接...")
//	conn, err := net.DialTimeout("tcp", *ADDRESS, 30*time.Second)
//	if err != nil {
//		log.Println("Caught Exception:", err.Error())
//		time.Sleep(5 * time.Second)
//		return
//	}
//	defer func(conn net.Conn) {
//		_ = conn.Close()
//		time.Sleep(5 * time.Second)
//	}(conn)
//	_, _ = conn.Write([]byte((*USER + ":" + *PASS + "\n")))
//	_ = conn.SetWriteDeadline(time.Now().Add(12 * time.Second))
//	write, _ := conn.Write(data)
//	log.Printf("发送成功：%dByte\n", write)
//}

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
	item.Uptime = sInfo.Uptime
	item.TCPNum, item.UDPNum = client.GetConns()
	item.TotalHDD, item.UsedHDD = client.GetHDDUsage()
	item.MemoryTotal = sInfo.MemoryTotal
	item.MemoryUsed = sInfo.MemoryUsed
	item.SwapTotal = sInfo.SwapTotal
	item.SwapUsed = sInfo.SwapUsed
	item.CPU = client.CalculateCPUUsage(1.2)
	//return item
}
