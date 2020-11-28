package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"

	yaml "gopkg.in/yaml.v2"
)

// Config struct
type Config struct {
	URL string `json:"url"`
}

// ClashConf struct
type ClashConf struct {
	Proxies     []Proxy    `json:"proxies"`
	ProxyGroups ProxyGroup `json:"proxy-groups"`
	Rules       []string   `json:"rules"`
}

// Proxy struct
type Proxy struct {
	Name      string    `json:"name"`
	Server    string    `json:"server"`
	Port      int64     `json:"port"`
	Type      string    `json:"type"`
	UUID      string    `json:"uuid"`
	AlterID   int8      `json:"alterId"`
	Cipher    string    `json:"cipher"`
	TLS       bool      `json:"tls"`
	Network   string    `json:"network"`
	WSPath    string    `json:"ws-path"`
	WSHeaders WSHerader `json:"ws-headers"`
	UDP       bool      `json:"udp"`
}

// WSHerader struct
type WSHerader struct {
	Host string `json:"host"`
}

// ProxyGroup struct
type ProxyGroup struct {
	Name    string   `json:"name"`
	Type    string   `json:"type"`
	Proxies []string `json:"proxies"`
}

var config Config

func main() {
	confPath := flag.String("c", "./config.yaml", "set configuration file (default: ./config.yaml)")
	flag.Parse()

	// check config
	if err := checkConfig(*confPath); err != nil {
		log.Fatalf("配置文件读取失败, err: %v", err)
		return
	}

	log.Println("服务已启动")
	http.HandleFunc("/", FetchClashConf)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// FetchClashConf func
func FetchClashConf(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get(config.URL)
	if err != nil {
		log.Printf("请求订阅配置信息失败, err: %v\n", err)
		w.Write(getCache())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("请求订阅配置信息失败, HTTP Status: %d\n", resp.StatusCode)
		w.Write(getCache())
		return
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("读取订阅配置信息内容失败, err: %v\n", err)
		w.Write(getCache())
		return
	}

	clash := ClashConf{}
	if err := yaml.Unmarshal(data, &clash); err != nil {
		log.Printf("订阅配置信息反序列化失败, err: %v\n", err)
		w.Write(getCache())
		return
	}

	if err := setCache(data); err != nil {
		log.Printf("写入缓存失败, err: %v\n", err)
	}

	w.Write(data)
	return
}

func getCache() []byte {
	data, err := ioutil.ReadFile("./cache.yaml")
	if err != nil {
		return []byte{}
	}
	return data
}

func setCache(data []byte) error {
	return ioutil.WriteFile("./cache.yaml", data, 0666)
}

func checkConfig(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, &config)
}
