package xk6_tcp

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	sdkclientpb "tevat.nd.org/toolchain/xk6/pkg/proto/sdk_client"

	"go.k6.io/k6/js/modules"
	"google.golang.org/protobuf/proto"
	"tevat.nd.org/basecode/goost/async"
	"tevat.nd.org/basecode/goost/encoding/binary"
	"tevat.nd.org/basecode/goost/errors"
	pb "tevat.nd.org/framework/proxy/proto"

	"tevat.nd.org/toolchain/xk6/internal/proxy"
)

type ProtoType int32

const (
	ProtoTypePROTOBUF ProtoType = 1
	//ProtoType_JSON     ProtoType = 2
)

type (
	Module struct {
		vu       modules.VU
		conn     net.Conn
		onRec    func(res Res)
		resChan  chan Res
		opts     map[string]interface{}
		apiDataS ApiDataS
	}
	RootModule struct{}
	ApiDataS   struct {
		data   map[string]ApiData
		config Config
		mu     sync.Mutex
	}
	Config struct {
		Proto           ProtoType
		LocalLogin      bool
		ProtoFileConfig ProtoFileConfig
	}
	ProtoFileConfig struct {
		Path  string
		Files []string
	}
	ApiData struct {
		ID       uint32
		Method   string
		Metadata map[string]interface{}
		Msg      map[string]interface{}
	}
	Res struct {
		ID       uint32
		Result   bool
		Method   string
		Msg      map[string]interface{}
		Metadata map[string]interface{}
	}
)

func (m *Module) GetConfig() Config {
	return m.apiDataS.config
}
func (m *Module) GetProto() ProtoType {
	return m.GetConfig().Proto
}
func (m *Module) GetLocalLogin() bool {
	return m.GetConfig().LocalLogin
}

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
	m.resChan = make(chan Res)
	//m.StartOnRec()
	m.Init()
	return nil
}

