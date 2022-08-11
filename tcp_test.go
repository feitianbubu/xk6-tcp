package tcp

import (
	"fmt"
	"testing"

	"tevat.nd.org/framework/proxy"
)

func TestProxyTCP(t *testing.T) {
	for i := 0; i < 5000; i++ {
		tcpSend()
	}
}

var ID uint32 = 1

func tcpSend() {
	client := &Module{}
	conn, err := client.Connect("127.0.0.1:12345")
	if err != nil {
		fmt.Printf("connect fail: %s \n", err.Error())
		return
	}
	defer conn.Close()
	ID++
	req := &Request{
		ID:     ID,
		Method: "/api/v1/config",
		Msg:    "{}",
	}
	err = client.Send(conn, req)
	if err != nil {
		fmt.Printf("send fail: %s \n", err.Error())
		return
	}
	v, err := client.Recv(conn)
	if err != nil {
		fmt.Printf("recv fail: %s \n", err.Error())
		return
	}
	fmt.Printf("recv: %d \n", v.(proxy.Request).ID)

}

//func tcpSend() {
//	addr := "127.0.0.1:12345"
//	client := &Module{}
//	handler, err := client.Connect(addr)
//	if err != nil {
//		fmt.Printf("connect fail: %s \n", err.Error())
//		return
//	}
//	fmt.Println("connect success", handler)
//	method := "/api/v1/config"
//	msg := `{}`
//	reqMap := &Request{
//		Method: method,
//		Msg:    msg,
//	}
//	resp, err := client.Send(handler, reqMap)
//	if err != nil {
//		fmt.Println(err)
//	}
//
//	if v, ok := resp.(proxy.Request); ok {
//		fmt.Println(v.ID)
//	}
//	fmt.Println("===================done")
//}
