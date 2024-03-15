package client

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type IPInfo struct {
	Query      string  `json:"query"`
	Source     string  `json:"source"`
	Continent  string  `json:"continent"`
	Country    string  `json:"country"`
	RegionName string  `json:"region_name"`
	City       string  `json:"city"`
	Isp        string  `json:"isp"`
	Org        string  `json:"org"`
	As         string  `json:"as"`
	ASName     string  `json:"asname"`
	Lat        float64 `json:"lat"`
	Lon        float64 `json:"lon"`
	TimeZone   string  `json:"timezone"`
}

const IPAPIURL string = "http://ip-api.com/json?fields=status,message,continent,continentCode,country,countryCode,region,regionName,city,district,zip,lat,lon,timezone,isp,org,as,asname,query&lang=zh-CN"

func GetIPInfo() IPInfo {
	url := IPAPIURL
	http.DefaultClient.Timeout = 5 * time.Second
	// 发起GET请求
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal("Error making GET request:", err)
		return IPInfo{}
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		log.Fatal("Invalid response status code:", resp.StatusCode)
		return IPInfo{}
	}

	// 读取响应体内容
	//bodyBytes, err := ioutil.ReadAll(resp.Body)
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Error reading response body:", err)
		return IPInfo{}
	}

	// 将字节流转换为字符串
	bodyString := string(bodyBytes)
	bodyBytesNow := []byte(bodyBytes)
	fmt.Println("Response Body:", bodyString)

	var bodyJsonData IPInfo
	bErr := json.Unmarshal(bodyBytesNow, &bodyJsonData)
	if bErr != nil {
		log.Fatal("Error to unmarshal get ipinfo body.", bErr)
		return IPInfo{}
	}
	//fmt.Printf("ip city: %v\n", bodyJsonData.Query)
	return IPInfo{
		Query:      bodyJsonData.Query,
		Source:     bodyJsonData.Source,
		Continent:  bodyJsonData.Continent,
		Country:    bodyJsonData.Country,
		RegionName: bodyJsonData.RegionName,
		City:       bodyJsonData.City,
		Isp:        bodyJsonData.Isp,
		Org:        bodyJsonData.Org,
		As:         bodyJsonData.As,
		ASName:     bodyJsonData.ASName,
		Lat:        bodyJsonData.Lat,
		Lon:        bodyJsonData.Lon,
		TimeZone:   bodyJsonData.TimeZone,
	}
}
