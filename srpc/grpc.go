package sprc

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"net"
	"time"
)

type GrpcServer struct {
	listener net.Listener
	gServer  *grpc.Server
	register []func(g *grpc.Server)
	ops      []grpc.ServerOption
}

func NewGrpcServer(addr string, ops ...GrpcOption) (*GrpcServer, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	s := &GrpcServer{listener: listener}
	for _, v := range ops {
		v.Apply(s)
	}
	s.gServer = grpc.NewServer(s.ops...)
	return s, nil
}

func (s *GrpcServer) Register(f func(s *grpc.Server)) {
	s.register = append(s.register, f)
}

func (s *GrpcServer) Run() error {
	for _, f := range s.register {
		f(s.gServer)
	}
	return s.gServer.Serve(s.listener)
}

func (s *GrpcServer) Stop() {
	s.gServer.Stop()
}

type GrpcOption interface {
	Apply(s *GrpcServer)
}

type DefaultGrpcOption struct {
	f func(s *GrpcServer)
}

func (dgo DefaultGrpcOption) Apply(s *GrpcServer) {
	dgo.f(s)
}

func DefaultWithGrpcOption(ops ...grpc.ServerOption) GrpcOption {
	return &DefaultGrpcOption{
		f: func(s *GrpcServer) {
			s.ops = append(s.ops, ops...)
		},
	}
}

type GrpcClient struct {
	Conn *grpc.ClientConn
}

func NewGrpcClient(conf *GrpcClientConfig) (*GrpcClient, error) {
	ctx := context.Background()
	dialOptions := conf.dialOptions
	if conf.Block {
		// 阻塞直至连接
		if conf.DialTimeout > time.Duration(0) {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, conf.DialTimeout)
			defer cancel()
		}
		dialOptions = append(dialOptions, grpc.WithBlock())
	}
	if conf.keepAlive != nil {
		dialOptions = append(dialOptions, grpc.WithKeepaliveParams(*conf.keepAlive))
	}
	conn, err := grpc.DialContext(ctx, conf.Address, dialOptions...)
	if err != nil {
		return nil, err
	}
	return &GrpcClient{Conn: conn}, nil
}

type GrpcClientConfig struct {
	Address     string
	Block       bool
	DialTimeout time.Duration
	ReadTimeout time.Duration
	Direct      bool
	keepAlive   *keepalive.ClientParameters
	dialOptions []grpc.DialOption
}

func DefaultGrpcClientConfig(addr string) *GrpcClientConfig {
	return &GrpcClientConfig{
		Address:     addr,
		Block:       true,
		DialTimeout: time.Second * 3,
		ReadTimeout: time.Second * 2,
		dialOptions: []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		},
	}
}
