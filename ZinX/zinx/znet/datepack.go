package znet

import (
	"ZinX/zinx/utils"
	"ZinX/zinx/ziface"
	"bytes"
	"encoding/binary"
	"errors"
)

// 封包、拆包的模块
type DataPack struct{}

// 拆包封包实例的一个初始化方法
func NewDataPack() *DataPack {
	return &DataPack{}
}

// 获取包的头的长度方法
func (dp *DataPack) GetHeadLen() uint32 {
	//Datalen uint32 (4字节) + ID uint32(4字节)
	return 8
}

// 封包方法
//
//	|dataLen|msgID|data|
func (dp *DataPack) Pack(msg ziface.IMessage) ([]byte, error) {
	//创建一个存放byte字节的缓冲
	dataBuff := bytes.NewBuffer([]byte{})

	//将datalen写进databuff中
	err := binary.Write(dataBuff, binary.LittleEndian, msg.GetMsgLen())
	if err != nil {
		return nil, err
	}

	//将MsgId 写进databuff中
	err = binary.Write(dataBuff, binary.LittleEndian, msg.GetMsgId())
	if err != nil {
		return nil, err
	}

	//将data数据写进databuff中
	err = binary.Write(dataBuff, binary.LittleEndian, msg.GetData())
	if err != nil {
		return nil, err
	}
	return dataBuff.Bytes(), nil
}

// 拆包方法（将包的Head信息读出来） 之后根据head信息里的data的长度，再进行一次读
func (dp *DataPack) Unpack(binaryData []byte) (ziface.IMessage, error) {
	//创建一个从输入二进制数据的ioreader
	dataBuff := bytes.NewReader(binaryData)

	//直接压head信息，得到datalen和MsgID
	msg := &Message{}

	//读dataLen
	err := binary.Read(dataBuff, binary.LittleEndian, &msg.DataLen)
	if err != nil {
		return nil, err
	}

	//读MsgID
	err = binary.Read(dataBuff, binary.LittleEndian, &msg.Id)
	if err != nil {
		return nil, err
	}

	if utils.GlobalObject.MaxPackageSize > 0 && msg.DataLen > utils.GlobalObject.MaxPackageSize {
		return nil, errors.New("too Large msg data recv!")
	}

	return msg, nil
}
