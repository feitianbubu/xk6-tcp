package xk6_tcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	_ "unsafe"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"

	"tevat.nd.org/toolchain/xk6/internal/proxy"
)

var methodMap = make(map[string]protoreflect.MethodDescriptor)

var loadMethodMapOnce sync.Once // 惰性初始化mesh
func (m *Module) GetMd(method string) protoreflect.MethodDescriptor {
	loadMethodMapOnce.Do(func() {
		m.initMethodMap()
	})
	return methodMap[method]
}

var initMapLock sync.Mutex

func (m *Module) initMethodMap() {
	files := m.GetConfig().ProtoFileConfig.Files
	for _, v := range files {
		m.GetMdFromFile(v)
	}
}

func (m *Module) GetMdFromFile(filename string) error {
	initMapLock.Lock()
	defer initMapLock.Unlock()
	path := m.GetConfig().ProtoFileConfig.Path
	registry, err := createProtoRegistry(path, filename)
	if err != nil {
		return err
	}

	//var md protoreflect.MethodDescriptor
	registry.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		sds := fd.Services()
		for i := 0; i < sds.Len(); i++ {
			sd := sds.Get(i)
			sdname := sd.FullName()

			mds := sd.Methods()
			for j := 0; j < mds.Len(); j++ {
				mdi := mds.Get(j)
				methodName := fmt.Sprintf("/%s/%s", sdname, mdi.Name())
				//fmt.Println("::methodName", methodName)
				methodMap[methodName] = mdi
			}
		}
		return true
	})
	return nil
}

func createProtoRegistry(srcDir string, filename string) (*protoregistry.Files, error) {
	// Create descriptors using the protoc binary.
	// Imported dependencies are included so that the descriptors are self-contained.
	fns := strings.Split(filename, "/")
	tmpFile := fns[len(fns)-1] + "-tmp.pb"
	cmd := exec.Command("protoc",
		"--include_imports",
		"--descriptor_set_out="+tmpFile,
		"-I",
		srcDir,
		path.Join(srcDir, filename))
	//cmd = exec.Command("protoc", "--include_imports --descriptor_set_out=auth.proto-tmp.pb -I proto auth.proto")

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Println("cmd err:", err)
		return nil, err
	}
	defer func(name string) {
		_ = os.Remove(name)
	}(tmpFile)

	marshalledDescriptorSet, err := os.ReadFile(tmpFile)
	if err != nil {
		return nil, err
	}
	descriptorSet := descriptorpb.FileDescriptorSet{}
	err = proto.Unmarshal(marshalledDescriptorSet, &descriptorSet)
	if err != nil {
		return nil, err
	}

	files, err := protodesc.NewFiles(&descriptorSet)
	if err != nil {
		return nil, err
	}

	return files, nil
}

func (m *Module) ParseRes(res proxy.Response) Res {
	r := Res{}
	r.ID = res.ID
	r.Result = res.Result
	r.Method = fmt.Sprintf("%s", res.Method)
	if r.ID == 0 {
		return r
	}
	if m.GetProto() == ProtoTypePROTOBUF {
		if !r.Result {
			r.Msg = map[string]interface{}{"msg": res.Msg}
		} else {
			md := m.GetMd(r.Method)
			resp := dynamicpb.NewMessage(md.Output())
			if err := proto.Unmarshal(res.Msg, resp); err != nil {
				r.Msg = map[string]interface{}{"msg": res.Msg}
			} else {
				msg := make(map[string]interface{})
				marshaler := protojson.MarshalOptions{EmitUnpopulated: true}
				raw, _ := marshaler.Marshal(resp)
				_ = json.Unmarshal(raw, &msg)
				r.Msg = msg
			}
		}
	} else {
		r.Msg = parse(res.Msg)
		r.Metadata = parse(res.Metadata)
	}
	return r
}

func (m *Module) MarshalMsg(method string, msgMap map[string]interface{}) ([]byte, error) {
	if m.GetProto() == ProtoTypePROTOBUF {
		md := m.GetMd(method)
		req := dynamicpb.NewMessage(md.Input())

		msgBytes, err := json.Marshal(msgMap)
		if err != nil {
			return nil, err
		}

		if err := protojson.Unmarshal(msgBytes, req); err != nil {
			return nil, fmt.Errorf("unable to serialise request object to protocol buffer: %w", err)
		}
		b, err := proto.Marshal(req)
		if err != nil {
			return nil, err
		}
		//fmt.Println("reqmarshal", req, b)
		return b, nil
	}
	return json.Marshal(msgMap)
}
