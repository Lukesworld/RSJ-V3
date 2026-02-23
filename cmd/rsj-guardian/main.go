package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

// The Security Layer for the Admin Console
func BasicAuth(handler http.HandlerFunc, user, pass string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		if !ok || u != user || p != pass {
			w.Header().Set("WWW-Authenticate", `Basic realm="RSJ-V3 Restricted Area"`)
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized Access. Intrusion Attempt Logged.\n"))
			return
		}
		handler(w, r)
	}
}

func main() {
	// ==========================================
	// 1. PUBLIC ROUTE (What the world sees)
	// ==========================================
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `
		<!DOCTYPE html>
		<html lang="en">
		<head>
			<meta charset="UTF-8">
			<meta name="viewport" content="width=device-width, initial-scale=1.0">
			<title>Newman Ventures | Public Edge Node</title>
			<style>
				body { background: #050505; color: #ccc; font-family: 'Courier New', monospace; margin: 0; padding: 0; }
				header { border-bottom: 2px solid #ff8c00; padding: 30px 20px; text-align: center; background: #0a0a0a; }
				h1 { color: #ff8c00; letter-spacing: 5px; margin: 0; text-transform: uppercase; font-size: 2.5em; }
				.subtitle { color: #00ff00; font-size: 1em; margin-top: 10px; font-weight: bold; }
				.container { max-width: 800px; margin: 40px auto; padding: 0 20px; }
				.card { background: #111; border: 1px solid #222; padding: 25px; margin-bottom: 25px; border-radius: 4px; border-left: 4px solid #ff8c00; }
				h2 { color: #fff; margin-top: 0; font-size: 1.5em; text-transform: uppercase; }
				p { line-height: 1.6; font-size: 1.1em; }
				.service-list { list-style: none; padding: 0; }
				.service-list li { margin-bottom: 10px; }
				.service-list li::before { content: "> "; color: #ff8c00; font-weight: bold; }
				a.btn { display: inline-block; background: transparent; color: #ff8c00; border: 1px solid #ff8c00; padding: 12px 25px; text-decoration: none; text-transform: uppercase; font-weight: bold; transition: 0.3s; letter-spacing: 2px; }
				a.btn:hover { background: #ff8c00; color: #000; }
				footer { text-align: center; margin-top: 50px; padding: 30px; border-top: 1px solid #222; font-size: 0.8em; color: #555; }
			</style>
		</head>
		<body>
			<header>
				<h1>Newman Ventures</h1>
				<div class="subtitle">[ RSJ-V3 PUBLIC NODE: ONLINE ]</div>
			</header>
			
			<div class="container">
				<div class="card">
					<h2>Welcome to the Sovereign Edge</h2>
					<p>You have accessed a publicly routed node operating on the Recursive Sovereign Juristic Architecture (RSJ-V3). This server runs entirely on independent hardware, bypassing centralized Big Tech infrastructure.</p>
				</div>
				
				<div class="card">
					<h2>Available Services</h2>
					<ul class="service-list">
						<li><b>Secure Data Vaulting:</b> Encrypted, hardware-level file storage.</li>
						<li><b>Autonomous Agents:</b> 24/7 background processing and data synthesis.</li>
						<li><b>Sovereign Web Hosting:</b> Immutable web presence delivered via encrypted tunnels.</li>
					</ul>
				</div>

				<div style="text-align: center; margin-top: 50px; padding: 30px; border: 1px dashed #333;">
					<p style="color: #666; font-size: 0.9em; text-transform: uppercase; margin-bottom: 20px;">Authorized Personnel Only</p>
					<a href="/admin" class="btn">Enter Admin Console</a>
				</div>
			</div>

			<footer>
				&copy; 2026 Newman Ventures | Data Center: Perth, WA | Connection: SECURE HTTP/2
			</footer>
		</body>
		</html>
		`)
	})

	// ==========================================
	// 2. PRIVATE ADMIN ROUTE (The Secure Vault)
	// ==========================================
	adminHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		baseDir := "/data/data/com.termux/files/home/projects/web/rsj-v3-monorepo"
		
		// File Reader Logic
		readFile := r.URL.Query().Get("file")
		if readFile != "" {
			content, err := os.ReadFile(filepath.Join(baseDir, readFile))
			if err != nil {
				fmt.Fprintf(w, "Error reading file.<br><a href='/admin' style='color:#ff8c00;'>Back</a>")
				return
			}
			fmt.Fprintf(w, `
			<!DOCTYPE html><html><head><meta charset="UTF-8"><title>Reading %s</title></head>
			<body style="background:#0a0a0a; color:#0f0; font-family:monospace; padding:20px;">
				<h2>📄 %s</h2>
				<a href="/admin" style="color:#ff8c00; text-decoration:none; border:1px solid #ff8c00; padding:5px;">[ RETURN TO VAULT ]</a>
				<hr style="border-color:#333; margin:20px 0;">
				<pre style="background:#111; padding:15px; border-left:3px solid #ff8c00; overflow-x:auto;">%s</pre>
			</body></html>`, readFile, readFile, string(content))
			return
		}
		
		// System Diagnostic Logic
		cmdRun := r.URL.Query().Get("cmd")
		cmdOutput := ""
		if cmdRun == "uptime" {
			out, _ := exec.Command("uptime").CombinedOutput()
			cmdOutput = fmt.Sprintf("<div style='border:1px dashed #0f0; padding:15px; margin-bottom:20px; background:#002200;'><b style='color:#ff8c00;'>⚡ TERMINAL OUTPUT (uptime):</b><br>%s</div>", string(out))
		}

		// List Files Logic
		entries, _ := os.ReadDir(baseDir)
		fileListHTML := "<ul>"
		for _, e := range entries {
			if e.IsDir() {
				fileListHTML += fmt.Sprintf("<li>📁 %s</li>", e.Name())
			} else {
				fileListHTML += fmt.Sprintf("<li>📄 <a href='?file=%s' style='color:#00ff00; text-decoration:none;'>%s</a></li>", e.Name(), e.Name())
			}
		}
		fileListHTML += "</ul>"

		// Render the Admin Page
		fmt.Fprintf(w, `
		<!DOCTYPE html>
		<html>
		<head>
			<meta charset="UTF-8">
			<meta name="viewport" content="width=device-width, initial-scale=1.0">
			<title>Admin Suite | RSJ-V3</title>
			<style>
				body { background: #0a0a0a; color: #00ff00; font-family: monospace; padding: 20px; max-width: 800px; margin: 0 auto; }
				.alert { color: #ff0000; border: 1px solid #ff0000; padding: 10px; margin-bottom: 20px; font-weight: bold; text-align: center; }
				ul { list-style-type: none; padding: 0; background: #111; border: 1px solid #333; padding: 15px; border-radius: 5px; }
				li { padding: 10px 0; border-bottom: 1px dashed #333; }
				a:hover { color: #fff !important; text-decoration: underline !important; }
				.btn { background: #ff8c00; color: #000; padding: 10px 15px; text-decoration: none; font-weight: bold; border-radius: 3px; display: inline-block; margin-bottom: 20px; border: none; cursor: pointer; }
				.btn:hover { background: #e07b00; }
				.public-link { display: block; margin-top: 30px; text-align: center; color: #ff8c00; text-decoration: none; border-top: 1px solid #333; padding-top: 20px; }
			</style>
		</head>
		<body>
			<h2>RSJ-V3 ADMIN CONSOLE</h2>
			<div class="alert">WARNING: RESTRICTED SOVEREIGN LAYER</div>
			
			<a href="?cmd=uptime" class="btn">⚡ RUN SYSTEM DIAGNOSTIC</a>
			%s
			
			<h3>Live Repository Vault:</h3>
			%s
			
			<a href="/" class="public-link">[ EXIT TO PUBLIC SITE ]</a>
		</body>
		</html>
		`, cmdOutput, fileListHTML)
	}

	http.HandleFunc("/admin", BasicAuth(adminHandler, "admin", "rsjv3"))

	fmt.Println("[RSJ-V3] Starting Dual-Layer Node on port 8081...")
	http.ListenAndServe(":8081", nil)
}
