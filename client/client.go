package main

import (
	"flag"
	"fmt"
	"net"
)

// Client ...
type Client struct {
	IP     string
	Port   int
	Name   string
	conn   net.Conn
	flag   int
	online chan bool
}

// NewClient 创建 client
func NewClient(IP string, port int) *Client {
	// 创建client
	client := &Client{
		IP:     IP,
		Port:   port,
		flag:   5,
		online: make(chan bool),
	}
	// 链接服务端
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", client.IP, client.Port))
	if err != nil {
		fmt.Println("net.Dial err", err)
		return nil
	}
	client.conn = conn
	return client
}

// DoResponse 返回的数据显示在控制台
func (c *Client) DoResponse() {
	for {
		buf := make([]byte, 4096)
		n, err := c.conn.Read(buf)
		if n == 0 || err != nil {
			fmt.Println("connection is closed")
			c.online <- false
			return
		}
		fmt.Print(string(buf))
	}

}

// UpdateName 更新用户名
func (c *Client) UpdateName() bool {
	fmt.Println("input username")
	fmt.Scanln(&c.Name)
	msg := "rename|" + c.Name + "\n"
	_, err := c.conn.Write([]byte(msg))
	if err != nil {
		fmt.Println("c.conn.Write err", err)
		return false
	}
	return true
}

// BroadCast 广播
func (c *Client) BroadCast() {
	fmt.Println("input msg (Type \"exit\" to exit)")
	var msg string
	fmt.Scanln(&msg)
	for msg != "exit" {
		fmt.Println("input msg:" + msg)
		_, err := c.conn.Write([]byte(msg + "\n"))
		if err != nil {
			fmt.Println("c.conn.Write err", err)
			break
		}
		fmt.Println("input msg (Type \"exit\" to exit)")
		fmt.Scanln(&msg)
	}

}

// GetList 获取列表
func (c *Client) GetList() {
	_, err := c.conn.Write([]byte("onlineList\n"))
	if err != nil {
		fmt.Println("c.conn.Write err", err)
		return
	}
}

// PrivateMessage 私聊
func (c *Client) PrivateMessage() {
	c.GetList()
	fmt.Println("input username (Type \"exit\" to exit)")
	var username string
	fmt.Scanln(&username)
	for username != "exit" {
		fmt.Println("input msg (Type \"exit\" to exit)")
		var msg string
		fmt.Scanln(&msg)
		for msg != "exit" {
			newMsg := "to|" + username + "|" + msg + "\n"
			_, err := c.conn.Write([]byte(newMsg))
			if err != nil {
				fmt.Println("c.conn.Write err", err)
				break
			}
			fmt.Println("input msg (Type \"exit\" to exit)")
			fmt.Scanln(&msg)
		}

		c.GetList()
		fmt.Println("input username (Type \"exit\" to exit)")
		fmt.Scanln(&username)
	}

}

// Menu 菜单
func (c *Client) Menu() bool {
	var input int
	fmt.Println("1:broadcast")
	fmt.Println("2:private")
	fmt.Println("3:rename")
	fmt.Println("0:exit")

	fmt.Scanln(&input)
	if input >= 0 && input <= 3 {
		c.flag = input
		return true
	}
	fmt.Println("Input format error")
	return false

}

// Run 运行
func (c *Client) Run() {

	select {
	case <-c.online:
		return
	default:
		for c.flag != 0 {
			for c.Menu() != true {
			}
			switch c.flag {
			case 1:
				fmt.Println("broadcast")
				c.BroadCast()
			case 2:
				fmt.Println("private")
				c.PrivateMessage()
			case 3:
				fmt.Println("rename")
				c.UpdateName()
			}

		}

	}

}

// IP ...
var IP string
var port int

func init() {
	flag.StringVar(&IP, "IP", "127.0.0.1", "IP(default:127.0.0.1)")
	flag.IntVar(&port, "port", 8080, "port(default:8080)")

}

func main() {
	client := NewClient(IP, port)
	if client == nil {
		fmt.Println("connect failed")
		return
	}
	fmt.Println("connect success")
	go client.DoResponse()

	client.Run()
	fmt.Println("exit...")

}
