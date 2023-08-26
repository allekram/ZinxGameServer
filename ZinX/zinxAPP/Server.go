package main

import (
	"fmt"
	"zinx0.1/zinx/ziface"
	"zinx0.1/zinx/znet"
)

// ping test 自定义路由
type PingRouter struct {
	znet.BaseRouter
}

func (this *PingRouter) PostHandle(request ziface.IRequest) {
}

// Test Handle
func (this *PingRouter) Handle(request ziface.IRequest) {
	fmt.Println("Call PingRouter Handle")
	//先读取客户端的数据，再会写ping...ping...ping

	fmt.Println("recv from client : msgID= ", request.GetMsgID(),
		",data = ", string(request.GetData()))

	err := request.GetConnection().SendMsg(200, []byte("ping...ping...ping"))
	if err != nil {
		fmt.Println(err)
		return
	}
}

// Hello Zinx test 自定义路由
type HelloZinxRouter struct {
	znet.BaseRouter
}

func (this *HelloZinxRouter) PostHandle(request ziface.IRequest) {
}

// Test Handle
func (this *HelloZinxRouter) Handle(request ziface.IRequest) {
	fmt.Println("Call HelloZinxRouter Handle")
	//先读取客户端的数据，再会写ping...ping...ping

	fmt.Println("recv from client : msgID= ", request.GetMsgID(),
		",data = ", string(request.GetData()))

	err := request.GetConnection().SendMsg(201, []byte("Hello Zinx ! Welcome to Zinx ! "))
	if err != nil {
		fmt.Println(err)
		return
	}
}

// 创建连接之后执行的钩子函数
func DoConnectionBegin(conn ziface.IConnection) {
	fmt.Println("=========>DoConnectionBegin is Called ...")
	err := conn.SendMsg(202, []byte("DoConnection BEGIN"))
	if err != nil {
		fmt.Println(err)
	}

	//给当前的连接设置一些属性
	fmt.Println("Set conn property ...")
	conn.SetProperty("Name", "------Ekk-----")
	conn.SetProperty("BiliBili", "https://www.bilibili.com/video/BV1wE411d7th?p=47&vd_source=410d796092f3a792db1ee107322f8122")
}

// 连接断开之前的需要执行的函数
func DoConnectionLost(conn ziface.IConnection) {
	fmt.Println("=========>DoConnectionLost is Call...")
	fmt.Println("Conn ID = ", conn.GetConnID(), "is Lost ...")

	//获取连接属性
	if name, err := conn.GetProperty("Name"); err == nil {
		fmt.Println("Name", name)
	}

	if BiliBili, err := conn.GetProperty("BiliBili"); err == nil {
		fmt.Println("BiliBili", BiliBili)
	}

}

func main() {
	//创建server句柄
	s := znet.NewServer("[zinx V0.6]")

	//注册连接Hook钩子函数
	s.SetOnConnStart(DoConnectionBegin)
	s.SetOnConnStop(DoConnectionLost)

	//给当前zinx框架添加自定义的router
	s.AddRouter(0, &PingRouter{})
	s.AddRouter(1, &HelloZinxRouter{})

	//启动server
	s.Serve()
}
