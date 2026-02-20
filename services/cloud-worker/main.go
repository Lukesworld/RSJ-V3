package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cloud.google.com/go/vertexai/genai"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	pb "rsj-v3-monorepo/pkg/worker"
	"rsj-v3-monorepo/pkg/workerpool"
)

// Config holds runtime configuration
type Config struct {
	GRPCPort      string
	HTTPPort      string
	WorkerCount   int
	JobQueueSize  int
	ProjectID     string
	GeminiModel   string
}

// AIJob encapsulates an AI analysis request
type AIJob struct {
	Evidence    string
	ResultChan  chan string
	ErrorChan   chan error
	AIClient    *genai.Client
	ModelName   string
}

func (j *AIJob) Process(ctx context.Context) error {
	if j.AIClient == nil {
		j.ResultChan <- "AI Analysis Unavailable (No Client)"
		return nil
	}

	model := j.AIClient.GenerativeModel(j.ModelName)
	resp, err := model.GenerateContent(ctx, genai.Text(fmt.Sprintf("Analyze this legal evidence for urgency and relevance: %s", j.Evidence)))
	if err != nil {
		j.ErrorChan <- err
		return err
	}

	analysis := "No insight generated"
	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		if txt, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
			analysis = string(txt)
		}
	}
	j.ResultChan <- analysis
	return nil
}

func (j *AIJob) ID() string {
	return "ai-job" // simplified ID
}

// WorkerServer implements the gRPC service and HTTP handler
type WorkerServer struct {
	pb.UnimplementedWorkerNodeServer
	pool     *workerpool.Pool
	aiClient *genai.Client
	config   Config
}

// SubmitEvidence implements the gRPC handler
func (s *WorkerServer) SubmitEvidence(ctx context.Context, in *pb.EvidencePacket) (*pb.WorkerResponse, error) {
	start := time.Now()
	
	// Create job for AI analysis
	resultChan := make(chan string, 1)
	errorChan := make(chan error, 1)
	
	job := &AIJob{
		Evidence:   in.GetRawData(),
		ResultChan: resultChan,
		ErrorChan:  errorChan,
		AIClient:   s.aiClient,
		ModelName:  s.config.GeminiModel,
	}

	// Submit to pool (blocking if full, providing backpressure)
	s.pool.Submit(job)

	// Wait for result or context cancellation
	var analysis string
	select {
	case res := <-resultChan:
		analysis = res
	case err := <-errorChan:
		log.Printf("AI Error: %v", err)
		analysis = "AI Error"
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(10 * time.Second): // Timeout
		analysis = "AI Timeout"
	}

	duration := time.Since(start).Milliseconds()
	log.Printf("[GRPC] Processed hash %s in %dms", in.GetHash(), duration)

	return &pb.WorkerResponse{
		Status:           "secured",
		Message:          "Persisted. Insight: " + analysis,
		ProcessingTimeMs: int32(duration),
	}, nil
}

// HTTPHandler handles JSON requests
func (s *WorkerServer) HTTPHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var packet struct {
		Hash      string `json:"hash"`
		RawData   string `json:"rawData"`
	}
	if err := json.NewDecoder(r.Body).Decode(&packet); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Similar logic to gRPC, reuse the job
	resultChan := make(chan string, 1)
	errorChan := make(chan error, 1)
	
	job := &AIJob{
		Evidence:   packet.RawData,
		ResultChan: resultChan,
		ErrorChan:  errorChan,
		AIClient:   s.aiClient,
		ModelName:  s.config.GeminiModel,
	}

	s.pool.Submit(job)

	// For HTTP, let's wait for result too (or fire-and-forget if desired)
	var analysis string
	select {
	case res := <-resultChan:
		analysis = res
	case err := <-errorChan:
		log.Printf("AI Error: %v", err)
		analysis = "Error"
	case <-time.After(10 * time.Second):
		analysis = "Timeout"
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "processed",
		"analysis": analysis,
	})
}

func main() {
	config := Config{
		GRPCPort:     ":50051",
		HTTPPort:     ":8081",
		WorkerCount:  5,
		JobQueueSize: 100,
		ProjectID:    os.Getenv("GOOGLE_CLOUD_PROJECT"),
		GeminiModel:  "gemini-pro",
	}

	// Initialize Worker Pool
	pool := workerpool.NewPool(config.WorkerCount, config.JobQueueSize)
	pool.Start()
	defer pool.Stop()

	// Initialize Vertex AI
	ctx := context.Background()
	var aiClient *genai.Client
	var err error
	if config.ProjectID != "" {
		aiClient, err = genai.NewClient(ctx, config.ProjectID, "us-central1")
		if err != nil {
			log.Printf("Warning: Failed to create AI client: %v", err)
		} else {
			defer aiClient.Close()
			log.Println("Vertex AI Client Initialized")
		}
	}

	server := &WorkerServer{
		pool:     pool,
		aiClient: aiClient,
		config:   config,
	}

	// Start gRPC Server
	lis, err := net.Listen("tcp", config.GRPCPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterWorkerNodeServer(grpcServer, server)
	reflection.Register(grpcServer)

	go func() {
		log.Printf("gRPC server listening on %s", config.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve gRPC: %v", err)
		}
	}()

	// Start HTTP Server
	mux := http.NewServeMux()
	mux.HandleFunc("/worker/process", server.HTTPHandler)
	httpServer := &http.Server{
		Addr:    config.HTTPPort,
		Handler: mux,
	}

	go func() {
		log.Printf("HTTP server listening on %s", config.HTTPPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Graceful Shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down servers...")
	grpcServer.GracefulStop()
	httpServer.Shutdown(context.Background())
	log.Println("Servers stopped.")
}
