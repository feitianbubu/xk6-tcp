package tcp

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"sync/atomic"
	"time"

	"github.com/dop251/goja"
	"go.k6.io/k6/js/common"
	"go.k6.io/k6/js/modules"
	"google.golang.org/protobuf/proto"
	"tevat.nd.org/basecode/goost/async"
	"tevat.nd.org/basecode/goost/encoding/binary"
	"tevat.nd.org/basecode/goost/errors"
	"tevat.nd.org/framework/proxy"
	pb "tevat.nd.org/framework/proxy/proto"
)

type (
	Module struct {
		vu    modules.VU
		conn  net.Conn
		onRec func(res *proxy.Response)
		opts  map[string]interface{}
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

func (m *Module) Connect(addr string) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		m.Throw(fmt.Errorf("conn Connect fail: %s \n", err.Error()))
	}
	m.conn = conn
	return
}

func (m *Module) ConnectOnRec(addr string, onRec func(res *proxy.Response)) {
	m.Connect(addr)
	m.onRec = onRec
}
func (m *Module) StartOnRec(onRec func(res *proxy.Response)) {
	defer func() {
		if r := recover(); r != nil {
			m.Throw(fmt.Errorf("=====on Rec panic=====%+v", r))
		}
	}()
	m.onRec = onRec
	async.GoRaw(func() {
		for {
			res := m.Rec()
			//fmt.Printf("[%v]::for onRec, res.ID:%+v, method:%v, msg:%+v \n", time.Now(), res.ID, m.ToString(res.Method), m.ToString(res.Msg))
			if m.onRec != nil {
				m.onRec(&res)
			}
			//time.Sleep(time.Millisecond * 10)
		}
	})
}

func createMd(m map[string]interface{}) binary.BytesWithUint16Len {
	metadata := make(map[string]*pb.Metadata_Value)
	for k, v := range m {
		metadata[k] = &pb.Metadata_Value{
			Values: []string{fmt.Sprintf("%v", v)},
		}
	}
	md := &pb.Metadata{
		Metadata: metadata,
	}
	b, _ := proto.Marshal(md)
	return b
}

var codec = &proxy.Codec{}

func (m *Module) Send(req *proxy.Request) {
	conn := m.conn
	if conn == nil {
		m.Throw(fmt.Errorf("conn is nil"))
	}
	var err error

	//fmt.Printf("[%v]::req, req.ID:%+v, method:%v, msg:%+v \n", time.Now(), req.ID, m.ToString(req.Method), m.Parse(req.Msg))
	fmt.Printf("+%v", req.ID)
	if req.Method == nil || len(req.Method) == 0 {
		m.Throw(fmt.Errorf("req is invalid, req: %+v \n", req))
	}
	err = codec.Encode(conn, req)
	if err != nil {
		m.Throw(fmt.Errorf("send fail: %s \n", err.Error()))
	}
	return
}
func getReqId(reqJson *goja.Object) (id uint32) {
	if reqJson.Get("id") == nil {
		return
	}
	idExport, ok := reqJson.Get("id").Export().(int64)
	if !ok {
		return
	}
	return uint32(idExport)
}

func (m *Module) Throw(err error) {
	fmt.Printf("[%v]::throw fail, err:%+v", time.Now(), err)
	common.Throw(m.vu.Runtime(), fmt.Errorf("conn is nil"))
}

func (m *Module) Decode(r io.Reader) (proxy.Response, error) {
	var h uint32
	//fmt.Println("read h")
	if err := binary.Read(r, binary.LittleEndian, &h); err != nil {
		return proxy.Response{}, err
	}

	//fmt.Println("read res", h)
	var res proxy.Response
	err := binary.Read(r, binary.LittleEndian, &res)
	//fmt.Printf("decode:%+v, err:%+v \n", res, err)

	return res, err
}

func (m *Module) Rec() proxy.Response {
	defer func() {
		if r := recover(); r != nil {
			m.Throw(fmt.Errorf("=====Rec panic=====%+v", r))
		}
	}()
	conn := m.conn
	if conn == nil {
		m.Throw(fmt.Errorf("conn is nil"))
	}
	//fmt.Printf("[%v]::codec.Decode,start \n", time.Now())
	//v2, err := m.Decode(conn)
	//fmt.Println("[go] m.Decode ", v2, err)
	v, err := m.Decode(conn)
	//fmt.Printf("[%v]:: m.Decode, v:%+v, err:%+v  \n", time.Now(), v, err)
	if err != nil {
		m.Throw(fmt.Errorf("recv fail: %s \n", err.Error()))
	}
	return v
}

func (m *Module) Stringify(obj any) string {
	fmt.Printf("stringify: %+v \n", obj)
	return fmt.Sprintf("%s", obj)
}
func (m *Module) Parse(bytes []byte) map[string]interface{} {
	resMap := make(map[string]interface{})
	err := json.Unmarshal(bytes, &resMap)
	if err != nil {
		fmt.Printf("parse fail: bytes:%+v, msg:%s, err:%+v", bytes, bytes, errors.WithStack(err))
	}
	return resMap
}
func (m *Module) ToString(data any) string {
	return fmt.Sprintf("%s", data)
}

