package tcp

import (
	"context"
	"testing"
	"time"

	"tevat.nd.org/basecode/goost/codec/binary"
	"tevat.nd.org/provider/tcp"
)

func TestTLS(t *testing.T) {
	s := tcp.Server{}
	cfg := tcp.DefaultServerCfg
	cfg.TLS.Enable = true
	cfg.TLS.Cert = ""
	s.Config(cfg)
	s.Run(context.Background(), nil)
	// t.Log(os.Getwd())
}

func TestDefault(t *testing.T) {
	simple := binary.NewSimple()
	simple.Register(123, (*int32)(nil))
	// simple.Register(123, (*int)(nil))

	cfg := tcp.DefaultServerCfg
	cfg.Addr = ":9092"
	s := tcp.NewServer()
	err := s.Config(cfg)
	if err != nil {
		t.Fatal(err)
	}
	c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	sh := tcp.NewDefaultServerHandler(simple)
	sh.HandlePeerFunc(func(peer *tcp.Peer) {
		// if peer == nil {
		// 	t.Log("closed")
		// 	return
		// }
		for {
			im, err := peer.Recv()
			if err != nil {
				return
			}
			switch m := im.(type) {
			case *int32:
				if *m != int32(123) {
					t.Error(*m)
					return
				}
				t.Log(*m)
				cancel()
			default:
				t.Errorf("%T", m)
				// case *int:
				// 	fmt.Println(*m)
			}
		}
	})
	go s.Run(c, sh)
	select {}

	ch := tcp.NewDefaultClientHandler(simple)
	ch.HandlePeerFunc(func(peer *tcp.Peer) {
		if peer == nil {
			t.Log("closed")
			return
		}
		err := peer.Send(int32(123))
		// err := peer.Write(123)
		if err != nil {
			t.Error(err)
		}
	})
	tcp.Dial("127.0.0.1:9092", ch)

	//<-c.Done()
}
