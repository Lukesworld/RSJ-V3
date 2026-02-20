package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/pquerna/otp/totp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	pb "rsj-v3-monorepo/pkg/juristic"
	"rsj-v3-monorepo/pkg/workerpool"
)

type server struct {
	pb.UnimplementedSovereignNodeServer
}

func (s *server) CommitState(ctx context.Context, in *pb.CommitStateRequest) (*emptypb.Empty, error) {
	log.Printf("[AGENT] EXECUTED: %s", in.NewState)
	return &emptypb.Empty{}, nil
}

const socketPath = "/data/data/com.termux/files/usr/tmp/juristic_gate.sock"


// ApprovalJob handles the serial execution of TOTP verification
type ApprovalJob struct {
	ResultChan chan error
}

func (j *ApprovalJob) Process(ctx context.Context) error {
	// Clean up any stale socket
	_ = os.Remove(socketPath)

	l, err := net.Listen("unix", socketPath)
	if err != nil {
		j.ResultChan <- fmt.Errorf("socket listen failed: %v", err)
		return err
	}
	defer l.Close()
	
	if err := os.Chmod(socketPath, 0600); err != nil {
		j.ResultChan <- fmt.Errorf("chmod failed: %v", err)
		return err
	}

	// Accept connection with timeout logic could be added here
	// For now, blocking accept as per original design, but in a worker
	conn, err := l.Accept()
	if err != nil {
		j.ResultChan <- fmt.Errorf("accept failed: %v", err)
		return err
	}
	defer conn.Close()

	secret := os.Getenv("RSJV3_SECRET")
	buf := make([]byte, 8)
	conn.SetReadDeadline(time.Now().Add(30 * time.Second)) // Add timeout
	n, err := conn.Read(buf)
	if err != nil {
		j.ResultChan <- fmt.Errorf("read failed: %v", err)
		return err
	}

	code := strings.TrimSpace(string(buf[:n]))
	if !totp.Validate(code, secret) {
		j.ResultChan <- status.Error(codes.Unauthenticated, "INVALID TOTP")
		return nil
	}

	j.ResultChan <- nil // Success
	return nil
}

func (j *ApprovalJob) ID() string {
	return "approval-gate"
}

// Global worker pool for serializing approvals
var approvalPool *workerpool.Pool

func PrincipalGateInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if info.FullMethod == "/rsj.v3.sovereign.SovereignNode/CommitState" {
		log.Printf("[JURISTIC] Queuing approval request...")
		
		// ... existing logic ...
		resultChan := make(chan error, 1)
		// Assuming approvalPool is initialized globally as before
		// We can't easily access the struct from here without refactoring, but the pool is global in the file.
		job := &ApprovalJob{ResultChan: resultChan}
		
		approvalPool.Submit(job)
		
		// ... wait for result ...
		// Simplified for replacement context match:
		select {
		case err := <-resultChan:
			if err != nil {
				return nil, status.Errorf(codes.PermissionDenied, "VETO ACTIVE: %v", err)
			}
		case <-ctx.Done():
			return nil, status.Error(codes.Canceled, "Request canceled")
		}
	}
	return handler(ctx, req)
}

func main() {
	// Initialize a single-threaded pool to serialize access to the socket
	approvalPool = workerpool.NewPool(1, 10)
	approvalPool.Start()
	defer approvalPool.Stop()

	lis, err := net.Listen("tcp", ":50053") // Updated to match agent default
	if err != nil { log.Fatalf("Port blocked: %v", err) }
	
	s := grpc.NewServer(grpc.UnaryInterceptor(PrincipalGateInterceptor))
	pb.RegisterSovereignNodeServer(s, &server{})
	
	log.Println("[RSJ-V3] JURISTIC GATE OPERATIONAL (V4.3)")
	if err := s.Serve(lis); err != nil { log.Fatalf("Server crash: %v", err) }
}
