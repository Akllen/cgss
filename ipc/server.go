package ipc

import (
	"encoding/json"
	"fmt"
)

type Request struct {
	Method string `json:"method"`
	Params string `json:"params"`
}

type Response struct {
	Code string `json:"code"`
	Body string `json:"body"`
}

type Server interface {
	Name() string
	Handle(method, params string) *Response
}

type IpcServer struct {
	Server
}

func NewIpcServer(server Server) *IpcServer {
	return &IpcServer{server}
}

func (server *IpcServer) Connect() chan string {
	session := make(chan string, 0)

	go func(c chan string) {
		for {
			request := <-c

			if request == "CLOSE" { //通道传递关闭该连接指令
				break
			}

			var req Request
			err := json.Unmarshal([]byte(request), &req) //Unmarshal用于解json:把字符串形式的json放到一个模板数据结构中去
			if err != nil {
				fmt.Println("无效的请求格式", request)
				return
			}
			resp := server.Handle(req.Method, req.Params)

			b, err := json.Marshal(resp) //封装成json格式传递

			c <- string(b) //返回结果
		}
		fmt.Println("Session closed.")
	}(session) //把通道c接受的数据赋给session

	fmt.Println("一个新的session成功创建.")
	return session
}
