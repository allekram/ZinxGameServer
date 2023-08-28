package core

import (
	"ZinX/mmo_zinx/pb"
	"ZinX/zinx/ziface"
	"fmt"
	"google.golang.org/protobuf/proto"
	"math/rand"
	"sync"
)

// 玩家对象
type Player struct {
	Pid  int32              //玩家ID
	Conn ziface.IConnection //当前玩家与客户端的连接
	X    float32            //平面的x坐标
	Y    float32            //高度
	Z    float32            //平面的y坐标
	V    float32            //旋转角度
}

/*
player ID 生成器
*/
var PidGen int32 = 1
var IdLock sync.Mutex

// 创建玩家的方法
func NewPlayer(conn ziface.IConnection) *Player {
	//生成一个玩家ID
	IdLock.Lock()
	id := PidGen
	PidGen++
	IdLock.Unlock()

	//创建一个玩家对象
	p := &Player{
		Pid:  id,
		Conn: conn,
		X:    float32(160 + rand.Intn(10)),
		Y:    0,
		Z:    float32(134 + rand.Intn(17)),
		V:    0,
	}
	return p
}

/*
	提供一个发送给客户端消息的方法
	主要是将pb的protobuf数据序列化之后，再调用zinx的SendMsg方法
*/

func (p *Player) SendMsg(msgId uint32, data proto.Message) {
	//将proto Message结构体序列化 转换成二进制
	msg, err := proto.Marshal(data)
	if err != nil {
		fmt.Println("marshal msg err : ", err)
		return
	}
	//将二进制文件 通过zinx框架的sendmsg将数据发送给客户端
	if p.Conn == nil {
		fmt.Println("connection in player is nil")
		return
	}

	if err := p.Conn.SendMsg(msgId, msg); err != nil {
		fmt.Println("Player SendMsg error !")
		return
	}
	return
}

// 告知哭护短玩家Pid，同步已经生成的玩家ID给客户端
func (p *Player) SyncPid() {
	//组建MsgID：0的proto数据
	data := &pb.SyncPid{
		Pid: p.Pid,
	}

	//将消息发送给客户端
	p.SendMsg(1, data)
}

// 广播玩家自己的出生地点
func (p *Player) BroadCastStartPosition() {
	msg := &pb.BroadCast{
		Pid: p.Pid,
		Tp:  2, //Tp=2代表广播坐标
		Data: &pb.BroadCast_P{
			&pb.Position{
				X: p.X,
				Y: p.Y,
				Z: p.Z,
				V: p.V,
			},
		},
	}

	p.SendMsg(200, msg)
}

func (p *Player) Talk(content string) {
	//1. 组件MsgID 200 proto数据
	msg := &pb.BroadCast{
		Pid: p.Pid,
		Tp:  1, //Tp 1 代表聊天广播
		Data: &pb.BroadCast_Content{
			Content: content,
		},
	}

	//2. 得到当前世界所有的在线玩家
	Players := WorldMgrObj.GetAllPlayers()

	//3.向所有的玩家发送MsgID：200消息
	for _, Player := range Players {
		Player.SendMsg(200, msg)
	}
}

func (p *Player) SyncSurrounding() {
	//1.获取当前玩家周围的玩家有哪些（九宫格）
	pids := WorldMgrObj.AOIMgr.GetPidsByPos(p.X, p.Z)

	//2. 根据pid得到所有玩家对象
	players := make([]*Player, 0, len(pids))

	//3. 给这些玩家发送MsgID：200消息，让自己出现在对方视野中
	for _, pid := range pids {
		players = append(players, WorldMgrObj.GetPlayerByPid(int32(pid)))
	}

	//3.1 组件MshID200 proto数据
	msg := &pb.BroadCast{
		Pid: p.Pid,
		Tp:  2, //TP 2 代表广播数据
		Data: &pb.BroadCast_P{
			P: &pb.Position{
				X: p.X,
				Y: p.Y,
				Z: p.Z,
				V: p.V,
			},
		},
	}

	//3.2 每个玩家分别给对应的客户端发送200消息，显示人物
	for _, player := range players {
		player.SendMsg(200, msg)
	}

	//4 让周围九宫格内的玩家出现在自己的视野中
	//4.1 制作Message SyncPlayers 数据
	playersData := make([]*pb.Player, 0, len(players))
	for _, player := range players {
		p := &pb.Player{
			Pid: player.Pid,
			P: &pb.Position{
				X: player.X,
				Y: player.Y,
				Z: player.Z,
				V: player.V,
			},
		}
		playersData = append(playersData, p)
	}

	//4.2 封装SyncPlayer protobuf数据
	SyncPlayersMsg := &pb.SyncPlayers{
		Ps: playersData[:],
	}

	//4.3 给当前玩家发送需要显示周围的全部玩家数据
	p.SendMsg(202, SyncPlayersMsg)
}

func (p *Player) UpdatePos(x float32, y float32, z float32, v float32) {
	// 更新玩家的位置信息
	p.X = x
	p.Y = y
	p.Z = z
	p.V = v

	//组装protobuf协议 ， 发送位置给周围玩家
	msg := &pb.BroadCast{
		Pid: p.Pid,
		Tp:  4, //tp 4 表示移动
		Data: &pb.BroadCast_P{
			P: &pb.Position{
				X: p.X,
				Y: p.Y,
				Z: p.Z,
				V: p.V,
			},
		},
	}

	//获取当前玩家周边的所有玩家
	players := p.GetSurroundingPlayers()

	//向周边的每个玩家发送Msg:200消息，移动位置更新消息
	for _, player := range players {
		player.SendMsg(200, msg)
	}
}

func (p *Player) GetSurroundingPlayers() []*Player {
	//得到当前AOI区域的所有pid
	pids := WorldMgrObj.AOIMgr.GetPidsByPos(p.X, p.Z)

	//将所有pid对应的player放到player切片中
	players := make([]*Player, 0, len(pids))
	for _, pid := range pids {
		players = append(players, WorldMgrObj.GetPlayerByPid(int32(pid)))
	}
	return players
}

func (p *Player) LostConnection() {
	//1 获取周围AOI九宫格内的玩家
	Players := p.GetSurroundingPlayers()

	//2 封装MsgID：201消息
	msg := &pb.SyncPid{
		Pid: p.Pid,
	}

	//3 向周围的玩家发送消息
	for _, player := range Players {
		player.SendMsg(201, msg)
	}

	//4 世界管理器将当前玩家从AOI中移除
	WorldMgrObj.AOIMgr.RemoveFromGridByPos(int(p.Pid), p.X, p.Z)
	WorldMgrObj.RemovePlayerByPid(p.Pid)
}