func (m *Module) ConnectOnRec(addr string, onRec func(res Res)) error {
	err := m.Connect(addr)
	if err != nil {
		return errors.WithStack(err)
	}
	m.setOnRec(onRec)
	return nil
}
func (m *Module) setOnRec(onRec func(res Res)) {
	m.onRec = onRec
}
func (m *Module) StartOnRec() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("=====StartOnRec panic=====%+v \n", r)
		}
	}()
	async.GoRaw(func() {
		for {
			res, err := m.Rec()
			if err != nil {
				if errors.As(err, io.EOF) {
					fmt.Printf("read EOF \n")
					return
				}
				fmt.Printf("[%v]::for onRec fail,err:%+v, res:%+v \n", time.Now(), err, res)
				time.Sleep(time.Second)
			}
			fmt.Printf("-%v", res.ID)
			//fmt.Printf("[%v]::for onRec,res:%+v \n", time.Now(), res)
			if res.ID == 0 {
				// 通知走回调
				if m.onRec != nil {
					m.onRec(res)
				}
			} else {
				select {
				case m.resChan <- res:
				default:
					//fmt.Println("ignor res to resChan:", res)
				}
			}
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

var codec = proxy.NewCodec()

func (m *Module) Send(reqAny any) (*proxy.Request, error) {
	var err error
	conn := m.conn
	if conn == nil {
		return nil, fmt.Errorf("conn is nil")
	}
	req := &proxy.Request{}
	switch v := reqAny.(type) {
	case *proxy.Request:
		req = v
	case map[string]interface{}:
		id, ok := v["id"].(int64)
		if ok {
			req.ID = uint32(id)
		}
		method := v["method"].(string)
		req.Method = []byte(method)
		msg := v["msg"].(map[string]interface{})
		msgBytes, err := m.MarshalMsg(method, msg)
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
		return nil, err
	}

	//fmt.Printf("[%v]::req, req.ID:%+v, method:%v, msg:%+v \n", time.Now(), req.ID, m.ToString(req.Method), m.Parse(req.Msg))
	//fmt.Printf("+%v", req.ID)
	if req.Method == nil || len(req.Method) == 0 {
		return nil, errors.WithStack(fmt.Errorf("req is invalid, req: %+v \n", req))
	}
	return req, codec.Encode(conn, req)
}
func (m *Module) SendAndRec(reqAny any) (Res, error) {
	var res Res
	req, err := m.Send(reqAny)
	if err != nil {
		return res, err
	}

	method := m.ToString(req.Method)
	if strings.HasSuffix(strings.ToLower(method), "/login") {
		return m.Rec()
	}
	for {
		select {
		case res := <-m.resChan:
			if res.ID != req.ID {
				continue
			}
			//fmt.Printf("sendWithRes: %+v \n", res)
			return res, nil
		case <-time.After(time.Second * 3):
			return res, fmt.Errorf("sendWithRes timeout, req:%+v \n", req)
		}
	}
}

func (m *Module) Decode(r io.Reader) (Res, error) {
	var res Res
	var h uint32
	//fmt.Println("read h")
	if err := binary.Read(r, binary.LittleEndian, &h); err != nil {
		return res, err
	}

	//fmt.Println("read res", h)
	resBase := proxy.Response{}
	err := binary.Read(r, binary.LittleEndian, &resBase)
	//method := fmt.Sprintf("%s", resBase.Method)
	//fmt.Println("decode:, err:", resBase, err, method)
	//if method == "tevat.example.logic.Logic.WatchEventsEvents_PropsEvent" {
	//	msg := &logic.PropsEvent{}
	//	err = proto.Unmarshal(resBase.Msg, msg)
	//	fmt.Println("PropsEvent:", err, msg, msg.Props, msg.ProtoReflect(), msg.String())
	//	for i, v := range msg.GetProps() {
	//		fmt.Println("GetProps:", i, v)
	//	}
	//}
	//if method == "tevat.example.logic.Logic.WatchEventsEvents_MoneyEvent" {
	//	msg := &logic.MoneyEvent{}
	//	err = proto.Unmarshal(resBase.Msg, msg)
	//	fmt.Println("MoneyEvent:", err, msg, msg.Count, msg.MoneyType, msg.ProtoReflect(), msg.String())
	//}
	res = m.ParseRes(resBase)
	return res, err
}

func (m *Module) Rec() (Res, error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("=====Rec panic=====%+v \n", r)
		}
	}()
	var res Res
	conn := m.conn
	if conn == nil {
		return res, fmt.Errorf("conn is nil")
	}
	return m.Decode(conn)
	//fmt.Printf("[%v]:: m.Decode, v:%+v, err:%+v  \n", time.Now(), v, err)
}

func parse(bytes []byte) map[string]interface{} {
	resMap := make(map[string]interface{})
	if len(bytes) == 0 {
		return resMap
	}
	err := json.Unmarshal(bytes, &resMap)
	if err != nil {
		fmt.Printf("parse fail: bytes:%+v, msg:%s, err:%+v", bytes, bytes, errors.WithStack(err))
	}
	return resMap
}
func (m *Module) Parse(bytes []byte) map[string]interface{} {
	return parse(bytes)
}
func (m *Module) ToString(data any) string {
	return fmt.Sprintf("%s", data)
}

func (m *Module) Close() {
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
	//fmt.Println("==============", req.Msg, msg, err, name)
	if err != nil {
		fmt.Printf("GetReqObject Unmarshal fail, err:%+v \n", errors.WithStack(err))
		return nil
	}
	for _, o := range options {
		o(msg)
	}

	method := m.ToString(req.Method)
	msgBytes, err := m.MarshalMsg(method, msg)
	if err != nil {
		fmt.Printf("GetReqObject Marshal fail, method:%+v, msg:%+v, err:%+v \n", method, msg, errors.WithStack(err))
		return nil
	}
	req.Msg = msgBytes
	return req
}

func initFile(d any, filename string) {
	jsonFile, err := os.Open(filename)
	if err != nil {
		fmt.Printf("open file fail:%+v \n", err)
		panic(err)
	}
	defer func(jsonFile *os.File) {
		err := jsonFile.Close()
		if err != nil {
			return
		}
	}(jsonFile)
	byteValue, _ := io.ReadAll(jsonFile)
	err = json.Unmarshal(byteValue, d)
	if err != nil {
		fmt.Printf("Unmarshal file fail:%+v \n", err)
		panic(err)
	}
}
func (m *Module) Init() {
	m.apiDataS.mu.Lock()
	defer m.apiDataS.mu.Unlock()
	if m.apiDataS.data != nil {
		return
	}

	initFile(&m.apiDataS.data, "config/apiData.json")
	initFile(&m.apiDataS.config, "config/config.json")
	fmt.Println(":::m.apiDataS", &m.apiDataS)
}

func (m *Module) GetApiDataByName(name string) ApiData {
	apiData := m.apiDataS.data[name]
	//if apiData.Metadata == nil {
	//	apiData.Metadata = make(map[string]interface{})
	//}
	if apiData.Msg == nil {
		apiData.Msg = make(map[string]interface{})
	}
	return apiData
}
func (m *Module) GetRequestFromJson(name string) (*proxy.Request, error) {
	//fmt.Printf("apiDataS:%+v \n", apiDataS)
	reqJson := m.GetApiDataByName(name)
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

func (m *Module) Login(accountId string) (map[string]interface{}, error) {
	opts := make([]func(map[string]interface{}), 0)
	if m.GetLocalLogin() {
		opts = append(opts, SetMsg("account_id", fmt.Sprintf("%v", accountId)))
	} else {
		var tokenRes *sdkclientpb.MsgLoginResp
		getAuthReq := &sdkclientpb.MsgLoginReq{
			PlatformID: 12345,
			AccountID:  accountId, //"testclearluo",
			RegionID:   30001,
			ServerID:   3000101,
			SessionID:  "DkVmYu6eywXJaatEj0nKJuKMo7m7a53C",
			ClientType: 0,
		}
		tokenRes, err := GetToken(getAuthReq)
		if err != nil {
			fmt.Println("getToken fail:", getAuthReq, tokenRes, err)
			return nil, err
		}
		opts = append(opts, SetMsg("account_id", fmt.Sprintf("%v", tokenRes.UserID)))
		opts = append(opts, SetMsg("account_token", tokenRes.Token))
	}
	req := m.GetReqObject("login", opts...)
	if req.ID == 0 {
		req.ID = 1
	}
	res, err := m.SendAndRec(req)
	if err != nil {
		fmt.Printf("login fail by SendWithRes, req:%+v, res:%+v, err:%+v \n", req, res, err)
		return nil, errors.WithStack(err)
	}

	fmt.Printf("[%v]:login, res.ID:%+v, method:%v, msg:%+v, msgType:%T \n", time.Now(), res.ID, res.Method, res.Msg, res.Msg)
	if !res.Result {
		return nil, fmt.Errorf("login fail by: req.msg:%s, res:%+v", req.Msg, res)
	}
	return res.Msg, nil
}

type Opts struct {
	AccountId string
	MoveTimes int64
}

func SetMsg(key string, value interface{}) func(map[string]interface{}) {
	return func(msg map[string]interface{}) {
		msg[key] = value
	}
}

func (m *Module) OnRec(res Res) {
	fmt.Printf("#%v", res.ID)
	//m.Parse(res.Msg)
	//fmt.Printf("[%v]::onRec, res:%+v \n", time.Now(), res)
}
