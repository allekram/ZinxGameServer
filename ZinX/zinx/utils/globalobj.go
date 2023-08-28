package utils

import (
	"ZinX/zinx/ziface"
	"encoding/json"
	"os"
)

/*
	存储一切有关zinx框架的全局参数，供其他模块使用
	一些参数可以通过zinx.json用户进行配置
*/

type GlobalObj struct {
	//server
	TcpServer ziface.IServer
	Host      string
	TcpPort   int
	Name      string //服务器名称

	//zinx
	Version          string
	MaxConn          int    //当前服务器允许的最大链接数
	MaxPackageSize   uint32 //当前Zinx框架数据包的最大值
	WorkerPoolSize   uint32 //当前业务工作Worker池的Goroutine的数量
	MaxWorkerTaskLen uint32 //Zinx框架允许用户最多开辟多少个Worker（限定条件）
}

// 定义一个全局的对外GolobalObj
var GlobalObject *GlobalObj

func (g *GlobalObj) Reload() {
	data, err := os.ReadFile("conf/zinx.json")
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(data, &GlobalObject)
	if err != nil {
		panic(err)
	}
}

func init() {
	//如果配置文件没有加载，默认的值
	GlobalObject = &GlobalObj{
		Name:             "ZinxServerApp",
		Version:          "V1.0",
		TcpPort:          8999,
		Host:             "0.0.0.0",
		MaxConn:          1000,
		MaxPackageSize:   4096,
		WorkerPoolSize:   10,
		MaxWorkerTaskLen: 1024,
	}

	//应该尝试从conf/zinx.json去加载用户自定义的参数
	GlobalObject.Reload()
}
