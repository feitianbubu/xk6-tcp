package xk6_tcp

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dop251/goja"
	"go.k6.io/k6/lib/netext/grpcext"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
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
	addr = "127.0.0.1:12345"
	opts := Opts{}
	opts.MoveTimes = int64(100)
	opts.AccountId = "131402"
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
	for i := 1; i < 2; i++ {
		opts.AccountId = "131402" //strconv.Itoa(i)
		gg.Go(f(opts))
	}
	gg.Wait()
	//time.Sleep(time.Second * 3)
}

//var ID uint32 = 0

//	func rec() *proxy.Request {
//		return client.Rec()
//	}
func sendReq(m *Module, name string) {
	req := m.GetReqObject(name)
	atomic.AddUint32(&ID, 1)
	req.ID = ID
	_, err := m.Send(req)
	if err != nil {
		panic(errors.WithStack(err))
	}
}

func start(addr string, opts Opts) error {
	var ID = uint32(0)
	var m = &Module{}
	var err error
	err = m.ConnectOnRec(addr, m.onRec)
	if err != nil {
		return errors.WithStack(err)
	}
	m.Init()
	//m.SetLocalLogin(true)
	fmt.Printf("start opts:%+v \n", opts)
	_, err = m.Login(opts.AccountId)
	if err != nil {
		return errors.WithStack(err)
	}
	sendReq(m, "airDrop")
	sendReq(m, "enter")
	m.StartOnRec()
	time.Sleep(time.Second)
	sendReq(m, "enterScene")
	time.Sleep(time.Second * 2)
	//sendReq(m, "unReg")
	//sendReq(m, "serverInfo")
	//sendReq(m, "userAction")
	//time.Sleep(time.Second * 2)

	//req3 := m.GetReqObject("battleRankUpdate")
	//atomic.AddUint32(&ID, 1)
	//req3.ID = ID
	//m.Send(req3)
	for i := 0; i < int(opts.MoveTimes); i++ {
		location := map[string]interface{}{}
		rs := rand.NewSource(time.Now().UnixNano())
		location["x"] = rand.New(rs).Intn(100)
		location["y"] = rand.New(rs).Intn(100)
		location["z"] = rand.New(rs).Intn(100)
		req := m.GetReqObject("move", SetMsg("location", location))
		atomic.AddUint32(&ID, 1)
		req.ID = ID

		//fmt.Println("req:::::::", req)
		_, err = m.Send(req)

		randSleep := time.Duration(rand.New(rs).Intn(100))
		time.Sleep(time.Millisecond * randSleep)
	}
	//time.Sleep(time.Millisecond * 3000)
	//m.Close()
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

func TestMockMD(t *testing.T) {
	method := []byte{116, 101, 118, 97, 116, 46, 101, 120, 97, 109, 112, 108, 101, 46, 115, 99, 101, 110, 101, 46, 83, 99, 101, 110, 101, 46, 87, 97, 116, 99, 104, 69, 118, 101, 110, 116, 115, 69, 118, 101, 110, 116, 115, 95, 76, 111, 99, 97, 116, 105, 111, 110}
	msg := []byte{110, 117, 108, 108}
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

func TestGrpc(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()

	notls := grpc.WithTransportCredentials(insecure.NewCredentials())

	addr := "grpcbin.test.k6.io:9000"
	addr = "127.0.0.1:9001"
	conn, err := grpcext.Dial(ctx, addr, notls)
	if err != nil {
		fmt.Printf("%+v", err)
	}
	defer conn.Close()

	fdset, err := conn.Reflect(ctx)
	if err != nil {
		fmt.Printf("dialing failed: %+v", err)
	}

	files, err := protodesc.NewFiles(fdset)
	if err != nil {
		fmt.Printf("dialing failed: %+v", err)
	}
	fmt.Println("files", files)
}

func TestProto(t *testing.T) {
	var waitToPlayers []uint64
	players := make([]uint64, 0)
	for i := 0; i < 10; i++ {
		waitToPlayers = append(waitToPlayers, uint64(i))
	}
	count := 0
	for i := 0; i < len(waitToPlayers); {
		if count < 2 {
			players = append(players, waitToPlayers[i])
			waitToPlayers = append(waitToPlayers[:i], waitToPlayers[i+1:]...)
			fmt.Println("", i, count, waitToPlayers, waitToPlayers[:i], waitToPlayers[i+1:])
			count++
		} else {
			//usr.Data.WaitToPlayer = xtime.Now().Unix() + 3
			break
		}
	}
	fmt.Println("=======", waitToPlayers, players)
}
