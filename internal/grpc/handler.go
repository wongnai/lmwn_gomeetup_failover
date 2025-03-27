package grpc

import (
	"context"
	"log"

	"lmwn_gomeetup_failover/internal/service"
	pb "lmwn_gomeetup_failover/proto"
)

type OrderService struct {
	pb.UnimplementedOrderServiceServer
	service *service.Service
}

func (o *OrderService) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	param := req.Param
	log.Printf("gRPC: Creating order %s", param)
	orderID, err := o.service.CreateOrder(param)
	if err != nil {
		return nil, err
	}
	return &pb.CreateOrderResponse{OrderId: orderID}, nil
}
