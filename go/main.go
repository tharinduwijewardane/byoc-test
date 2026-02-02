package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"

	pb "github.com/MrSupiri/choreo-byoc-examples/go/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct {
	pb.UnimplementedTestServiceServer
}

func (s *server) CheckActive(ctx context.Context, req *pb.Empty) (*pb.CheckActiveResponse, error) {
	return &pb.CheckActiveResponse{Active: true}, nil
}

func (s *server) HealthCheck(ctx context.Context, req *pb.Empty) (*pb.HealthCheckResponse, error) {
	return &pb.HealthCheckResponse{Healthy: true}, nil
}

func (s *server) Hello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
	return &pb.HelloResponse{Message: fmt.Sprintf("Hello %s", req.GetName())}, nil
}

func (s *server) PrintBody(ctx context.Context, req *pb.PrintBodyRequest) (*pb.PrintBodyResponse, error) {
	body := req.GetBody()
	log.Printf("Received body: %s", body)
	return &pb.PrintBodyResponse{
		Message:    "Body received and printed",
		BodyLength: int32(len(body)),
	}, nil
}

func (s *server) PrintHeaders(ctx context.Context, req *pb.PrintHeadersRequest) (*pb.PrintHeadersResponse, error) {
	headers := req.GetHeaders()
	var headerParts []string
	for name, value := range headers {
		headerParts = append(headerParts, fmt.Sprintf("%s: %s", name, value))
	}
	log.Printf("Request Headers: %s", strings.Join(headerParts, " | "))
	return &pb.PrintHeadersResponse{
		Message:     "Headers received and printed",
		HeaderCount: int32(len(headers)),
	}, nil
}

func (s *server) SimulateInternalError(ctx context.Context, req *pb.Empty) (*pb.Empty, error) {
	return nil, status.Error(codes.Internal, "internal server error")
}

func (s *server) SimulateConflict(ctx context.Context, req *pb.Empty) (*pb.Empty, error) {
	return nil, status.Error(codes.Aborted, "conflict")
}

func (s *server) Proxy(ctx context.Context, req *pb.ProxyRequest) (*pb.ProxyResponse, error) {
	host := req.GetHost()
	args := req.GetArgs()

	if len(host) == 0 {
		host = "http://postman-echo.com"
	}
	if len(args) == 0 {
		args = "get?foo1=bar1&foo2=bar2"
	}

	resp, err := http.Get(fmt.Sprintf("%s/%s", strings.TrimRight(host, "/"), strings.TrimLeft(args, "/")))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "proxy request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to read proxy response: %v", err)
	}

	return &pb.ProxyResponse{Body: string(body)}, nil
}

func loggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	log.Printf("gRPC call: %s", info.FullMethod)
	return handler(ctx, req)
}

func main() {
	port := 9090
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer(grpc.UnaryInterceptor(loggingInterceptor))
	pb.RegisterTestServiceServer(s, &server{})

	fmt.Printf("gRPC server listening on %v\n", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
