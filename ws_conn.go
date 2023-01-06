package xk6_tcp

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type WSConn struct {
	conn *websocket.Conn

	keyBuf1   []byte
	keyBuf2   []byte
	nSendPos1 int
	nSendPos2 int
	nRecvPos1 int
	nRecvPos2 int
	isEncrypt bool
	sync.RWMutex
	bFaceClient bool // 是否面向客户端，false为面向游服
}

func (wsConn *WSConn) SetReadDeadline(timeOut time.Duration) error {
	return wsConn.conn.SetReadDeadline(time.Now().Add(timeOut))
}
func (wsConn *WSConn) SetWriteDeadline(timeOut time.Duration) error {
	return wsConn.conn.SetWriteDeadline(time.Now().Add(timeOut))
}

func (wsConn *WSConn) buildKeyClient(key string) {
	var key1 int
	var key2 int
	_, err := fmt.Fscanf(strings.NewReader(key), "%08X%08X",
		&key1, &key2)
	if err != nil {
		fmt.Println("NewWSConn", true, err)
	}
	wsConn.keyBuf1 = make([]byte, 256)
	wsConn.keyBuf2 = make([]byte, 256)
	var (
		aa = 0x7E
		bb = 0x33
		cc = 0xA1
	)
	a1 := ((key1 >> 0) & 0xFF) ^ aa
	b1 := ((key1 >> 8) & 0xFF) ^ bb
	c1 := ((key1 >> 24) & 0xFF) ^ cc
	fst1 := (key1 >> 16) & 0xFF
	a2 := ((key2 >> 0) & 0xFF) ^ aa
	b2 := ((key2 >> 8) & 0xFF) ^ bb
	c2 := ((key2 >> 24) & 0xFF) ^ cc
	fst2 := (key2 >> 16) & 0xFF
	nCode := fst1
	for i := 0; i < 256; i++ {
		wsConn.keyBuf1[i] = byte(nCode)
		nTemp := (a1 * nCode) % 256
		nCode = (nTemp*nCode + b1*nCode + c1) % 256
	}
	if nCode != fst1 {
		//fmt.Println("err", "err:nCode!=fst1")
	}
	nCode = fst2
	for i := 0; i < 256; i++ {
		wsConn.keyBuf2[i] = byte(nCode)
		nCode = (a2*nCode*nCode + b2*nCode + c2) % 256
	}
	if nCode != fst2 {
		//fmt.Println("err", "err:nCode!=fst2")
	}
	//fmt.Printf("a1:%02X,b1:%02X,c1:%02X,first1:%02X\n", a1, b1, c1, fst1)
	//fmt.Printf("a2:%02X,b2:%02X,c2:%02X,first2:%02X\n", a2, b2, c2, fst2)
	return
}

