package proxy

import (
	"context"
	"fmt"
	"strings"
	"encoding/json"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
	xlog "tevat.nd.org/basecode/goost/log"
	grpcerr "tevat.nd.org/basecode/grpc-error"

	"tevat.nd.org/framework/app"
	"tevat.nd.org/framework/consts"
	"tevat.nd.org/framework/errors"
)

//nolint:gochecknoglobals // global handler collection
var handlerMgr = handlerManager{handlers: make(map[string]handlerTpl)}

type (
	Info struct {
		Method     string
		Metadata   map[string][]string
		RemoteAddr string
		Msg        proto.Message
	}

	notifyOptions struct {
		method    []byte
		methodStr string
	}

	notifyOption interface {
		apply(*notifyOptions)
	}

	notifyMethod string

	client interface {
		header(string) metadata.MD
		setHeader(string, metadata.MD)
		remoteAddr() string
		send(ProtoCodec, []byte)
		notify(ProtoCodec, []byte, ...notifyOption)
		error(ProtoCodec, int, []byte)
		onClose(string, func())
		close()
		reqID(string) int64
	}

	handler interface {
		Serve(context.Context, []byte)
	}

	handlerTpl interface {
		IsStream() bool
		New(client, goGroup) handler
		Marshal(v any) ([]byte,error)
		UnMarshal(b []byte) (any,error)
	}

	handlerManager struct {
		handlers map[string]handlerTpl
	}
)
func GetHandlerMgr() handlerManager {
	return handlerMgr
}
func (gh grpcHandler[Req, Res, Rq, Rs]) UnMarshal(b []byte) (any,error) {
	var res Rs = new(Res)
	err := gh.opt.codec.Unmarshal(b, res)
	return res,err
}
func (gh grpcHandler[Req, Res, Rq, Rs]) Marshal(msgMap any) ([]byte,error) {
	var req Rq = new(Req)
	msgBytes,err := json.Marshal(msgMap)
	if err!=nil{
		return nil, err
	}
	err = json.Unmarshal(msgBytes, req)
	if err!=nil{
		return nil, err
	}
	return gh.opt.codec.Marshal(req)
}

func parseNotifyOption(opts ...notifyOption) notifyOptions {
	opt := notifyOptions{}
	for _, o := range opts {
		o.apply(&opt)
	}

	return opt
}

func (m notifyMethod) apply(o *notifyOptions) {
	o.method = []byte(m)
	o.methodStr = string(m)
}

func convertMethod(m string) string {
	method := []byte(m)
	for idx := len(method) - 1; idx >= 0; idx-- {
		c := method[idx]

		switch {
		case c >= 'A' && c <= 'Z':
			method[idx] = c + 'a' - 'A'
		case c == '/':
			method[idx] = '.'
			if idx == 0 {
				return string(method[1:])
			}
		}
	}

	return string(method)
}

func (hm *handlerManager) Add(method string, tpl handlerTpl) {
	method = convertMethod(method)

	if _, ok := hm.handlers[method]; ok {
		panic("handler already exists")
	}

	hm.handlers[method] = tpl
}

type goGroup interface {
	GoRaw(func())
}

func (hm *handlerManager) Get(m string) (handlerTpl, error) {
	m = convertMethod(m)

	h, ok := hm.handlers[m]
	if !ok {
		return nil, errors.HandlerNotRegister.Err()
	}

	return h, nil
}

func RegisterHandler[Req, Res any, Rq Message[Req], Rs Message[Res]](service, method string, opts ...HandlerOpt) {
	opt := parseOption(opts...)

	fn := fullname(service, method)

	cm := fn
	if opt.method != "" {
		cm = opt.method
	}

	handler := grpcHandler[Req, Res, Rq, Rs]{
		service:  service,
		method:   cm,
		fullname: fn,
		opt:      opt,
	}
	handlerMgr.Add(cm, handler)

	if opt.keepOrigin && cm != fn {
		handlerMgr.Add(fn, handler)
	}
}

type grpcHandler[Req, Res any, Rq Message[Req], Rs Message[Res]] struct {
	service, method, fullname string
	opt                       handlerOpts

	cli client
	gg  goGroup
}

func (gh grpcHandler[Req, Res, Rq, Rs]) New(cli client, gg goGroup) handler {
	h := gh
	h.cli = cli
	h.gg = gg

	return h
}

func (gh grpcHandler[Req, Res, Rq, Rs]) IsStream() bool {
	return gh.opt.stream
}

func (gh grpcHandler[Req, Res, Rq, Rs]) Error(ctx context.Context, req any, err error) {
	l := app.Logger(ctx)
	l.WithFields(
		xlog.KV("method", gh.method),
		xlog.KV("remote_addr", gh.cli.remoteAddr()),
		xlog.KV("req", req),
	).Warnf("%+v", err)

	ei := grpcerr.Reason(err)

	msg, er := gh.opt.codec.Marshal(ei)
	if er != nil {
		return
	}

	gh.cli.error(gh.opt.codec, grpcerr.StatusCode(err), msg)
}

