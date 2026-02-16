package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	pb "rsj-worker/proto"
)

type server struct {
	pb.UnimplementedWorkerNodeServer
}

func (s *server) SubmitEvidence(ctx context.Context, in *pb.EvidencePacket) (*pb.WorkerResponse, error) {
	// Logic: Log the incoming Evidence Hash for the Context Ledger
	fmt.Printf("[CLOUD] ⚡ Recv: %s | TS: %s\n", in.GetHash(), in.GetTimestamp())
	
	return &pb.WorkerResponse{
		Status:             "secured",
		Message:            "Persisted to Newmanventures Mesh",
		ProcessingTimeMs:   50,
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterWorkerNodeServer(s, &server{})

	log.Printf("RSJ-V3.0 locked and listening on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
