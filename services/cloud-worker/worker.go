package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// EvidencePacket matches the JSON structure from your Postman collection
type EvidencePacket struct {
	Hash      string    `json:"hash"`
	RawData   string    `json:"rawData"`
	Timestamp time.Time `json:"timestamp"`
}

func processPacket(p EvidencePacket) {
	// Simulate heavy decentralized processing
	fmt.Printf("[WORKER] 📥 Processing Packet: %s\n", p.Hash)
	time.Sleep(2 * time.Second) 
	fmt.Printf("[WORKER] ✅ Hash %s committed to vault at %s\n", p.Hash, time.Now().Format(time.RFC3339))
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var packet EvidencePacket
	if err := json.NewDecoder(r.Body).Decode(&packet); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Hand off to a goroutine so the API remains responsive
	go processPacket(packet)

	w.WriteHeader(http.StatusAccepted)
	fmt.Fprint(w, `{"status": "processing", "message": "Worker has queued the evidence"}`)
}

func main() {
	port := ":8081"
	http.HandleFunc("/worker/process", handler)
	log.Printf("RSJ-V3 Worker active on %s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
