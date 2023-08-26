package znet

import (
	"fmt"
	"io"
	"net"
	"testing"
)

/*
运行测试需要针对错误信息对globalobj.go的内容进行注释
*/
func TestDataPack(t *testing.T) {
	/*
		模拟的服务器
	*/
	listenner, err := net.Listen("tcp", "127.0.0.1:8999")
	if err != nil {
		fmt.Println("server listen err:", err)
		return
	}

	//创建一个go，承载负责从客户端处理业务
	go func() {
		//从客户端读数据，拆包处理
		for {
			conn, err := listenner.Accept()
			if err != nil {
				fmt.Println("server accept error :", err)
			}

			go func(conn net.Conn) {
				//处理客户端的请求
				//---------拆包的过程------------

				dp := NewDataPack()
				for {
					//第一次从conn读，读取head
					headData := make([]byte, dp.GetHeadLen())
					_, err := io.ReadFull(conn, headData)
					if err != nil {
						fmt.Println("read head error")
						break
					}

					msgHead, err := dp.Unpack(headData)
					if err != nil {
						fmt.Println("server unpack err", err)
						return
					}

					if msgHead.GetMsgLen() > 0 {
						//msg有数据，需要进行第二次读取
						//第二次从conn读，读data
						msg := msgHead.(*Message)
						msg.Data = make([]byte, msg.GetMsgLen())

						//根据datalen的长度再次从io流中读取
						_, err := io.ReadFull(conn, msg.Data)
						if err != nil {
							fmt.Println("server unpack data err:", err)
							return
						}

						//完整的一个消息已经读取完毕
						fmt.Println("Recv MsgID:", msg.Id, "datalen= ", msg.DataLen, "data = ", string(msg.Data))
					}

				}

			}(conn)
		}
	}()

	/*
		模拟的客户端
	*/
	conn, err := net.Dial("tcp", "127.0.0.1:8999")
	if err != nil {
		fmt.Println("client dial err :", err)
		return
	}

	//创建一个封包对象dp
	dp := NewDataPack()

	//模拟粘包过程，封装两个msg一起发送
	//封装第一个msg1包
	msg1 := &Message{
		Id:      1,
		DataLen: 4,
		Data:    []byte{'z', 'i', 'n', 'x'},
	}
	sendData1, err := dp.Pack(msg1)
	if err != nil {
		fmt.Println("client pack msg1 error:", err)
		return
	}
	//封装第二个msg2包
	msg2 := &Message{
		Id:      2,
		DataLen: 7,
		Data:    []byte{'n', 'i', 'h', 'a', 'o', '!', '!'},
	}
	sendData2, err := dp.Pack(msg2)
	if err != nil {
		fmt.Println("client pack msg2 error:", err)
		return
	}
	//将两个包黏在一起
	sendData1 = append(sendData1, sendData2...)
	//一次性发送给服务端
	_, err = conn.Write(sendData1)
	if err != nil {
		return
	}

	//客户端阻塞
	select {}
}
