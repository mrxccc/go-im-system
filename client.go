package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp  string
	ServePort int
	Name      string
	conn      net.Conn
	flag      int
}

func NewClient(serverIp string, serverPort int) *Client {
	client := &Client{
		ServerIp:  serverIp,
		ServePort: serverPort,
		flag:      9999,
	}
	// 连接服务器
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.Dial error:", err)
	}

	client.conn = conn

	return client
}

func (client *Client) dealResponse() {
	// 一旦clinet.conn有数据，就直接copy到stdout标准输出上
	io.Copy(os.Stdout, client.conn)
}

var serverIp string
var serverPort int

func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置ip")
	flag.IntVar(&serverPort, "port", 8888, "设置端口")
}

func (client *Client) selectUsers() {
	sendMsg := "who\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn write err:", err)
		return
	}
}

func (client *Client) menu() bool {
	var flag int
	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("4.退出")

	fmt.Scanln(&flag)
	if flag >= 0 && flag <= 3 {
		client.flag = flag
		return true
	} else {
		fmt.Println(">>>>请输入合法返回的数字>>>>")
		return false
	}
}

// 修改用户名
func (client *Client) UpdateName() bool {
	fmt.Println("请输入用户名")
	fmt.Scanln(&client.Name)
	sendMsg := "rename|" + client.Name + "\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.write err:", err)
		return false
	}

	return true
}

func (client *Client) ShowOnlineUsers() {
	sendMsg := "who\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Printf("ShowOnlineUsers conn.write error: %v", err)
		return
	}
}

func (client *Client) PublicChat() {
	var chatMsg string
	// 提示用户输入信息
	fmt.Printf(">>>>请输入聊天内容, 输入exit退出\n")
	fmt.Scanln(&chatMsg)

	for chatMsg != "exit" {
		// 消息不为空时发送
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Printf("conn.write error: %v", err)
				break
			}
		}
		chatMsg = ""
		fmt.Printf(">>>>请输入聊天内容, 输入exit退出\n")
		fmt.Scanln(&chatMsg)
	}
}

func (client *Client) PrivateChat() {
	// 选择私聊的用户名
	var remoteName string
	var chatMsg string
	// 展示在线用户
	client.ShowOnlineUsers()
	fmt.Printf(">>>>请输入用户名, 输入exit退出\n")
	fmt.Scanln(&remoteName)

	for remoteName != "exit" {
		fmt.Printf(">>>>请输入消息内容, 输入exit退出\n")
		fmt.Scanln(&chatMsg)

		for chatMsg != "exit" {
			// 消息不为空发送
			if len(remoteName) != 0 {
				sendMsg := "to|" + remoteName + "|" + chatMsg + "\n\n"
				_, err := client.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("PriviteChat.conn.write.error", err)
					break
				}
			}
			chatMsg = ""
			fmt.Printf(">>>>请输入消息内容, 输入exit退出\n")
			fmt.Scanln(&chatMsg)
		}
		client.ShowOnlineUsers()
		fmt.Printf(">>>>请输入用户名, 输入exit退出\n")
		fmt.Scanln(&remoteName)
	}

}

func (client *Client) Run() {
	for client.flag != 0 {
		for client.menu() != true {

		}
		switch client.flag {
		case 1:
			// 群聊模式
			client.PublicChat()
			break
		case 2:
			// 私聊模式
			client.PrivateChat()
			break
		case 3:
			client.UpdateName()
			break
		}
	}
}

func main() {
	// 命令行解析
	flag.Parse()

	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println(">>>>服务器连接失败>>>>")
		return
	}

	fmt.Println(">>>>服务器连接成功>>>>")

	// 单独开启一个goroutine处理server返回的消息
	go client.dealResponse()

	// 启动客户端业务
	client.Run()

}
