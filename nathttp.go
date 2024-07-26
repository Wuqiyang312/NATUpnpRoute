package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/net/websocket"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"
)

// 定义错误常量
const (
	errCreateRequest               = "创建请求时出错:"
	errSendRequest                 = "发送请求时出错:"
	errCloseResponseBody           = "关闭响应体时出错:"
	errReadResponse                = "读取响应时出错:"
	errParseResponse               = "解析响应时出错:"
	errResolvesTheLocalAddress     = "解析本地地址时出错:"
	errOccurredWhileGettingAddress = "获取接口地址时出错"
	errOther                       = "其他错误"
)

var nativeIP = "192.168.1.15"

func initHttp() {
	nativeCIDR := "192.168.1.0/24"

	_, ipNet, err := net.ParseCIDR(nativeCIDR)
	if err != nil {
		err = fmt.Errorf("%s %w", errOther, err)
		return
	}

	interfaces, err := net.Interfaces()
	if err != nil {
		err = fmt.Errorf("%s %w", errOther, err)
		return
	}

	for _, iface := range interfaces {
		// 获取每个接口的地址
		addrs, err := iface.Addrs()
		if err != nil {
			err = fmt.Errorf("%s %w", errOccurredWhileGettingAddress, err)
			continue
		}

		// 打印每个地址
		for _, addr := range addrs {
			// 检查地址类型并转换为 IP 地址
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			// 排除回环地址
			if ip != nil && !ip.IsLoopback() {
				// 打印 IPv4 地址
				if ip.To4() != nil {
					//fmt.Println("IPv4:", ip.String())

					// 如果 IP 不为空且不是回环地址，检查是否属于指定网段
					if ip != nil && !ip.IsLoopback() && ipNet.Contains(ip) {
						fmt.Println("选择IPv4:", ip.String())
						nativeIP = ip.String()
						break
					}
				}
			}
		}
	}
}

func httpUpdate(method string, url string, body *bytes.Buffer, sendPort int) (ip string, port string, err error) {
	// 使用context控制请求超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		err = fmt.Errorf("%s %w", errCreateRequest, err)
		return
	}
	request.Header.Set("Content-Type", "application/json")

	// 自定义的 Dialer，指定本地 IP 和端口
	localAddr, err := net.ResolveIPAddr("ip", nativeIP) // 替换为你的本地 IP 地址
	if err != nil {
		err = fmt.Errorf("%s %w", errResolvesTheLocalAddress, err)
		return
	}

	dialer := &net.Dialer{
		LocalAddr: &net.TCPAddr{
			IP:   localAddr.IP,
			Port: sendPort, // 替换为你需要的本地端口
		},
		Timeout: 30 * time.Second,
	}

	// 自定义的 Transport
	transport := &http.Transport{
		DialContext: dialer.DialContext,
	}

	client := &http.Client{
		Transport: transport,
	}
	resp, err := client.Do(request)
	if err != nil {
		err = fmt.Errorf("%s %w: 请求URL：%s", errSendRequest, err, url)
		return
	}
	defer func() {
		client.CloseIdleConnections()
		if cerr := resp.Body.Close(); cerr != nil {
			// 如果关闭时出错，且之前没有错误，则将关闭错误返回
			if err == nil {
				err = fmt.Errorf("%s %w", errCloseResponseBody, cerr)
			} else {
				// 优化：如果已有错误，添加关闭响应体时的错误信息作为上下文
				err = fmt.Errorf("%s, %s %w", err, errCloseResponseBody, cerr)
			}
		}
		fmt.Println("关闭http连接")
	}()

	if resp.StatusCode != http.StatusOK {
		// 优化：添加响应状态码作为错误信息的一部分
		err = fmt.Errorf("%s 请求失败, 状态码：%d", errSendRequest, resp.StatusCode)
		return
	}

	decoder := json.NewDecoder(resp.Body)
	var result map[string]interface{}
	if err = decoder.Decode(&result); err != nil {
		err = fmt.Errorf("%s %w", errReadResponse, err)
		return
	}
	// fmt.Println(result)

	ip, err = parseJSONResult(result, "ip", errParseResponse)
	if err != nil {
		return
	}

	port, err = parseJSONResult(result, "port", errParseResponse)
	if err != nil {
		return
	}
	return
}

func httpWebSocket(ip string, post string) (err error) {
	fmt.Println("端口保护启动")
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// 设置WebSocket服务器的URL
	origin := "http://localhost/"
	url := "ws://" + ip + ":" + post + "/ping"

	// 建立连接
	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		return
	}
	defer func(ws *websocket.Conn) {
		err := ws.Close()
		if err != nil {
			return
		}
	}(ws)

	done := make(chan struct{})

	// 接收消息的goroutine
	go func() {
		defer close(done)
		var msg = make([]byte, 512)
		for {
			_, err := ws.Read(msg)
			if err != nil {
				return
			}
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-interrupt:
			// 关闭WebSocket连接
			err = ws.Close()
			if err != nil {
				return
			}
			return
		}
	}
}

// 增加一个辅助函数来解析JSON结果并处理错误
func parseJSONResult(result map[string]interface{}, key string, errMsg string) (string, error) {
	if value, ok := result[key]; ok {
		if strVal, ok := value.(string); ok {
			return strVal, nil
		} else {
			return "", fmt.Errorf("%s '%s' 不是字符串类型", errMsg, key)
		}
	} else {
		return "", fmt.Errorf("%s '%s' 字段缺失", errMsg, key)
	}
}
