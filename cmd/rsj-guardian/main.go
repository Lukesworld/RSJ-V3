package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"rsj-v3-monorepo/pkg/auth"
	"strings"
	"time"
)

const (
	MasterScript    = "/sdcard/rsj_v3_master.sh"
	CheckInterval   = 300 * time.Second
	CredentialsFile = "credentials.json"
)

func checkIntegrity() {
	cmd := exec.Command("settings", "get", "global", "window_animation_scale")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("[RSJ-V3] Error checking animation scale: %v", err)
	}

	scale := strings.TrimSpace(string(output))
	if scale != "0" {
		fmt.Println("[RSJ-V3] Integrity Breach Detected. Re-hardening...")
		if err := exec.Command("sh", MasterScript).Run(); err != nil {
			log.Printf("[RSJ-V3] Failed to execute master script: %v", err)
		}
	} else {
		fmt.Println("[RSJ-V3] System Integrity: SECURE")
	}
}

func main() {
	fmt.Println("[RSJ-V3] Guardian Module Started.")

	// Integrated Authentication (Optional if file exists)
	if _, err := os.Stat(CredentialsFile); err == nil {
		fmt.Println("[RSJ-V3] Credentials found. Initializing Secure Services...")
		ctx := context.Background()
		_, err := auth.InitializeServices(ctx, CredentialsFile)
		if err != nil {
			log.Printf("[RSJ-V3] Auth Initialization Warning: %v", err)
		} else {
			fmt.Println("[RSJ-V3] Secure Services Initialized.")
		}
	} else {
		fmt.Println("[RSJ-V3] No credentials.json found. Continuing in local-only mode.")
	}

	for {
		checkIntegrity()
		time.Sleep(CheckInterval)
	}
}
