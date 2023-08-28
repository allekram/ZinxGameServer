package znet

import (
	"ZinX/zinx/utils"
	"ZinX/zinx/ziface"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
)

// 连接模块
type Connection struct {
	//当前Conn隶属于哪个Server
	TcpServer ziface.IServer

	//当前连接的socket，tcp套接字
	Conn *net.TCPConn

	//连接的ID
	ConnID uint32

	//当前的连接状态
	isClosed bool

	//告知当前连接已经退出/停止的channel
	ExitChan chan bool

	//无缓冲的管道，用于读、写Goroutine之间的消息通信
	msgChan chan []byte

	//消息的管理MsgID 和 对应的处理业务API关系
	MsgHandler ziface.IMsgHandle

	//连接属性集合
	property map[string]interface{}

	//保护连接属性的锁
	propertyLock sync.RWMutex
}

func NewConnection(server ziface.IServer, conn *net.TCPConn, connID uint32, msgHandler ziface.IMsgHandle) *Connection {
	c := &Connection{
		TcpServer:  server,
		Conn:       conn,
		ConnID:     connID,
		isClosed:   false,
		MsgHandler: msgHandler,
		msgChan:    make(chan []byte),
		ExitChan:   make(chan bool, 1),
		property:   make(map[string]interface{}),
	}

	//将conn加入到ConnManager中
	c.TcpServer.GetConnMsg().Add(c)
	return c
}

// 写消息的Goroutine，专门发送给客户端消息的模块
func (c *Connection) StartWrite() {
	fmt.Println("[Writer] Goroutine is running...")
	defer fmt.Println(c.RemoteAddr().String(), "[conn Writer exit!]")

	//不断的阻塞的等待channel的消息，写给客户端
	for {
		select {
		case data := <-c.msgChan:
			if _, err := c.Conn.Write(data); err != nil {
				fmt.Println("Send data error :", err)
				return
			}
		case <-c.ExitChan:
			//代表Reader已经退出，此时Writer也要退出
			return
		}
	}
}

// 连接的读业务方法
func (c *Connection) StartReader() {
	fmt.Println("[Reader] Goroutine is running...")
	defer fmt.Println("connID = ", c.ConnID, "Reader is exit,remote addr is", c.RemoteAddr().String())
	defer c.Stop()

	for {
		//创建一个拆包解包的对象
		dp := NewDataPack()

		//读取客户端的Msg Head 二进制流 8字节
		headData := make([]byte, dp.GetHeadLen())
		_, err := io.ReadFull(c.GetTCPConnection(), headData)
		if err != nil {
			fmt.Println("read msg head error:", err)
			break
		}

		//拆包，得到MsgID和msgDatalen，让在msg消息中
		msg, err := dp.Unpack(headData)
		if err != nil {
			fmt.Println("unpack error:", err)
			break
		}

		//根据dataLen，再次读取Data，放在msg.Data中
		var data []byte
		if msg.GetMsgLen() > 0 {
			data = make([]byte, msg.GetMsgLen())
			_, err := io.ReadFull(c.GetTCPConnection(), data)
			if err != nil {
				fmt.Println("read msg data error :", err)
				return
			}
		}

		msg.SetData(data)

		//得到当前conn数据的Request请求数据
		req := Request{
			conn: c,
			msg:  msg,
		}

		if utils.GlobalObject.WorkerPoolSize > 0 {
			//已经开启了工作池机制，将消息发送给worker工作池处理即可
			c.MsgHandler.SendMsgToTaskQueue(&req)
		} else {
			//从路由中，找到注册绑定的Conn对应的router调用
			//根据绑定好的MsgID找到对应处理api业务 执行
			go c.MsgHandler.DoMsgHandler(&req)
		}
	}
}

func (c *Connection) Start() {
	fmt.Println("Conn Start()..ConnID=", c.ConnID)

	//启动从当前连接的读数据的业务
	go c.StartReader()
	//TODO 启动从当前连接写数据的业务
	go c.StartWrite()

	//按照开发者传递进来的创建连接之后需要调用的处理业务，执行对应的Hook函数
	c.TcpServer.CallOnConnStart(c)
}

func (c *Connection) Stop() {
	println("Conn Stop()..ConnID=", c.ConnID)

	//如果当前连接已经关闭
	if c.isClosed == true {
		return
	}
	c.isClosed = true

	//调用开发者注册的 销毁连接之前 需要执行的业务Hook函数
	c.TcpServer.CallOnConnStop(c)

	//关闭socket连接
	c.Conn.Close()

	//告知Write关闭
	c.ExitChan <- true

	//将当前连接从ConnMsg中摘除掉
	c.TcpServer.GetConnMsg().Remove(c)

	//回收资源
	close(c.ExitChan)
	close(c.msgChan)

}

func (c *Connection) GetTCPConnection() *net.TCPConn {
	return c.Conn
}

func (c *Connection) GetConnID() uint32 {
	return c.ConnID
}

func (c *Connection) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

// 提供一个SendMsg方法，将我们要发送给客户端的数据，先进性封包，再发送
func (c *Connection) SendMsg(msgId uint32, data []byte) error {
	if c.isClosed == true {
		return errors.New("Connection closed when send msg")
	}

	//将data进行封包
	dp := NewDataPack()

	binaryMsg, err := dp.Pack(NewMsgPackage(msgId, data))
	if err != nil {
		fmt.Println("Pack error msg id = ", msgId)
		return errors.New("Pack error msg")
	}

	//将数据发送给Writer
	c.msgChan <- binaryMsg

	return nil
}

// 设置连接属性
func (c *Connection) SetProperty(key string, value interface{}) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	//添加一个连接属性
	c.property[key] = value
}

// 获取连接属性
func (c *Connection) GetProperty(key string) (interface{}, error) {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()

	//读取属性
	if value, ok := c.property[key]; ok {
		return value, nil
	} else {
		return nil, errors.New("no property found")
	}
}

// 移除连接属性
func (c *Connection) RemoveProperty(key string) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	//删除属性
	delete(c.property, key)
}