func (gh grpcHandler[Req, Res, Rq, Rs]) Serve(ctx context.Context, data []byte) {
	var req Rq = new(Req)
	if err := gh.opt.codec.Unmarshal(data, req); err != nil {
		return
	}

	gh.perform(ctx, req)
}

func (gh grpcHandler[Req, Res, Rq, Rs]) perform(ctx context.Context, req Rq) {
	md := gh.cli.header(gh.service)
	remoteAddr := gh.cli.remoteAddr()

	if !HandlerPolicy.Allow(ctx, Info{
		Method:     gh.method,
		Metadata:   md,
		RemoteAddr: remoteAddr,
		Msg:        req,
	}) {
		gh.Error(ctx, req, errors.HandlerDeny.Err())

		return
	}

	md.Append(consts.AddrHeader, remoteAddr)

	key, err := gh.opt.shardKeyGetter(md, req)
	if err != nil {
		gh.Error(ctx, req, err)

		return
	}

	gcli, err := resolveGrpcClient(gh.service, withKey(key))
	if err != nil {
		gh.Error(ctx, req, err)

		return
	}

	ctx = metadata.NewOutgoingContext(ctx, md)

	if gh.opt.stream {
		gh.grpcStream(ctx, gcli, req)

		return
	}

	reqID := gh.cli.reqID(gh.service)
	md.Append(consts.ReqIDHeader, fmt.Sprintf("%d-%d-%s", reqID-1, reqID, remoteAddr))
	gh.gg.GoRaw(func() {
		gh.grpcUnary(ctx, gcli, req)
	})
}

func (gh grpcHandler[Req, Res, Rq, Rs]) grpcUnary(ctx context.Context, gcli *grpc.ClientConn, req proto.Message) {
	var res Rs = new(Res)

	ctx, cancel := context.WithTimeout(ctx, gh.opt.timeout)
	defer cancel()

	var header metadata.MD

	err := gcli.Invoke(ctx, gh.fullname, req, res, grpc.Header(&header))
	if err != nil {
		gh.Error(ctx, req, err)

		return
	}

	msg, err := gh.opt.codec.Marshal(res)
	if err != nil {
		gh.Error(ctx, req, err)

		return
	}

	if err := gh.watchEvent(ctx, header); err != nil {
		if !gh.opt.ignoreWatchError {
			gh.Error(ctx, req, err)

			return
		}

		app.Logger(ctx).Warnf("%+v\n", err)
	}

	header.Delete(consts.ContentTypeHeader)

	for k := range header {
		if strings.HasPrefix(k, "_.") {
			header.Delete(k)
		}
	}

	gh.cli.setHeader(gh.service, header)
	gh.cli.send(gh.opt.codec, msg)
}

func (gh grpcHandler[Req, Res, Rq, Rs]) watchEvent(
	ctx context.Context, header metadata.MD,
) error {
	watchHeader := gh.opt.watchHeaderGetter(header)
	if watchHeader.Method == "" {
		return nil
	}

	return watch(ctx, gh.cli.remoteAddr(), watchHeader, gh.opt.codec,
		eventCallback(func(key, method string, msg []byte, at ActionType) {
			if method == "" {
				method = watchHeader.Method
			}
			if !at.Has(ActionNotForward) {
				gh.cli.notify(gh.opt.codec, msg, notifyMethod(method))
			}
			if at.Has(ActionUnwatch) {
				topicManager.Leave(gh.cli.remoteAddr(), key, method)
			}
			if at.Has(ActionClose) {
				app.Logger(ctx).Infof("close connection: %s by %s", gh.cli.remoteAddr(), method)
				gh.cli.close()
			}
		}))
}

func (gh grpcHandler[Req, Res, Rq, Rs]) grpcStream(
	ctx context.Context, gcli *grpc.ClientConn, req proto.Message,
) {
	stream, err := gcli.NewStream(ctx, &grpc.StreamDesc{
		ServerStreams: true,
	}, gh.fullname)
	if err != nil {
		gh.Error(ctx, req, err)

		return
	}

	if err := stream.SendMsg(req); err != nil {
		gh.Error(ctx, req, err)

		return
	}

	if err := stream.CloseSend(); err != nil {
		gh.Error(ctx, req, err)

		return
	}

	gh.gg.GoRaw(func() {
		var res Rs = new(Res)
		for {
			if err := stream.RecvMsg(res); err != nil {
				return
			}
			msg, err := gh.opt.codec.Marshal(res)
			if err != nil {
				continue
			}

			if m := gh.opt.nofityMethod(res); m != "" {
				gh.cli.notify(gh.opt.codec, msg, notifyMethod(m))
			} else {
				gh.cli.notify(gh.opt.codec, msg)
			}
		}
	})
}

type eventCallback func(string, string, []byte, ActionType)

func (e eventCallback) OnEvent(key, method string, event []byte, action ActionType) {
	e(key, method, event, action)
}
