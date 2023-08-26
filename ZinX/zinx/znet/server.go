package znet

import (
	"fmt"
	"net"
	"zinx0.1/zinx/utils"
	"zinx0.1/zinx/ziface"
)

// iServer的接口实现，定义一个Server的服务器模块
type Server struct {
	//服务器名称
	Name string
	//服务器绑定的IP版本
	IPVersion string
	//服务器监听的IP
	IP string
	//服务器监听的端口
	Port int
	//当前server的消息管理模块，用来绑定MsgID和对应的处理业务API关系
	MsgHandle ziface.IMsgHandle
	//该server的连接管理器
	ConnMgr ziface.IConnManager
	//该Server销毁连接之前自动调用的Hook函数--OnConnStart
	OnConnStart func(conn ziface.IConnection)
	//该Server销毁连接之前自动调用的Hook函数--OnConnStop
	OnConnStop func(conn ziface.IConnection)
}

func (s *Server) AddRouter(msgID uint32, router ziface.IRouter) {
	s.MsgHandle.AddRouter(msgID, router)
	fmt.Println("Add Router Succ!")
}

// 初始化Server模块的方法
func NewServer(name string) ziface.IServer {
	s := &Server{
		Name:      utils.GlobalObject.Name,
		IPVersion: "tcp4",
		IP:        utils.GlobalObject.Host,
		Port:      utils.GlobalObject.TcpPort,
		MsgHandle: NewMsgHandle(),
		ConnMgr:   NewConnManager(),
	}

	return s
}

func (s *Server) Start() {
	fmt.Printf("[zinx] Server Name : %s,listenner at IP : %s,Port: %d is starting \n",
		utils.GlobalObject.Name, utils.GlobalObject.Host, utils.GlobalObject.TcpPort)

	fmt.Printf("[zinx] Version %s,MaxConn : %d, MaxPackeetSize : %d \n",
		utils.GlobalObject.Version, utils.GlobalObject.MaxConn, utils.GlobalObject.MaxPackageSize)

	go func() {
		//开启消息队列及worker工作池
		s.MsgHandle.StartWorkerPool()

		addr, err := net.ResolveTCPAddr(s.IPVersion, fmt.Sprintf("%s:%d", s.IP, s.Port))
		if err != nil {
			fmt.Println("resolve tcp addr error :", err)
			return
		}

		//2.监听服务器的地址
		listenner, err := net.ListenTCP(s.IPVersion, addr)
		if err != nil {
			fmt.Println("listen")
			return
		}

		fmt.Println("start zinx server succ,", s.Name, "succ,listening...")

		var cid uint32
		cid = 0

		//3.阻塞的等待客户端连接，处理客户端连接业务
		for {
			//如果有客户端连接，阻塞会返回
			conn, err := listenner.AcceptTCP()
			if err != nil {
				fmt.Println("Accept err:", err)
				continue
			}

			//设置最大连接个数的判断，如果超过最大连接的数量，那么则关闭新连接
			if s.ConnMgr.Len() >= utils.GlobalObject.MaxConn {
				fmt.Println("=========> Too Many Connections MaxConn = ", utils.GlobalObject.MaxConn)
				conn.Close()
				continue
			}

			//将处理新连接的业务方法和conn进行绑定，得到我们的连接模块
			dealConn := NewConnection(s, conn, cid, s.MsgHandle)
			cid++

			//启动当前的连接业务模块
			dealConn.Start()
		}
	}()
}

// 停止服务器
func (s *Server) Stop() {
	//将一些服务器的资源、状态或者一些已经开辟的连接信息进行停止或者回收
	fmt.Println("[STOP] Zinx server name ", s.Name)
	s.ConnMgr.ClearConn()
}

func (s *Server) Serve() {
	//启动server的服务功能
	s.Start()

	//todo 做一些启动服务器之后的额外业务

	//阻塞状态
	select {}
}

func (s *Server) GetConnMsg() ziface.IConnManager {
	return s.ConnMgr
}

// 注册OnConnStart
func (s *Server) SetOnConnStart(hookFunc func(connection ziface.IConnection)) {
	s.OnConnStart = hookFunc
}

// 注册OnConnStop
func (s *Server) SetOnConnStop(hookFunc func(connection ziface.IConnection)) {
	s.OnConnStop = hookFunc
}

// 调用OnConnStart
func (s *Server) CallOnConnStart(connection ziface.IConnection) {
	if s.OnConnStart != nil {
		fmt.Println("--------->Call OnConnStart...")
		s.OnConnStart(connection)
	}
}

// 调用OnConnStop
func (s *Server) CallOnConnStop(connection ziface.IConnection) {
	if s.OnConnStop != nil {
		fmt.Println("--------->Call OnConnStop ...")
		s.OnConnStop(connection)
	}
}
