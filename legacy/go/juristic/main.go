package main
import (
	"context"
	"database/sql"
	"log"
	"net"
	"os"
	"strings"

	"github.com/pquerna/otp/totp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	_ "github.com/go-sql-driver/mysql"
	pb "rsj-v3/juristic/proto"
)

const socketPath = "/data/data/com.termux/files/usr/tmp/juristic_gate.sock"

// server implements pb.SovereignNodeServer
type server struct {
	pb.UnimplementedSovereignNodeServer
	db *sql.DB
}

func (s *server) CommitState(ctx context.Context, in *pb.CommitStateRequest) (*emptypb.Empty, error) {
	log.Printf("[JURISTIC] CommitState Executed by %s to state %s.", in.PrincipalId, in.NewState)
	return &emptypb.Empty{}, nil
}

func (s *server) SubmitEvidence(ctx context.Context, in *pb.EvidencePacket) (*pb.WorkerResponse, error) {
	log.Printf("[JURISTIC] ⚡ Recv Evidence: %s", in.Hash)

	_, err := s.db.Exec("INSERT INTO evidence (hash, raw_data) VALUES (?, ?)", in.Hash, in.RawData)
	if err != nil {
		log.Printf("[JURISTIC] ❌ Database Error: %v", err)
		return nil, status.Errorf(codes.Internal, "Database failure: %v", err)
	}

	return &pb.WorkerResponse{
		Status:           "secured",
		Message:          "Persisted to Local Sovereign Vault",
		ProcessingTimeMs: 10,
	}, nil
}

func PrincipalGateInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Updated to match the actual proto definition
	if info.FullMethod == "/rsj.v3.sovereign.SovereignNode/CommitState" {
		log.Printf("[JURISTIC] CRITICAL AUTHORIZATION REQUIRED: %s", info.FullMethod)
		if err := waitForPrincipalApproval(); err != nil {
			return nil, status.Errorf(codes.PermissionDenied, "VETO ACTIVE: %v", err)
		}
	}
	return handler(ctx, req)
}

func waitForPrincipalApproval() error {
	_ = os.Remove(socketPath)
	l, err := net.Listen("unix", socketPath)
	if err != nil { log.Fatalf("FATAL: %v", err) }
	defer l.Close()
	_ = os.Chmod(socketPath, 0600)
	
	// Accept connection with timeout logic could be added here, but sticking to simple blocking for now
	conn, err := l.Accept()
	if err != nil { return err }
	defer conn.Close()

	secret := os.Getenv("RSJV3_SECRET")
	if secret == "" { log.Fatalf("FATAL: RSJV3_SECRET NOT SET") }

	buf := make([]byte, 8)
	n, _ := conn.Read(buf)
	if !totp.Validate(strings.TrimSpace(string(buf[:n])), secret) {
		return status.Error(codes.Unauthenticated, "INVALID TOTP")
	}
	return nil
}

func main() {
	db, err := sql.Open("mysql", "root@tcp(127.0.0.1:3306)/rsj_v3")
	if err != nil {
		log.Fatalf("FATAL: Failed to connect to MariaDB: %v", err)
	}
	defer db.Close()

	s := grpc.NewServer(grpc.UnaryInterceptor(PrincipalGateInterceptor))
	
	// Register the service
	pb.RegisterSovereignNodeServer(s, &server{db: db})

	lis, err := net.Listen("tcp", ":50053")
	if err != nil {
		log.Fatalf("FATAL: Failed to listen on port 50053: %v", err)
	}
	log.Println("[RSJ-V3] JURISTIC GATE OPERATIONAL (V4.1)")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("FATAL: Failed to serve: %v", err)
	}
}
