package main

import (
	"net"
	"strings"
)

// User 用户
type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server
}

// NewUser 创建一个 user
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}
	// 启动 user 消息监听协程
	go user.ListenMessage()
	return user
}

// Online 上线
func (u *User) Online() {

	u.server.mapLock.Lock()
	u.server.OnlineMap[u.Name] = u
	u.server.mapLock.Unlock()

	// 广播消息
	u.server.BroadCast(u, "online")
}

// Offline 下线
func (u *User) Offline() {
	u.server.mapLock.Lock()
	delete(u.server.OnlineMap, u.Name)
	u.server.mapLock.Unlock()

	// 广播消息
	u.server.BroadCast(u, "offline")
}

// SendPrivateMessage 私聊
func (u *User) SendPrivateMessage(msg string) {
	u.conn.Write([]byte(msg))
}

// SendMessage 发送消息
func (u *User) SendMessage(msg string) {
	if msg == "onlineList" {
		u.server.mapLock.Lock()
		allMsg := ""
		for _, user := range u.server.OnlineMap {
			userMsg := "[" + u.Addr + "]" + user.Name + ":" + "online\n"
			allMsg += userMsg
		}
		u.server.mapLock.Unlock()
		u.SendPrivateMessage(allMsg)

	} else if len(msg) > 7 && msg[:7] == "rename|" { //改名
		newName := strings.Split(msg, "|")[1]
		_, ok := u.server.OnlineMap[newName]
		if ok {
			u.SendPrivateMessage("this username already exists\n")
		} else {
			u.server.mapLock.Lock()
			delete(u.server.OnlineMap, u.Name)
			u.Name = newName
			u.server.OnlineMap[newName] = u
			u.server.mapLock.Unlock()
			u.SendPrivateMessage("update username success\n")
		}

	} else if len(msg) > 3 && msg[:3] == "to|" { //私聊
		toUsername := strings.Split(msg, "|")[1]
		if toUsername == "" {
			u.SendPrivateMessage("format is err,please use \"to|username|msg\"")
			return
		}

		toUser, ok := u.server.OnlineMap[toUsername]
		if ok {
			toUser.SendMessage(u.Name + " send:" + strings.Split(msg, "|")[2])
		} else {
			u.SendPrivateMessage(toUsername + ",is offline")
		}

	} else {
		u.server.BroadCast(u, msg)
	}

}

// ListenMessage 监听 user channel 的方法 有消息直接发送给客户端
func (u *User) ListenMessage() {
	for {
		msg := <-u.C
		u.conn.Write([]byte(msg + "\n"))
	}

}