func (m *Module) SendWithRes(reqJson *proxy.Request) proxy.Response {
	m.Send(reqJson)
	return m.Rec()
}

func (m *Module) Close() error {
	conn := m.conn
	if conn == nil {
		m.Throw(fmt.Errorf("conn is nil"))
	}
	return conn.Close()
}

var ID = uint32(0)

func (m *Module) GetReqObject(name string, options ...func(map[string]interface{})) *proxy.Request {
	req, err := GetRequestFromJson(name)
	if err != nil {
		m.Throw(fmt.Errorf("GetRequestFromJson fail, err:%+v", errors.WithStack(err)))
	}
	//reqMap.Set("method", "tevat.example.auth.Auth/login")
	//ID++
	atomic.AddUint32(&ID, 1)
	req.ID = ID
	msg := map[string]interface{}{}
	json.Unmarshal(req.Msg, &msg)
	//switch name {
	//case "login":
	//	msg["account_id"] = fmt.Sprintf("%d", ID)
	//	//msg["account_token"] = "1234561"
	//}
	for _, o := range options {
		o(msg)
	}
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		m.Throw(fmt.Errorf("GetReqObject Marshal fail, err:%+v", errors.WithStack(err)))
	}
	req.Msg = msgBytes
	return req
}

type ApiData struct {
	ID       uint32
	Method   string
	Metadata map[string]interface{}
	Msg      map[string]interface{}
}
type ApiDatas map[string]ApiData

var apiDatas *ApiDatas

func initApiDatas() error {
	if apiDatas != nil {
		return nil
	}
	jsonFile, err := os.Open("config/apiData.json")
	if err != nil {
		fmt.Printf("open file fail:%+v \n", err)
		return err
	}
	defer jsonFile.Close()
	byteValue, _ := io.ReadAll(jsonFile)
	err = json.Unmarshal(byteValue, &apiDatas)
	if err != nil {
		fmt.Printf("Unmarshal file fail:%+v \n", err)
		return err
	}
	fmt.Printf("initApiDatas:%+v \n", apiDatas)
	return nil
}

func GetRequestFromJson(name string) (*proxy.Request, error) {
	//fmt.Printf("apiDatas:%+v \n", apiDatas)
	reqJson := (*apiDatas)[name]
	msg, _ := json.Marshal(reqJson.Msg)
	req := &proxy.Request{
		ID:     reqJson.ID,
		Method: []byte(reqJson.Method),
		Msg:    msg,
	}
	if reqJson.Metadata != nil {
		metadata, _ := json.Marshal(reqJson.Metadata)
		req.Metadata = metadata
	}
	return req, nil
}

func (m *Module) Login(accountId string) (float64, error) {
	req := m.GetReqObject("login", SetMsg("account_id", accountId))
	res := m.SendWithRes(req)

	fmt.Printf("[%v]:login, res.ID:%+v, method:%v, msg:%+v \n", time.Now(), res.ID, m.ToString(res.Method), m.Parse(res.Msg))
	if !res.Result {
		return 0, fmt.Errorf("login fail by:%v", req.Msg)
	}
	msg := m.Parse(res.Msg)
	uid := msg["uid"].(float64)
	return uid, nil
}

type Opts struct {
	AccountId    string
	MoveTimes    int64
	WatchEnabled bool
}

//	func WithMoveTimes(times int64) Opts {
//		return func(m *Module) {
//			m.opts["moveTimes"] = times
//		}
//	}
func (m *Module) Start(addr string, opts Opts) error {
	m.Connect(addr)
	//defer m.Close()
	//m.Connect("127.0.0.1:12345")
	initApiDatas()
	fmt.Printf("start opts:%+v \n", opts)
	uid, err := m.Login(opts.AccountId)
	if err != nil {
		m.Throw(err)
	}
	m.StartOnRec(m.OnRec)
	if opts.WatchEnabled {
		m.Send(m.GetReqObject("event"))
	}
	for i := 0; i < int(opts.MoveTimes); i++ {
		location := map[string]interface{}{}
		location["uid"] = uid
		rs := rand.NewSource(time.Now().UnixNano())
		location["x"] = rand.New(rs).Int()
		location["y"] = rand.New(rs).Int()
		//msg := map[string]interface{}{}
		//msg["location"] = location
		m.Send(m.GetReqObject("move", SetMsg("location", location)))
		rand := time.Duration(rand.New(rs).Intn(60))
		time.Sleep(time.Millisecond * rand)
	}
	m.Send(m.GetReqObject("leave", SetMsg("uid", uid)))
	time.Sleep(time.Millisecond * 1000)
	return nil
}
func SetMsg(key string, value interface{}) func(map[string]interface{}) {
	return func(msg map[string]interface{}) {
		msg[key] = value
	}
}

func (m *Module) OnRec(res *proxy.Response) {
	fmt.Printf("-%v", res.ID)
	m.Parse(res.Msg)
	//fmt.Printf("[%v]::onRec, res.ID:%+v, method:%v, msg:%+v \n", time.Now(), res.ID, m.ToString(res.Method), m.Parse(res.Msg))
}
