package tcp

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dop251/goja"
	"google.golang.org/protobuf/proto"
	"tevat.nd.org/basecode/goost/async"
	"tevat.nd.org/basecode/goost/errors"
	"tevat.nd.org/framework/proxy"
	pb "tevat.nd.org/framework/proxy/proto"
)

//func OnRec(res *proxy.Request) {
//	fmt.Printf("[test]OnRec:%+v \n", res)
//}

var vm = goja.New()
var ID = uint32(0)

func TestProxyTCP(t *testing.T) {

	//client.Connect("172.24.140.131:12345")
	addr := "172.24.140.131:12345"
	//addr = "127.0.0.1:12345"
	opts := Opts{}
	opts.MoveTimes = int64(2)
	opts.AccountId = "2"
	opts.WatchEnabled = true
	//err := start(addr, opts)
	//if err != nil {
	//	fmt.Printf("start fail:%+v", err)
	//}

	gg := async.NewGoGroup()
	f := func(opts Opts) func(context.Context) {
		return func(_ context.Context) {
			fmt.Println("start:", addr, opts)
			err := start(addr, opts)
			if err != nil {
				fmt.Printf("start fail:%+v", err)
			}
		}
	}
	for i := 1; i < 10; i++ {
		opts.AccountId = strconv.Itoa(i)
		gg.Go(f(opts))
	}
	gg.Wait()
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
func start(addr string, opts Opts) error {
	var m = &Module{}
	var err error
	err = m.Connect(addr)
	if err != nil {
		return errors.WithStack(err)
	}
	err = m.Init()
	if err != nil {
		return errors.WithStack(err)
	}
	fmt.Printf("start opts:%+v \n", opts)
	uid, err := m.Login(opts.AccountId)
	if err != nil {
		return errors.WithStack(err)
	}
	m.StartOnRec(m.OnRec)
	if opts.WatchEnabled {
		err := m.Send(m.GetReqObject("event"))
		if err != nil {
			return errors.WithStack(err)
		}
	}
	for i := 0; i < int(opts.MoveTimes); i++ {
		location := map[string]interface{}{}
		location["uid"] = uid
		rs := rand.NewSource(time.Now().UnixNano())
		location["x"] = rand.New(rs).Intn(100)
		location["y"] = rand.New(rs).Intn(100)
		//msg := map[string]interface{}{}
		//msg["location"] = location
		err = m.Send(m.GetReqObject("move", SetMsg("location", location)))
		//randSleep := time.Duration(rand.New(rs).Intn(6000))
		//time.Sleep(time.Millisecond * randSleep)
	}
	req := m.GetReqObject("leave", SetMsg("uid", uid))
	//atomic.AddUint32(&ID, 1)
	//req.ID = ID
	err = m.Send(req)
	time.Sleep(time.Millisecond * 1000)
	m.Close()
	return err
}

func sTestRuntime_ExportToFunc1(t *testing.T) {
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

func sTestMockMD(t *testing.T) {
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
func sTestGojaValue(t *testing.T) {
	const JS = `
	function f(param){
		return Math.floor(Math.random()*param*(new Date()))%param
	}
	`
	vm := goja.New()
	var err error
	_, err = vm.RunString(JS)
	if err != nil {
		fmt.Println("vm.RunString(JS) fail", err, JS)
	}
	var fn func(i string) string
	err = vm.ExportTo(vm.Get("f"), &fn)
	if err != nil {
		fmt.Println("vm.ExportTo fail", err, &fn)
	}
	fmt.Println(fn("40"))
	fmt.Println(fn("40"))
	fmt.Println(fn("40"))

	f, ok := goja.AssertFunction(vm.Get("f"))
	if !ok {
		fmt.Println("goja.AssertFunction fail", err, &f)
	}
	res, err := f(goja.Undefined(), vm.ToValue(100))
	if err != nil {
		fmt.Println("f(goja.Undefined(), vm.ToValue(100)) fail", err, &res)
	}
	fmt.Println(res)
}

func TestAtomic(t *testing.T) {
	t.Parallel()
	req := &proxy.Request{}
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		atomic.AddUint32(&ID, 1)
		req.ID = ID
		go func(r proxy.Request) {
			fmt.Println("req:", r.ID)
			wg.Done()
		}(*req)
	}
	wg.Wait()
}

func TestAtomic2(t *testing.T) {
	t.Parallel()
	req := &proxy.Request{}
	gg := async.NewGoGroup()
	for i := 1; i < 50; i++ {
		gg.Go(func() func(context.Context) {
			return func(_ context.Context) {
				atomic.AddUint32(&req.ID, 1)
			}
		}())
	}
	gg.Wait()
	fmt.Println("req2:", req.ID)
}
