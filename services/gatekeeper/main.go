package main

import (
	"context"
	"log"
	"net"
	"os"
	"strings"

	"github.com/pquerna/otp/totp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	pb "rsj-v3/juristic/juristic"
)

type server struct {
	pb.UnimplementedJuristicServiceServer
}

func (s *server) CommitState(ctx context.Context, in *pb.StateRequest) (*pb.StateResponse, error) {
	log.Printf("[AGENT] EXECUTED: %s", in.Action)
	return &pb.StateResponse{Status: "SUCCESS", Message: "Authorized by Principal"}, nil
}

const socketPath = "/data/data/com.termux/files/usr/tmp/juristic_gate.sock"

func PrincipalGateInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if info.FullMethod == "/juristic.v1.JuristicService/CommitState" {
		log.Printf("[JURISTIC] AWAITING VETO LIFT...")
		if err := waitForPrincipalApproval(); err != nil {
			return nil, status.Errorf(codes.PermissionDenied, "VETO ACTIVE: %v", err)
		}
	}
	return handler(ctx, req)
}

func waitForPrincipalApproval() error {
	_ = os.Remove(socketPath)
	l, err := net.Listen("unix", socketPath)
	if err != nil { return err }
	defer l.Close()
	_ = os.Chmod(socketPath, 0600)
	conn, err := l.Accept()
	if err != nil { return err }
	defer conn.Close()
	secret := os.Getenv("RSJV3_SECRET")
	buf := make([]byte, 8)
	n, _ := conn.Read(buf)
	if !totp.Validate(strings.TrimSpace(string(buf[:n])), secret) {
		return status.Error(codes.Unauthenticated, "INVALID TOTP")
	}
	return nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil { log.Fatalf("Port blocked: %v", err) }
	
	s := grpc.NewServer(grpc.UnaryInterceptor(PrincipalGateInterceptor))
	pb.RegisterJuristicServiceServer(s, &server{})
	
	log.Println("[RSJ-V3] JURISTIC GATE OPERATIONAL (V4.3)")
	if err := s.Serve(lis); err != nil { log.Fatalf("Server crash: %v", err) }
}
