# RSJ-V3: Sovereign Jurisdictional Framework

## Overview
RSJ-V3 is a distributed sovereign node framework designed for jurisdictional autonomy and decentralized evidence processing. Originally a hybrid Python/Go system, it has been fully refactored into a high-performance, strictly typed Go monorepo.

## Architecture

### Services
- **Juristic (Gatekeeper):** The central authority node handling state synchronization and directive dispatch. Uses strict gRPC contracts.
- **Cloud Worker:** Scalable evidence processing unit integrating Vertex AI for analysis.
- **Mesh Node:** Edge node for secure evidence submission.
- **Audit Engine:** Forensic accounting module using Bayesian inference for fraud detection.
- **RSJ Agent:** Autonomous agent running on sovereign hardware, maintaining heartbeat and executing directives.
- **RSJ Guardian:** System integrity monitor.

### Key Technologies
- **Language:** Go 1.25+
- **Communication:** gRPC / Protobuf
- **AI:** Google Vertex AI (Gemini Pro)
- **Security:** mTLS, TOTP (Time-based One-Time Password)
- **Concurrency:** Worker Pools (pattern-based)

## Directory Structure
```
rsj-v3-monorepo/
├── cmd/                # Entry points for binaries
│   ├── audit-engine    # Forensic audit tool
│   ├── mesh-node       # Edge client
│   ├── rsj-agent       # Sovereign agent daemon
│   └── rsj-guardian    # Integrity monitor
├── pkg/                # Shared libraries
│   ├── audit           # Bayesian inference logic
│   ├── juristic        # Generated gRPC code (SovereignNode)
│   ├── worker          # Generated gRPC code (WorkerNode)
│   └── workerpool      # Concurrency utility
├── services/           # Backend services
│   ├── cloud-worker    # Evidence processor
│   └── gatekeeper      # Control plane (Juristic)
├── proto/              # Protocol Buffer definitions
├── bin/                # Compiled binaries
└── legacy/             # Archived Python code
```

## Getting Started

### Prerequisites
- Go 1.25 or higher
- Protoc (Protocol Buffers Compiler)
- Google Cloud Project with Vertex AI enabled (for Cloud Worker)

### Building
Use the provided Makefile to build all services:
```bash
make
```
This produces binaries in the `bin/` directory.

### Running Tests
```bash
make test
```

### Deployment
Binaries are self-contained. Deploy `bin/gatekeeper` as the central node, and `bin/rsj-agent` on edge devices.

## API Documentation
See `proto/` for service definitions (`juristic.proto`, `rsj.proto`).

## License
Private / Sovereign Use Only.
