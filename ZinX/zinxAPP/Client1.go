package main

import (
	"fmt"
	"io"
	"net"
	"time"
	"zinx0.1/zinx/znet"
)

func main() {
	fmt.Printf("client start ...")

	time.Sleep(1 * time.Second)

	//直接连接远程服务器，得到一个conn链接
	conn, err := net.Dial("tcp", "127.0.0.1:8999")
	if err != nil {
		fmt.Println("client1 start err , exit!")
		return
	}

	for {
		//发送封包的message消息  MsgID：0
		dp := znet.NewDataPack()
		binaryMsg, err := dp.Pack(znet.NewMsgPackage(1, []byte("zinxV0.6 client test message")))
		if err != nil {
			fmt.Println("Pack error:", err)
			return
		}

		_, err = conn.Write(binaryMsg)
		if err != nil {
			fmt.Println("write error", err)
			return
		}

		//服务器给我们回复一个message数据

		//先读取流中的head部分，得到ID 和dataLen

		//再根据DataLen进行第二次读取，将data读出来
		binaryHead := make([]byte, dp.GetHeadLen())
		if _, err := io.ReadFull(conn, binaryHead); err != nil {
			fmt.Println("read head error", err)
			break
		}

		//将二进制的head拆包到msg结构体中
		msgHead, err := dp.Unpack(binaryHead)
		if err != nil {
			fmt.Println("client1 unpack msgHead error", err)
			return
		}

		if msgHead.GetMsgLen() > 0 {
			//第二次数据读取
			msg := msgHead.(*znet.Message)
			msg.Data = make([]byte, msg.GetMsgLen())

			if _, err := io.ReadFull(conn, msg.Data); err != nil {
				fmt.Println("read msg data error , ", err)
				return
			}

			fmt.Println("————> Recv Server Msg : ID = ", msg.Id, "Len = ",
				msg.DataLen, "data = ", string(msg.Data))
		}

		if err != nil {
			return
		}

		//阻塞
		time.Sleep(1 * time.Second)
	}
}
