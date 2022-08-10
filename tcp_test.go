package tcp

import (
	"fmt"
	"testing"

	"tevat.nd.org/basecode/goost/codec/binary"
	xtcp "tevat.nd.org/provider/tcp"
)

func TestProxyTCP(t *testing.T) {
	//	client := &TCP{}
	//	client.Send([]byte("/api/v1/login"), []byte(`{
	//  "userAddress": "0x279A4C36098c4e76182706511AB0346518ad6049",
	//  "content": "1659516267",
	//  "signature": "0xd5f2199e375be586e78e454595047edfe6688989a674eca1dc847658aa387e041ceed79ca17be21bd6c6541822ea87e1bf1f3a73cbdf3c0d76bc176a523584b600"
	//}`))
	simple := binary.NewSimple()
	simple.Register(123, (*int32)(nil))

	handler := xtcp.NewDefaultClientHandler(simple)
	err := xtcp.Dial("127.0.0.1:12345", handler)
	if err != nil {
		return
	}
	handler.HandlePeerFunc(func(peer *xtcp.Peer) {
		defer peer.Close()
		ID++
		peer.Send(int32(123))

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
	select {}
}
