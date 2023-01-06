package xk6_tcp

import (
	"fmt"

	sdkclientpb "tevat.nd.org/toolchain/xk6/pkg/proto/sdk_client"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

func RecvLoginRes(conn *WSConn) (*sdkclientpb.MsgLoginResp, error) {
	var err error
	_, p, _ := conn.ReadMessage()
	//fmt.Printf("ReadMessage, messageType:%+v, p:%+v(%s), err:%+v \n", messageType, p, p, err)
	msg := &sdkclientpb.Message{}
	err = proto.Unmarshal(p, msg)
	switch msg.MsgID {
	case sdkclientpb.MsgType_MsgID_MsgHeartBeat:
		//data := &sdkclientpb.MsgHeartBeat{}
		//err = proto.Unmarshal(msg.Data, data)
		//fmt.Printf("msg:%+v, data:%+v \n", msg, data)
		return RecvLoginRes(conn)
	case sdkclientpb.MsgType_MsgID_MsgLoginResp:
		data := &sdkclientpb.MsgLoginResp{}
		err = proto.Unmarshal(msg.Data, data)
		//dataJson, _ := json.MarshalIndent(data, "", "\t")
		//fmt.Printf("msg:%+v, data:%s \n", msg, dataJson)
		return data, err
	}
	return nil, fmt.Errorf("no found login res")

}

func GetToken(req *sdkclientpb.MsgLoginReq) (*sdkclientpb.MsgLoginResp, error) {
	urlStr := "ws://192.168.91.5:9011/ws"
	var err error
	c, _, err := websocket.DefaultDialer.Dial(urlStr, nil)
	if err != nil {
		fmt.Errorf("dial fail:%+v", err)
	}
	defer func(c *websocket.Conn) {
		_ = c.Close()
	}(c)
	conn := NewWSConn(c, true, true, "")
	data, err := proto.Marshal(req)
	if err != nil {
		fmt.Errorf("proto.Marshal fail:%+v", err)
	}
	message := &sdkclientpb.Message{
		MsgID: sdkclientpb.MsgType_MsgID_MsgLoginReq,
		Data:  data,
	}
	messageByte, err := proto.Marshal(message)
	if err != nil {
		fmt.Errorf("proto.Marshal fail:%+v", err)
	}

	err = conn.WriteMessage(websocket.BinaryMessage, messageByte)
	if err != nil {
		fmt.Errorf("WriteMessage fail:%+v \n", err)
	}
	//fmt.Printf("WriteMessage, data:%+v(%s), err:%+v \n", data, data, err)
	return RecvLoginRes(conn)
}
