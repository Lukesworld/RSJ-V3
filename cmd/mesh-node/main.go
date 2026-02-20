package main

import (
	"context"
	"crypto/tls"
	"flag"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	pb "rsj-v3-monorepo/pkg/worker"
)

var (
	target = flag.String("target", "rsj-worker-fm2qsd2x4a-ts.a.run.app", "The target Cloud Run endpoint")
	port   = flag.String("port", "443", "The target port")
)

func main() {
	flag.Parse()

	addr := *target + ":" + *port
	log.Println("[EDGE] 📡 Connecting to Vault:", addr)

	// Secure SSL Channel (Cloud Run requires this)
	// Use InsecureSkipVerify: true ONLY for debugging if needed, but Cloud Run usually has valid certs.
	// We'll stick to valid certs but default system roots.
	creds := credentials.NewTLS(&tls.Config{})
	
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewWorkerNodeClient(conn)

	hashID := "TERMUX-INIT-001"
	data := "RSJ-V3 Uplink Established"

	packet := &pb.EvidencePacket{
		Hash:      hashID,
		RawData:   data,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	log.Printf("[EDGE] ⚡ Firing Packet: %s...", hashID)
	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	r, err := c.SubmitEvidence(ctx, packet)
	if err != nil {
		log.Fatalf("\n[FAIL] ❌ RPC Error: %v", err)
	}

	latency := time.Since(start).Milliseconds()

	log.Printf("[RECV] ✅ Status: %s", r.GetStatus())
	log.Printf("       📝 Message: %s", r.GetMessage())
	log.Printf("       ☁️  Cloud Time: %dms", r.GetProcessingTimeMs())
	log.Printf("       📶 Latency: %dms", latency)
}
