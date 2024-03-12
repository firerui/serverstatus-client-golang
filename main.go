package main

import (
	"fmt"
	//"pkg/sysinfo.go"
	//sysinfo "serverstatus-client/pkg"
)

func main() {
	a := getSysInfo()
	fmt.Printf("systeminfo: %v", a)
}
