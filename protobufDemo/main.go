package main

import (
	"fmt"
	"google.golang.org/protobuf/proto"
	"protobufDemo/pb"
)

func main() {
	person := &pb.Person{
		Name:   "ekk",
		Age:    16,
		Emails: []string{"ekram@4593.com"},
		Phones: []*pb.PhoneNumber{
			&pb.PhoneNumber{
				Number: "13131313",
				Type:   pb.PhoneType_HOME,
			},
		},
	}

	//将对象进行序列化
	data, err := proto.Marshal(person)
	if err != nil {
		fmt.Printf("marshal err: ", err)
	}

	newdata := &pb.Person{}
	err = proto.Unmarshal(data, newdata)
	if err != nil {
		fmt.Println("unmarshal err: ", err)
	}
	fmt.Println("源数据 ： ", person)
	fmt.Println("解码之后的数据： ", newdata)

}
