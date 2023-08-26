package znet

import (
	"fmt"
	"strconv"
	"zinx0.1/zinx/utils"
	"zinx0.1/zinx/ziface"
)

/*
	消息处理模块的实现
*/

type MsgHandle struct {
	//存放每个MsgID对应的处理方法
	Apis map[uint32]ziface.IRouter

	//负责Worker取任务的消息队列
	TaskQueue []chan ziface.IRequest

	//业务工作Worker池的worker数量
	WorkerPoolSize uint32
}

// 初始化
func NewMsgHandle() *MsgHandle {
	return &MsgHandle{
		Apis:           make(map[uint32]ziface.IRouter),
		WorkerPoolSize: utils.GlobalObject.WorkerPoolSize, //从全局配置种获取
		TaskQueue:      make([]chan ziface.IRequest, utils.GlobalObject.WorkerPoolSize),
	}
}

// 调度/执行对应的Router消息处理方法
func (mh *MsgHandle) DoMsgHandler(request ziface.IRequest) {
	//1 从Request中找到msgID
	handle, ok := mh.Apis[request.GetMsgID()]
	if !ok {
		fmt.Println("api msgID =", request.GetMsgID(), "is NOT FOUND ! Need Register !")
	}
	//2根据MsgID调度对应的router即可
	handle.PreHandle(request)
	handle.Handle(request)
	handle.PostHandle(request)
}

// 为消息添加具体的处理逻辑
func (mh *MsgHandle) AddRouter(msgID uint32, router ziface.IRouter) {
	//1 判断当前msg绑定的API处理方法是否存在
	if _, ok := mh.Apis[msgID]; ok {
		//id 已经注册
		panic("repeat api , msgID = " + strconv.Itoa(int(msgID)))
	}

	//2 添加msg与API的绑定关系
	mh.Apis[msgID] = router
	fmt.Println("Add api MsgID = ", msgID, "succ !")
}

// 启动一个Worker工作池
func (mh *MsgHandle) StartWorkerPool() {
	//根据workerPoolSize，分别开启Worker，每个Worker用一个go来承载
	for i := 0; i < int(mh.WorkerPoolSize); i++ {
		//一个worker被启动
		//1当前的worker嘴硬的channel消息队列，开辟空间
		mh.TaskQueue[i] = make(chan ziface.IRequest, utils.GlobalObject.MaxWorkerTaskLen)

		//2启动当前的worker，阻塞的等待消息从channel传递进来
		go mh.StartOneWorker(i, mh.TaskQueue[i])

	}
}

// 启动一个Worker工作流程
func (mh *MsgHandle) StartOneWorker(workerID int, taskQueue chan ziface.IRequest) {
	fmt.Println("Worker ID = ", workerID, "is started ...")

	for {
		select {
		case request := <-taskQueue:
			mh.DoMsgHandler(request)
		}
	}
}

// 将消息交给TaskQueue，由Worker进行处理
func (mh *MsgHandle) SendMsgToTaskQueue(request ziface.IRequest) {
	//1 将消息平均分配给不同的worker
	//根据客户端建立的ConnID来进行分配
	//基本的平均分配的轮询法则
	workerID := request.GetConnection().GetConnID() % mh.WorkerPoolSize
	fmt.Println("Add ConnID = ", request.GetConnection().GetConnID(),
		"request MsgID = ", request.GetMsgID(),
		"to WorkerID = ", workerID)
	//2 将消息发送给对应的worker的TaskQueue即可
	mh.TaskQueue[workerID] <- request
}
