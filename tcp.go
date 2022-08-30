package tcp

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"runtime/debug"
	"sync"
	"time"

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
		vu       modules.VU
		conn     net.Conn
		onRec    func(res *proxy.Response)
		opts     map[string]interface{}
		apiDataS ApiDataS
	}
	RootModule struct{}
	ApiDataS   struct {
		data map[string]ApiData
		mu   sync.Mutex
	}
	ApiData struct {
		ID       uint32
		Method   string
		Metadata map[string]interface{}
		Msg      map[string]interface{}
	}
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
	modules.Register("k6/x/tcp", new(RootModule))
}

func (m *Module) Connect(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("conn Connect fail: %s \n", err.Error())
	}
	m.conn = conn
	return nil
}

func (m *Module) ConnectOnRec(addr string, onRec func(res *proxy.Response)) error {
	err := m.Connect(addr)
	if err != nil {
		return errors.WithStack(err)
	}
	m.onRec = onRec
	return nil
}
func (m *Module) StartOnRec(onRec func(res *proxy.Response)) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("=====StartOnRec panic=====%+v \n", r)
		}
	}()
	m.onRec = onRec
	async.GoRaw(func() {
		for {
			if m.onRec == nil {
				return
			}
			res, err := m.Rec()
			if err != nil {
				fmt.Printf("[%v]::for onRec fail,err:%+v, res:%+v \n", time.Now(), err, res)
			}
			fmt.Printf("[%v]::for onRec, res.ID:%+v, method:%v, msg:%+v, err:%+v \n", time.Now(), res.ID, m.ToString(res.Method), m.ToString(res.Msg), err)
			if m.onRec != nil && res != nil {
				m.onRec(res)
			} else {
				fmt.Printf("stop rec!! \n")
				return
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

func (m *Module) Send(reqAny any) error {
	var err error
	conn := m.conn
	if conn == nil {
		return fmt.Errorf("conn is nil")
	}

	req := &proxy.Request{}
	switch v := reqAny.(type) {
	case *proxy.Request:
		req = v
	case map[string]interface{}:
		id := v["id"].(int64)
		req.ID = uint32(id)
		method := v["method"].(string)
		req.Method = []byte(method)
		msg := v["msg"].(map[string]interface{})
		msgBytes, err := json.Marshal(msg)
		if err != nil {
			fmt.Printf("send fail by json.Marshal(msg) msg:%+v \n", msg)
		}
		req.Msg = msgBytes
		if v["metadata"] != nil {
			metadata := v["metadata"].(map[string]interface{})
			req.Metadata = createMd(metadata)
		}
	default:
		err = errors.WithStack(fmt.Errorf("send fail by invalid req:%+v", reqAny))
		fmt.Println(err)
		//debug.PrintStack()
		return err
	}

	//fmt.Printf("[%v]::req, req.ID:%+v, method:%v, msg:%+v \n", time.Now(), req.ID, m.ToString(req.Method), m.Parse(req.Msg))
	fmt.Printf("+%v", req.ID)
	if req.Method == nil || len(req.Method) == 0 {
		return errors.WithStack(fmt.Errorf("req is invalid, req: %+v \n", req))
	}
	err = codec.Encode(conn, req)
	if err != nil {
		return fmt.Errorf("send fail: %s \n", err.Error())
	}
	return nil
}

func (m *Module) Decode(r io.Reader) (*proxy.Response, error) {
	var h uint32
	//fmt.Println("read h")
	if err := binary.Read(r, binary.LittleEndian, &h); err != nil {
		return nil, err
	}

	//fmt.Println("read res", h)
	var res proxy.Response
	err := binary.Read(r, binary.LittleEndian, &res)
	//fmt.Printf("decode:%+v, err:%+v \n", res, err)

	return &res, err
}

func (m *Module) Rec() (*proxy.Response, error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("=====Rec panic=====%+v \n", r)
		}
	}()
	conn := m.conn
	if conn == nil {
		return nil, fmt.Errorf("conn is nil")
	}
	//fmt.Printf("[%v]::codec.Decode,start \n", time.Now())
	//v2, err := m.Decode(conn)
	//fmt.Println("[go] m.Decode ", v2, err)
	return m.Decode(conn)
	//fmt.Printf("[%v]:: m.Decode, v:%+v, err:%+v  \n", time.Now(), v, err)
}

func (m *Module) Stringify(obj any) string {
	fmt.Printf("stringify: %+v \n", obj)
	return fmt.Sprintf("%s", obj)
}
func (m *Module) Parse(bytes []byte) map[string]interface{} {
	resMap := make(map[string]interface{})
	err := json.Unmarshal(bytes, &resMap)
	if err != nil {
		fmt.Printf("parse fail: bytes:%+v, msg:%s", bytes, bytes)
		debug.PrintStack()
	}
	return resMap
}
func (m *Module) ToString(data any) string {
	return fmt.Sprintf("%s", data)
}

func (m *Module) SendWithRes(reqJson *proxy.Request) (*proxy.Response, error) {
	err := m.Send(reqJson)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return m.Rec()
}

func (m *Module) Close() {
	//m.onRec = nil
	err := m.conn.Close()
	if err != nil {
		debug.PrintStack()
		fmt.Printf("close fail:%+v", errors.WithStack(err))
	}
}

func (m *Module) GetReqObject(name string, options ...func(map[string]interface{})) *proxy.Request {
	var err error
	req, err := m.GetRequestFromJson(name)
	if err != nil {
		fmt.Printf("GetRequestFromJson fail, err:%+v \n", errors.WithStack(err))
		return nil
	}
	msg := map[string]interface{}{}
	err = json.Unmarshal(req.Msg, &msg)
	if err != nil {
		fmt.Printf("GetReqObject Unmarshal fail, err:%+v \n", errors.WithStack(err))
		return nil
	}
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
		fmt.Printf("GetReqObject Marshal fail, err:%+v \n", errors.WithStack(err))
		return nil
	}
	req.Msg = msgBytes
	return req
}

func (m *Module) Init() error {
	if m.apiDataS.data != nil {
		return nil
	}
	m.apiDataS.mu.Lock()
	defer m.apiDataS.mu.Unlock()
	jsonFile, err := os.Open("config/apiData.json")
	if err != nil {
		fmt.Printf("open file fail:%+v \n", err)
		return err
	}
	defer func(jsonFile *os.File) {
		err := jsonFile.Close()
		if err != nil {
			return
		}
	}(jsonFile)
	byteValue, _ := io.ReadAll(jsonFile)
	err = json.Unmarshal(byteValue, &m.apiDataS.data)
	if err != nil {
		fmt.Printf("Unmarshal file fail:%+v \n", err)
		return err
	}
	fmt.Printf("\ninitApiDatas:%+v \n", m.apiDataS.data)
	return nil
}

func (m *Module) GetRequestFromJson(name string) (*proxy.Request, error) {
	//fmt.Printf("apiDataS:%+v \n", apiDataS)
	reqJson := m.apiDataS.data[name]
	msg, _ := json.Marshal(reqJson.Msg)
	req := &proxy.Request{
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
	res, err := m.SendWithRes(req)
	if err != nil {
		return 0, errors.WithStack(err)
	}

	fmt.Printf("[%v]:login, res.ID:%+v, method:%v, msg:%+v \n", time.Now(), res.ID, m.ToString(res.Method), m.Parse(res.Msg))
	if !res.Result {
		return 0, fmt.Errorf("login fail by:%s", req.Msg)
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
	var err error
	err = m.Connect(addr)
	if err != nil {
		return errors.WithStack(err)
	}
	//defer m.Close()
	//m.Connect("127.0.0.1:12345")
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
		//randSleep := time.Duration(rand.New(rs).Intn(1000))
		//time.Sleep(time.Millisecond * randSleep)
	}
	err = m.Send(m.GetReqObject("leave", SetMsg("uid", uid)))
	time.Sleep(time.Millisecond * 1000)
	return err
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
