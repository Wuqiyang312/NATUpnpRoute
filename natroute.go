package main

import (
	"bytes"
	"fmt"
	"gitlab.com/NebulousLabs/go-upnp"
	"log"
	"os"
	"time"
)

func main() {
	upnpPort := 44300

	// connect to router
	d, err := upnp.Discover()
	if err != nil {
		log.Fatal(err)
	}

	// 发现外部 IP
	ip, err := d.ExternalIP()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("您的IP是:", ip)

	// forward a port
	err = d.Forward(uint16(upnpPort), "upnp Ip")
	if err != nil {
		log.Fatal(err)
	}

	// 使用外部服务器的IP
	initHttp()
	ip, port, err := httpUpdate("POST", "https://me.wqyblog.cn/", bytes.NewBufferString(""), upnpPort)
	if err != nil {
		fmt.Println(err)
		err = Exit(d, upnpPort)
		if err != nil {
			return
		}
	}
	//fmt.Println("您的IP是:", ip)
	//fmt.Println("您的PORT是:", port)
	fmt.Println("访问地址:", "http://"+ip+":"+port)

	go func() {
		err = initNatProxy(upnpPort, ip, port)
		if err != nil {
			fmt.Println(err)
			err = Exit(d, upnpPort)
			if err != nil {
				return
			}
		}
	}()
	time.Sleep(time.Second * 5)
	go func() {
		err = httpWebSocket(ip, port)
		if err != nil {
			fmt.Println(err)
			err = Exit(d, upnpPort)
			if err != nil {
				return
			}
		}
	}()
	go Quit(d, upnpPort)
	select {}
}

func Quit(d *upnp.IGD, upnpPort int) {
	time.Sleep(time.Second * 10)
	// 输入任何按钮退出
	fmt.Println("输入任何按钮退出^_^...")
	b := make([]byte, 1)
	_, err := os.Stdin.Read(b)
	if err != nil {
		select {}
	}
	Exit(d, upnpPort)
}
func Exit(d *upnp.IGD, upnpPort int) (err error) {
	println("关闭UPNP端口")
	// un-forward a port
	err = d.Clear(uint16(upnpPort))
	if err != nil {
		return
	}
	// 记录路由器的位置
	loc := d.Location()
	// 直接连接到路由器
	d, err = upnp.Load(loc)
	if err != nil {
		return
	}
	os.Exit(0)
	return
}
