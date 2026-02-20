package main

import (
	"context"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	pb "rsj-worker/proto"
)

type server struct {
	pb.UnimplementedWorkerNodeServer
}

func (s *server) SubmitEvidence(ctx context.Context, in *pb.EvidencePacket) (*pb.SecureResponse, error) {
	start := time.Now()
	log.Printf("[gRPC WORKER] ⚡ Received Hash: %s | Timestamp: %s", in.GetHash(), in.GetTimestamp())
	
	// Simulate processing
	elapsed := time.Since(start).Milliseconds()
	
	return &pb.SecureResponse{
		Status:           "secured",
		ProcessingTimeMs: elapsed,
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	
	s := grpc.NewServer()
	pb.RegisterWorkerNodeServer(s, &server{})
	
	log.Printf("RSJ-V3 gRPC Worker listening on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
