package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"google.golang.org/grpc"
	pb "rsj-v3-monorepo/pkg/juristic"
)

// AgentState tracks internal state
type AgentState struct {
	AgentID            string
	StartTime          time.Time
	ResourceAllocation float64
	mu                 sync.Mutex
}

func (s *AgentState) UpdateResources() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ResourceAllocation = 10.0 + rand.Float64()*(90.0-10.0)
}

func (s *AgentState) GetUptime() int64 {
	return int64(time.Since(s.StartTime).Seconds())
}

// Server implements pb.SovereignNodeServer
type Server struct {
	pb.UnimplementedSovereignNodeServer
	state *AgentState
}

func (s *Server) DispatchDirective(ctx context.Context, in *pb.Directive) (*pb.DirectiveAck, error) {
	log.Printf("gRPC: DispatchDirective received for %s", in.GetDirectiveId())
	return &pb.DirectiveAck{
		DirectiveId: in.GetDirectiveId(),
		Accepted:    true,
		Reason:      "Executed",
	}, nil
}

func main() {
	agentPort := os.Getenv("AGENT_PORT")
	if agentPort == "" {
		agentPort = "35507"
	}
	juristicPort := os.Getenv("JURISTIC_PORT")
	if juristicPort == "" {
		juristicPort = "50053"
	}

	agentID := fmt.Sprintf("agent-%d", rand.Intn(9000)+1000)
	state := &AgentState{
		AgentID:   agentID,
		StartTime: time.Now(),
	}

	log.Printf("Initializing RSJ-V3 Agent: %s", agentID)

	// Start Server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", agentPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterSovereignNodeServer(grpcServer, &Server{state: state})

	go func() {
		log.Printf("Agent Listener Active on Port %s", agentPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// Connect to Juristic (Control Plane)
	// Using Insecure for now as per Python script
	conn, err := grpc.Dial(fmt.Sprintf("localhost:%s", juristicPort), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect to Juristic: %v", err)
	}
	defer conn.Close()

	client := pb.NewSovereignNodeClient(conn)

	// Heartbeat Loop
	stopChan := make(chan struct{})
	go func() {
		log.Println("State Reporting Loop Started.")
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				state.UpdateResources()
				uptime := state.GetUptime()
				req := &pb.HeartbeatRequest{
					AgentId:       state.AgentID,
					LoadAverage:   float32(state.ResourceAllocation),
					UptimeSeconds: uptime,
				}
				log.Println("Sending Heartbeat...")
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				_, err := client.SendHeartbeat(ctx, req)
				cancel()
				if err != nil {
					log.Printf("Error sending heartbeat: %v", err)
				}
			case <-stopChan:
				return
			}
		}
	}()

	// Graceful Shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down agent...")
	close(stopChan)
	grpcServer.GracefulStop()
}
