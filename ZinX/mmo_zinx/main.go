package main

import (
	"ZinX/mmo_zinx/apis"
	"ZinX/mmo_zinx/core"
	"ZinX/zinx/ziface"
	"ZinX/zinx/znet"
	"fmt"
)

func OnConnectionAdd(conn ziface.IConnection) {
	//创建一个Player对象
	player := core.NewPlayer(conn)

	//给客户端发送MsgID：1的消息
	player.SyncPid()

	//给客户端发送MsgID：200的消息
	player.BroadCastStartPosition()

	//将当前新上线的玩家添加到WorldManager中
	core.WorldMgrObj.AddPlayer(player)

	//将该连接绑定属性pid
	conn.SetProperty("pid", player.Pid)

	//同步周边玩家上线信息，显示周边玩家信息
	player.SyncSurrounding()

	fmt.Println("===========> Player pid = ", player.Pid, "arrived  <============")
}

// 当客户端断开连接的时候的hook函数
func OnConnectionLost(conn ziface.IConnection) {
	//获取当前连接的Pid属性
	pid, _ := conn.GetProperty("pid")

	//根据pid获取对应的玩家对象
	player := core.WorldMgrObj.GetPlayerByPid(pid.(int32))

	//触发玩家下线业务
	if pid != nil {
		player.LostConnection()
	}

	fmt.Println("================> player ", pid, "leave <==========")
}

func main() {
	//创建zinx server句柄
	s := znet.NewServer("MMO Game Zinx")

	//连接创建和销毁的HOOK钩子函数
	s.SetOnConnStart(OnConnectionAdd)
	s.SetOnConnStop(OnConnectionLost)

	//注册一些路由业务
	s.AddRouter(2, &apis.WorldChatApi{})
	s.AddRouter(3, &apis.MoveApi{})
	//启动服务
	s.Serve()

}