func (wsConn *WSConn) buildKeyServer(key string) {
	var key1 int
	var key2 int
	_, err := fmt.Fscanf(strings.NewReader(key), "%08X%08X",
		&key1, &key2)
	if err != nil {
		fmt.Println("NewWSConn", true, err)
	}
	wsConn.keyBuf1 = make([]byte, 256)
	wsConn.keyBuf2 = make([]byte, 256)
	var (
		aa = 0x7E
		bb = 0x33
		cc = 0xA1
	)
	a1 := ((key1 >> 0) & 0xFF) ^ aa
	b1 := ((key1 >> 8) & 0xFF) ^ bb
	c1 := ((key1 >> 24) & 0xFF) ^ cc
	fst1 := (key1 >> 16) & 0xFF
	a2 := ((key2 >> 0) & 0xFF) ^ aa
	b2 := ((key2 >> 8) & 0xFF) ^ bb
	c2 := ((key2 >> 24) & 0xFF) ^ cc
	fst2 := (key2 >> 16) & 0xFF
	nCode := fst1
	for i := 0; i < 256; i++ {
		wsConn.keyBuf1[i] = byte(nCode)
		nTemp := (a1 * nCode) % 256
		nCode = (nTemp*nCode + b1*nCode + c1) % 256
	}
	//if nCode != fst1 {
	//	zaplog.Warn("err:nCode!=fst1")
	//}
	nCode = fst2
	for i := 0; i < 256; i++ {
		wsConn.keyBuf2[i] = byte(nCode)
		nCode = (a2*nCode*nCode + b2*nCode + c2) % 256
	}
	//if nCode != fst2 {
	//	zaplog.Warn("err:nCode!=fst2")
	//}
	//fmt.Printf("a1:%02X,b1:%02X,c1:%02X,first1:%02X\n", a1, b1, c1, fst1)
	//fmt.Printf("a2:%02X,b2:%02X,c2:%02X,first2:%02X\n", a2, b2, c2, fst2)
	return
}
func (wsConn *WSConn) DecryptRecvCode(bufMsg []byte, nLen int) []byte {
	for i := 0; i < nLen; i++ {
		bufMsg[i] ^= wsConn.keyBuf1[wsConn.nRecvPos1]
		bufMsg[i] ^= wsConn.keyBuf2[wsConn.nRecvPos2]
		wsConn.nRecvPos1++
		if (wsConn.nRecvPos1) >= 256 {
			wsConn.nRecvPos1 = 0
			wsConn.nRecvPos2++
			if wsConn.nRecvPos2 >= 256 {
				wsConn.nRecvPos2 = 0
			}
		}
	}

	return bufMsg
}
func (wsConn *WSConn) EncryptSendCode(bufMsg []byte, nLen int) []byte {
	for i := 0; i < nLen; i++ {
		bufMsg[i] ^= wsConn.keyBuf1[wsConn.nSendPos1]
		bufMsg[i] ^= wsConn.keyBuf2[wsConn.nSendPos2]
		wsConn.nSendPos1++
		if (wsConn.nSendPos1) >= 256 {
			wsConn.nSendPos1 = 0
			wsConn.nSendPos2++
			if wsConn.nSendPos2 >= 256 {
				wsConn.nSendPos2 = 0
			}
		}
	}

	return bufMsg
}

func NewWSConn(conn *websocket.Conn, isEncode, bFaceClient bool, key string) *WSConn {
	if key == "" {
		key = "ADF482EADB98654F"
	}
	wsConn := &WSConn{
		conn:        conn,
		keyBuf1:     nil,
		keyBuf2:     nil,
		nSendPos1:   0,
		nSendPos2:   0,
		nRecvPos1:   0,
		nRecvPos2:   0,
		isEncrypt:   isEncode,
		bFaceClient: bFaceClient,
	}

	if bFaceClient {
		wsConn.buildKeyClient(key)
	} else {
		wsConn.buildKeyServer(key)
	}
	return wsConn
}

func (wsConn *WSConn) Close() error {
	return wsConn.conn.Close()
}

func (wsConn *WSConn) LocalAddr() net.Addr {
	return wsConn.conn.LocalAddr()
}

func (wsConn *WSConn) RemoteAddr() net.Addr {
	return wsConn.conn.RemoteAddr()
}

func (wsConn *WSConn) ReadMessage() (messageType int, p []byte, err error) {
	messageType, p, err = wsConn.conn.ReadMessage()
	if err != nil {
		return 0, nil, err
	}
	if wsConn.isEncrypt {
		p = wsConn.DecryptRecvCode(p, len(p))
	}
	return
}
func (wsConn *WSConn) WriteMessage(messageType int, data []byte) error {
	if !wsConn.bFaceClient {
		wsConn.Lock()
		defer wsConn.Unlock()
	}
	if wsConn.isEncrypt {
		data = wsConn.EncryptSendCode(data, len(data))
	}
	_ = wsConn.SetWriteDeadline(time.Second * 3)
	return wsConn.conn.WriteMessage(messageType, data)
}
