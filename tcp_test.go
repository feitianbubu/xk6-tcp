package tcp

import (
	"fmt"
	"testing"

	"github.com/dop251/goja"
	"google.golang.org/protobuf/proto"
	pb "tevat.nd.org/framework/proxy/proto"
)

var client = &Module{}

//func OnRec(res *proxy.Request) {
//	fmt.Printf("[test]OnRec:%+v \n", res)
//}

var vm = goja.New()

func TestProxyTCP(t *testing.T) {

	//client.Connect("172.24.140.131:12345")
	addr := "172.24.140.131:12345"
	//addr = "127.0.0.1:12345"
	opts := Opts{}
	opts.MoveTimes = int64(10)
	opts.AccountId = "2"
	opts.WatchEnabled = true
	err := client.Start(addr, opts)
	if err != nil {
		fmt.Printf("start fail:%+v", err)
	}
	//loginRes := client.SendWithRes(getReqObject("login", nil))
	//loginMsg := client.Parse(loginRes.Msg)
	//uid := loginMsg["uid"].(float64)
	//client.StartOnRec(OnRec)
	//client.Send(getReqObject("watch", nil))
	//for i := 0; i < 1; i++ {
	//	location := vm.NewObject()
	//	location.Set("uid", uid)
	//	location.Set("x", 1)
	//	location.Set("y", 1)
	//	msg := vm.NewObject()
	//	msg.Set("location", location)
	//	client.Send(getReqObject("move", msg))
	//}
	//time.Sleep(time.Second * 10)
}

//var ID uint32 = 0

//	func rec() *proxy.Request {
//		return client.Rec()
//	}
func getReqObject(name string, msg *goja.Object) *goja.Object {
	//defer conn.Close()
	ID++
	reqMap := vm.NewObject()
	//reqMap.Set("method", "tevat.example.auth.Auth/login")
	reqMap.Set("id", ID)
	msgMap := vm.NewObject()
	switch name {
	case "login":
		reqMap.Set("method", "tevat.example.auth.Auth/login")
		msgMap.Set("account_id", fmt.Sprintf("%d", ID))
		msgMap.Set("account_token", "123456")
	case "watch":
		reqMap.Set("method", "/tevat.example.scene.Scene/WatchEvents")
	case "move":
		reqMap.Set("method", "/tevat.example.scene.Scene/Move")
	}
	if msg != nil {
		reqMap.Set("msg", msg)
	} else {
		reqMap.Set("msg", msgMap)
	}
	return reqMap
	//metaMap := vm.NewObject()
	//metaMap.Set("uid", "1")
	//metaMap.Set("session-uid", "1")
	//reqMap.Set("metadata", metaMap)
	//req := &proxy.Request{
	//	ID:     ID,
	//	Method: []byte("tevat.example.auth.Auth/login"),
	//	Msg:    []byte(`{"account_id":"` + fmt.Sprintf("%d", ID) + `","account_token":"123456"}`),
	//}
	//req := &Request{
	//	ID:       ID,
	//	Method:   "tevat.example.logic.Logic/Info",
	//	Msg:      "{}",
	//}
	//req := &Request{
	//	ID:     ID,
	//	Method: "/api/v1/config",
	//	Msg:    "{}",
	//}

	//fmt.Printf("[test]reqMap:%+v \n", reqMap.Export())
	//return client.SendWithRes(reqMap)
	//if name == "watch" {
	//	return proxy.Request{}
	//}
	//v := client.GetResChan()
	//fmt.Printf("[test]resChan id:%d, msg:%s \n", v.ID, v.Msg)
	//return v
}

func TestRuntime_ExportToFunc1(t *testing.T) {
	const SCRIPT = `
    function f(param, cb) {
        cb(+param + 2);
    }

    r(f);
    `

	vm := goja.New()

	var fn goja.Callable

	r := func(call goja.FunctionCall) goja.Value {
		if f, ok := goja.AssertFunction(call.Argument(0)); ok {
			fn = f
		} else {
			panic(vm.NewTypeError("Not a function"))
		}
		return nil
	}

	vm.Set("r", r)
	_, err := vm.RunString(SCRIPT)
	if err != nil {
		t.Fatal(err)
	}

	if res, err := fn(nil, vm.ToValue("40")); err != nil {
		t.Fatal(err)
	} else {
		println("res:", fmt.Sprintf("%+v", res))
		//if !res.StrictEquals(vm.ToValue(42)) {
		//	t.Fatalf("Unexpected value: %v", res)
		//}
	}

}

func TestMockMD(t *testing.T) {
	method := []byte{116, 101, 118, 97, 116, 46, 101, 120, 97, 109, 112, 108, 101, 46, 115, 99, 101, 110, 101, 46, 83, 99, 101, 110, 101, 46, 87, 97, 116, 99, 104, 69, 118, 101, 110, 116, 115, 69, 118, 101, 110, 116, 115, 95, 76, 111, 99, 97, 116, 105, 111, 110}
	msg := []byte{0, 0, 0, 0, 52, 0, 116, 101, 118, 97, 116, 46, 101, 120, 97, 109, 112, 108, 101, 46, 115, 99, 101, 110, 101, 46, 83, 99, 101, 110, 101, 46, 87, 97, 116, 99, 104, 69, 118, 101, 110, 116, 115, 69, 118, 101, 110, 116, 115, 95, 76, 111, 99, 97, 116, 105, 111, 110, 1, 0, 0, 84, 0, 0, 0, 123, 34, 69, 118, 101, 110, 116, 34, 58, 123, 34, 76, 111, 99, 97, 116, 105, 111, 110, 34, 58, 123, 34, 117, 105, 100, 34, 58, 51, 52, 50, 56, 50, 44, 34, 120, 34, 58, 55, 56, 52, 53, 56, 53, 53, 56, 55, 54, 56, 57, 52, 53, 54, 52, 48, 48, 48, 44, 34, 121, 34, 58, 54, 49, 49, 49, 53, 48, 57, 50, 55, 56, 56, 57, 53, 52, 49, 55, 48, 48, 48, 125, 125, 125, 2, 0, 0, 0}
	fmt.Printf("%3X\n", msg)
	fmt.Printf("method:%s \n", method)
	fmt.Printf("msg:%s \n", msg)
	md := &pb.Metadata{
		Metadata: map[string]*pb.Metadata_Value{
			"uid": {
				Values: []string{"1"},
			},
		},
	}
	b, _ := proto.Marshal(md)
	t.Log(b)
}
