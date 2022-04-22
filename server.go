package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int
	// 在线用户的列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	// 消息广播的channel
	Message chan string
}

func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}

	return server
}

// 监听message广播消息的channel的gorouting,一旦有消息发送 给全部的在线user
func (this *Server) ListenMessager() {
	for {
		msg := <-this.Message
		this.mapLock.Lock()
		// 发送给给全部的在线User
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}

		this.mapLock.Unlock()
	}
}

//广播消息的方法
func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	this.Message <- sendMsg
}

func (this *Server) Handle(conn net.Conn) {
	// fmt.Println("建立连接成功")
	user := NewUser(conn, this)
	// 用户上线，将用户加入到online中
	user.OnLine()
	isLive := make(chan bool)
	// 接受客户端发来的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.OffLine()
				return
			}

			if err != nil && err != io.EOF {
				fmt.Println("Conn read err", err)
				return
			}
			msg := string(buf[:n-1])

			user.DoMessage(msg)

			isLive <- true
		}
	}()

	// 当前handle阻塞
	for {
		select {
		case <-isLive:
			//当前用户是活跃的，应该重置定时器
			//不做任何事情，为了激活select,更新下面的定时器
		case <-time.After(time.Minute * 5):
			user.sendMsg("你被踢了")
			close(user.C)
			conn.Close()
			return
		}
	}
}

func (this *Server) Start() {
	fmt.Println("启动")
	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net Listen err:", err)
		return
	}

	// close listen socket
	defer listener.Close()

	// 启动监听Message的goroutine
	go this.ListenMessager()

	for {
		// accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Listen Accept err:", err)
			return
		}

		// do handle
		go this.Handle(conn)

	}

}
