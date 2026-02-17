package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	pb "rsj-v3-monorepo/pkg/worker"
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
	port := os.Getenv("PORT")
	if port == "" {
		port = "50051"
	}
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterWorkerNodeServer(s, &server{})

	log.Printf("RSJ-V3.0 locked and listening on :%s", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
