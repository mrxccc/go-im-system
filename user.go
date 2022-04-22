package main

import (
	"net"
	"strings"
)

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server
}

// 创建一个用户的API
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}

	go user.ListenMessage()

	return user
}

// 监听当前User channel的方法，一旦有消息，就直接发送给对端客户端
func (this *User) ListenMessage() {
	for {
		msg := <-this.C

		this.conn.Write([]byte(msg + "\n"))
	}
}

func (this *User) OnLine() {
	// 用户上线，将用户加入到online中
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()

	// 广播当前用户上线消息
	this.server.BroadCast(this, "已上线")
}

func (this *User) OffLine() {
	// 用户上线，将用户加入到online中
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()
	// 广播当前用户下线消息
	this.server.BroadCast(this, "已下线")
}

func (this *User) DoMessage(msg string) {
	if msg == "who" {
		// 查询当前在线用户
		this.server.mapLock.Lock()
		for _, user := range this.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ":在线。。。\n"
			this.sendMsg(onlineMsg)
		}
		this.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		newName := msg[7:]
		_, ok := this.server.OnlineMap[newName]
		if ok {
			this.sendMsg("当前用户名已被使用\n")
		} else {
			this.server.mapLock.Lock()
			delete(this.server.OnlineMap, this.Name)
			this.server.OnlineMap[newName] = this
			this.server.mapLock.Unlock()

			this.Name = newName
			this.sendMsg("您已经更新用户名：" + this.Name + "\n")
		}
	} else if len(msg) > 4 && msg[:3] == "to|" {
		remoteMsg := strings.Split(msg, "|")
		if len(remoteMsg) < 3 {
			this.sendMsg("消息格式不正确，请使用\"to|张三|你好啊\"格式。\n")
		}
		remoteUser, ok := this.server.OnlineMap[remoteMsg[1]]
		if !ok {
			this.sendMsg("该用户名:" + remoteMsg[1] + "不存在\n")
		} else {
			if remoteMsg[2] == "" {
				this.sendMsg("无消息内容，请重发\n")
				return
			}
			remoteUser.sendMsg(this.Name + "对您说" + remoteMsg[2])
		}
	} else {
		this.server.BroadCast(this, msg)
	}
}

func (this *User) sendMsg(msg string) {
	this.conn.Write([]byte(msg))
}
