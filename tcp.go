package tcp

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"

	"go.k6.io/k6/js/modules"
	"tevat.nd.org/framework/proxy"
	xtcp "tevat.nd.org/provider/tcp"
)

func init() {
	modules.Register("k6/x/tcp", new(TCP))
}

// TCP is the k6 extension for a TCP client.
type TCP struct{}

// Client is the TCP client wrapper.
//type Client struct {
//	conn *net.Conn
//}

//func (r *TCP) XClient(ctxPtr *context.Context) interface{} {
//	rt := common.GetRuntime(*ctxPtr)
//	return common.Bind(rt, &Client{}, ctxPtr)
//}

func (c *TCP) Connect(addr string) (net.Conn, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (c *TCP) Write(conn net.Conn, data []byte) error {
	_, err := conn.Write(data)
	if err != nil {
		return err
	}
	return nil
}

type Codec struct{}

func (Codec) Encode(w io.Writer, data interface{}) error {
	l := binary.Size(data) + 4
	err := binary.Write(w, binary.LittleEndian, uint32(l))
	if err != nil {
		return err
	}
	return binary.Write(w, binary.LittleEndian, data)
}
func (Codec) Decode(r io.Reader) (interface{}, error) {
	var h uint32
	err := binary.Read(r, binary.LittleEndian, &h)
	if err != nil {
		return nil, err
	}
	var req proxy.Response
	err = binary.Read(r, binary.LittleEndian, &req)
	return req, err
}

var ID = uint32(0)

func (c *TCP) Send(method []byte, msg []byte) error {
	handler := xtcp.NewDefaultClientHandler(&Codec{})
	err := xtcp.Dial("127.0.0.1:12345", handler)
	if err != nil {
		return err
	}
	handler.HandlePeerFunc(func(peer *xtcp.Peer) {
		defer peer.Close()
		ID++
		peer.Send(&proxy.Request{
			ID:       ID,
			Method:   method,
			Metadata: nil,
			Msg:      msg,
		})

		v, err := peer.Recv()
		if err != nil {
			fmt.Printf("perr recv fail: %s", err.Error())
		}
		fmt.Printf("recv:%+v", v)

		//for data := range peer.DataChan() {
		//	resp, ok := data.(proxy.Response)
		//	if !ok {
		//		return
		//	}
		//	fmt.Printf("%+v, msg:%s", resp, resp.Msg)
		//}
	})
	return nil
}
