package main

import (
	"golang.org/x/net/websocket"
	"io"
	"net/http"
	"strconv"
	"time"
)

func websocketHandler(ws *websocket.Conn) {
	var err error
	for {
		time.Sleep(time.Second * 10)
		// 循环发送 hi
		_, err = ws.Write([]byte("hi\n"))
		if err != nil {
			return
		}
	}
}
func initNatProxy(upnpPort int) (err error) {
	helloHandler := func(w http.ResponseWriter, req *http.Request) {
		_, err := io.WriteString(w, "Hello, world!\n")
		if err != nil {
			return
		}
	}

	http.Handle("/ping", websocket.Handler(websocketHandler))
	http.HandleFunc("/", helloHandler)
	err = http.ListenAndServe(":"+strconv.Itoa(upnpPort), nil)
	if err != nil {
		return
	}
	return
}
