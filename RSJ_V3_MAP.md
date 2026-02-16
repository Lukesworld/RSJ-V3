# RSJ-V3 Architecture Map (Current State)

This document visualizes the current structure of your distributed `rsj-v3` system.

## 📂 Core Components

### 1. **Local Authorization Gate** (`/rsj-v3-juristic/`)
**Role:** The security brain. Running on Termux.
**Status:** Active. Protects critical actions with 2FA (TOTP).
```
rsj-v3-juristic/
├── main.go             # The Guard. Intercepts gRPC calls & checks TOTP.
├── juristic.proto      # Defines the service contract (CommitState).
├── juristic_agent      # Compiled binary of the gatekeeper.
├── trigger.sh          # Helper script to trigger actions.
└── go.mod / go.sum     # Go dependencies.
```

### 2. **Mesh Network / Edge Nodes** (`/rsj-v3-mesh/`)
**Role:** The hands. Python scripts running on Termux to gather evidence.
**Status:** Active. Connects to Cloud Run via mTLS.
```
rsj-v3-mesh/
├── nexus_node.py       # Robust client. Self-heals (recompiles protos) & sends data.
├── edge_node.py        # Lightweight client. Alternative implementation.
├── proto/              # Contains the .proto definitions for the mesh.
├── dashboard.sh        # Shell script for visualization?
└── rsj-edge            # Likely a compiled or symlinked binary.
```

### 3. **Cloud Backend** (`/rsj-worker-cloud/`)
**Role:** The vault. Running on Google Cloud Run.
**Status:** Deployed. Accepts evidence packets.
```
rsj-worker-cloud/
├── worker_grpc.go      # The Server. Logs received evidence hashes.
├── proto/              # Contains .proto definitions for the worker.
└── Dockerfile          # Instructions for building the cloud container.
```

### 4. **Python Source Libs** (`/rsj_v3/`)
**Role:** Utility libraries.
```
rsj_v3/
└── src/
    └── audit_engine.py # Likely logic for auditing actions locally.
```

### 5. **Legacy / Fragmented Folders**
**Status:** These appear to be older versions or build artifacts.
*   `/rsj-v3/`: Contains `juristic/` subfolder with generated Go code. Likely duplicative of `rsj-v3-juristic`.

---

## 🔄 Data Flow

1.  **User Action** -> Triggers `rsj-v3-juristic/main.go` (The Gate).
2.  **Gate Check** -> Verifies TOTP via local socket.
3.  **Approval** -> Allows `rsj-v3-mesh/nexus_node.py` to execute.
4.  **Transmission** -> `nexus_node.py` sends `EvidencePacket` (Protobuf) to `rsj-worker-cloud`.
5.  **Storage** -> `rsj-worker-cloud` receives and logs the evidence (Cloud Run).

## ⚠️ Recommendations for Restructuring

To make this easier to manage, I recommend consolidating into a **Monorepo Structure**:

```
~/rsj-v3-monorepo/
├── services/
│   ├── gatekeeper/     (Move rsj-v3-juristic here)
│   └── cloud-worker/   (Move rsj-worker-cloud here)
├── clients/
│   └── mesh-node/      (Move rsj-v3-mesh contents here)
├── libs/
│   └── python-utils/   (Move rsj_v3/src contents here)
└── proto/              (Centralize all .proto files here)
```
