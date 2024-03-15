package client

import (
	"encoding/json"
	"fmt"
	net2 "github.com/shirou/gopsutil/v3/net"
	"log"
	"net"
	"os/exec"
	"strings"
	"time"
)

//  理解 vnstat json 转换成结构体的时候应该是怎样的一个对应关系
//
//	type VnstatV struct {
//		// 用来判断 vnstat 的版本，以此读取到不同的json结构中去
//		Jsonversion string `json:"jsonversion"`
//	}

type vnStatJsonCommon struct {
	VnstatVersion string `json:"vnstatversion"`
	Jsonversion   string `json:"jsonversion"`
	Interfaces    []struct {
		ID      string `json:"id,omitempty"`
		Name    string `json:"name"`
		Traffic struct {
			Total struct {
				Tx uint64 `json:"tx"`
				Rx uint64 `json:"rx"`
			} `json:"total"`
			Months []struct {
				Date struct {
					Year  int `json:"year"`
					Month int `json:"month"`
				} `json:"date"`
				Tx uint64 `json:"tx"`
				Rx uint64 `json:"rx"`
			} `json:"month,omitempty"`
		} `json:"traffic"`
	} `json:"interfaces"`
}

// no old vnStat to test
//type vnStatJsonV1 struct {
//	VnstatVersion string `json:"vnstatversion"`
//	Jsonversion   string `json:"jsonversion"`
//	Interfaces    []struct {
//		ID      string `json:"id,omitempty"`
//		Name    string `json:"name"`
//		Traffic struct {
//			Total struct {
//				Tx uint64 `json:"tx"`
//				Rx uint64 `json:"rx"`
//			} `json:"total"`
//			Months []struct {
//				Date struct {
//					Year  int `json:"year"`
//					Month int `json:"month"`
//				} `json:"date"`
//				Tx uint64 `json:"tx"`
//				Rx uint64 `json:"rx"`
//			} `json:"months,omitempty"`
//		} `json:"traffic"`
//	} `json:"interfaces"`
//}

var invalidInterface = []string{"lo", "tun", "kube", "docker", "vmbr", "br-", "vnet", "veth"}
var perNetIn uint64
var perNetOut uint64
var bandwidthFactor uint8 //这个元素用于区分vnstat版本，在不同版本中返回的byte数据不一样

// 检测 IPv4 / IPv6 是否能够访问互联网
func CheckNetwork(IPVersion int, INTERVAL float64) bool {
	var domain string
	switch IPVersion {
	case 4:
		domain = "1.1.1.1:53"
	case 6:
		domain = "[2001:4860:4860::8888]:53"
	default:
		return false
	}
	conn, err := net.DialTimeout("tcp", domain, time.Duration(INTERVAL)*time.Second)
	if err != nil {
		log.Fatal("Error to check v4/v6.")
		return false
	}
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Fatal("Error to Close check v4/v6 conn.")
		}
	}(conn)
	return true
}

// GetConns read tcp & udp connections
func GetConns() (tcpActiveCount uint32, udpActiveCount uint32) {
	//var tcpActiveCount, udpActiveCount uint64
	tcpConns, tcpErr := net2.Connections("tcp")
	if tcpErr != nil {
		log.Fatal("Error get tcp connections.")
	} else {
		// tcp active count
		for _, conn := range tcpConns {
			if conn.Status == "ESTABLISHED" || conn.Status == "LISTEN" { // 这里可以根据实际情况调整状态过滤条件
				tcpActiveCount++
			}
		}
	}

	udpConns, udpErr := net2.Connections("udp")
	if udpErr != nil {
		log.Fatal("Error get tcp connections.")
	} else {
		//for _, conn := range udpConns {
		//	// 对于UDP，由于没有像TCP那样的连接状态，通常活跃的UDP连接可以视为所有已知的UDP socket
		//	fmt.Printf("udp status: %v\n", conn.Status)
		//	udpActiveCount++
		//}
		udpActiveCount = uint32(len(udpConns))
	}

	//fmt.Printf("tcp: %v\nudp: %v\n", tcpConns, udpConns)
	fmt.Printf("tcp: %v\nudp: %v\n", tcpActiveCount, udpActiveCount)

	return tcpActiveCount, udpActiveCount
}

// 读取系统网卡计算流量，计算网速 (重启流量信息丢失)
func GetSysTraffic(INTERVAL float64) (netIn uint64, netOut uint64, netRx uint64, netTx uint64) {
	interfaces, err := net2.IOCounters(true)
	if err != nil {
		log.Fatal("Error to getting IOCounters.")
	}
	// 统计有效的网卡流量
	for _, i := range interfaces {
		if checkInterface(i.Name) {
			netIn += i.BytesRecv
			netOut += i.BytesSent
		}
	}
	netRx = (netIn - perNetIn) / uint64(INTERVAL)
	netTx = (netOut - perNetOut) / uint64(INTERVAL)
	perNetIn = netIn
	perNetOut = netOut
	fmt.Printf("netin: %v, netout: %v, netrx: %v, nettx: %v", netIn, netOut, netRx, netTx)
	return netIn, netOut, netRx, netTx
}

func GetVnStatTraffic() (netIn uint64, netOut uint64, mnetIn uint64, mnetOut uint64) {
	now := time.Now()
	cmd := exec.Command("/usr/bin/vnstat", "--json")
	output, err := cmd.Output()
	if err != nil {
		log.Fatal("Error to run vnstat command.")
		return 0, 0, 0, 0
	}
	vnStatJsonData := []byte(output)
	// default vnStat json version 2
	var jsonData vnStatJsonCommon
	vErr := json.Unmarshal(vnStatJsonData, &jsonData)
	if vErr != nil {
		log.Fatal("Error to unmarshal vnStat v2 json data.")
		return 0, 0, 0, 0
	}
	// check vnStat json version, if json version is 1
	if jsonData.Jsonversion == "1" {
		return 0, 0, 0, 0
	}
	//vnData := jsonData
	year := now.Year()
	month := int(now.Month())
	// sum all traffic
	for _, b := range jsonData.Interfaces {
		if checkInterface(b.Name) {
			netIn += b.Traffic.Total.Rx
			netOut += b.Traffic.Total.Tx
			fmt.Printf("Name: %v\n", b.Name)
			// read this month traffic
			if b.Traffic.Months[0].Date.Year == year && b.Traffic.Months[0].Date.Month == month {
				mnetIn += b.Traffic.Months[0].Rx
				mnetOut += b.Traffic.Months[0].Tx
			}
		}
	}
	//log.Fatalf("%v-%v", year, month)
	log.Fatalf("netIn: %v,netOut: %v,mnetIn: %v,mnetOut: %v", netIn, netOut, mnetIn, mnetOut)
	return netIn, netOut, netIn - mnetIn, netOut - mnetOut
}

func getIOCounters() {
	a, _ := net2.IOCounters(false)
	bytesSent := a[0].BytesRecv
	fmt.Printf("io counters: %v\n", bytesSent)
}

// 排除掉一些无效网卡，比如docker, lo等
func checkInterface(name string) bool {
	for _, v := range invalidInterface {
		if strings.Contains(name, v) {
			return false
		}
	}
	return true
}
