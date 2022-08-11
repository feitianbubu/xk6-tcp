package tcp

import (
	"fmt"
	"net"

	"go.k6.io/k6/js/modules"
	"tevat.nd.org/framework/proxy"
)

type (
	Module struct {
		vu modules.VU
	}
	RootModule struct{}
)

// Ensure the interfaces are implemented correctly.
var (
	_ modules.Module   = &RootModule{}
	_ modules.Instance = &Module{}
)

func (*RootModule) NewModuleInstance(vu modules.VU) modules.Instance {
	return &Module{
		vu: vu,
	}
}
func (m *Module) Exports() modules.Exports {
	return modules.Exports{
		Default: m,
	}
}

func init() {
	modules.Register("k6/x/tcp", &RootModule{})
}

func (m *Module) Connect(addr string) (net.Conn, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Printf("connect fail: %s \n", err.Error())
		return nil, err
	}
	return conn, nil
}

type Request struct {
	ID     uint32
	Method string
	Msg    string
}

var codec = &proxy.Codec{}

func (m *Module) Send(conn net.Conn, sReq *Request) error {
	req := &proxy.Request{
		ID:     sReq.ID,
		Method: []byte(sReq.Method),
		Msg:    []byte(sReq.Msg),
	}
	err := codec.Encode(conn, req)
	if err != nil {
		fmt.Printf("send fail: %s \n", err.Error())
		return err
	}
	return nil
}

func (m *Module) Recv(conn net.Conn) (any, error) {
	v, err := codec.Decode(conn)
	if err != nil {
		fmt.Printf("recv fail: %s \n", err.Error())
		return nil, err
	}
	return v, nil
}
func (m *Module) Close(conn net.Conn) error {
	return conn.Close()
}

//func (m *Module) Send2(handler *xtcp.DefaultClientHandler, sReq *Request) (any, error) {
//	//fmt.Printf("send: %s %s %s\n", addr, sReq.Method, sReq.Msg)
//
//	var resChan = make(chan any)
//	var res any
//
//	addr := "127.0.0.1:12345"
//	handler, _ = m.Connect(addr)
//	fmt.Println("connect success", handler)
//
//	handler.HandlePeerFunc(func(peer *xtcp.Peer) {
//		defer peer.Close()
//		if peer == nil {
//			fmt.Println("peer is nil")
//			return
//		}
//
//		req := &proxy.Request{
//			Method: []byte(sReq.Method),
//			Msg:    []byte(sReq.Msg),
//		}
//		peer.SetDeadline(time.Now().Add(time.Second * 3))
//		err := peer.Send(req)
//		if err != nil {
//			fmt.Println(err)
//			return
//		}
//
//		res, err = peer.Recv()
//		if err != nil {
//			fmt.Printf("perr recv fail: %s \n", err.Error())
//		}
//		resChan <- res
//		//fmt.Printf("res: %s \n", res)
//		//wg.Done()
//	})
//	return <-resChan, nil
//	//job.Wait()
//	//res = <-resChan
//	//
//	//return res, nil
//}

//func (m *Module) SendDebug(addr string, method string, msg string) (any, error) {
//	res, err := m.Send(addr, method, msg)
//	if err != nil {
//		return nil, err
//	}
//	resM := make(map[string]interface{})
//	if v, ok := res.(proxy.Request); ok {
//		resM["id"] = v.ID
//		resM["method"] = string(v.Method)
//		resM["metadata"] = v.Metadata
//		resM["msg"] = string(v.Msg)
//	} else {
//		return nil, fmt.Errorf("res is not proxy.Response")
//	}
//	return resM, nil
//}
