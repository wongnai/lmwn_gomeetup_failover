package grpc

import (
	"context"
	"log"
	"net"
	"runtime/debug"

	"lmwn_gomeetup_failover/internal/service"
	pb "lmwn_gomeetup_failover/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCServer struct {
	server   *grpc.Server
	listener net.Listener
}

func NewGRPCServer(svc *service.Service) (*GRPCServer, error) {
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		return nil, err
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(unaryInterceptorRecovery),
	)
	pb.RegisterOrderServiceServer(grpcServer, &OrderService{service: svc})

	return &GRPCServer{
		server:   grpcServer,
		listener: listener,
	}, nil
}

// unaryInterceptorRecovery recovers from panics in gRPC calls
func unaryInterceptorRecovery(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic: %v\nStack Trace: %s", r, debug.Stack())
			err = status.Errorf(codes.Internal, "internal server error")
		}
	}()

	return handler(ctx, req)
}

func (g *GRPCServer) Start() {
	log.Println("Starting gRPC server on port 50051")
	if err := g.server.Serve(g.listener); err != nil {
		log.Fatalf("Failed to start gRPC server: %v", err)
	}
}

func (g *GRPCServer) Stop(ctx context.Context) {
	log.Println("Shutting down gRPC server...")
	done := make(chan struct{})
	go func() {
		g.server.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		log.Println("gRPC server shutdown complete.")
	case <-ctx.Done():
		log.Println("gRPC shutdown timeout exceeded, forcing stop.")
		g.server.Stop()
	}
}
