import grpc
import time
import sys
from grpc_tools import protoc

# Auto-compile the schema on launch
protoc.main((
    '',
    '-I.',
    '--python_out=.',
    '--grpc_python_out=.',
    'proto/rsj.proto',
))

import rsj_pb2
import rsj_pb2_grpc

# --- CONFIGURATION ---
TARGET = "rsj-worker-fm2qsd2x4a-ts.a.run.app"
PORT = "443"

def transmit(hash_id, data):
    print(f"\n[EDGE] 📡 Connecting to Vault: {TARGET}...")
    
    # Secure SSL Channel (Cloud Run requires this)
    creds = grpc.ssl_channel_credentials()
    
    try:
        with grpc.secure_channel(f'{TARGET}:{PORT}', creds) as channel:
            stub = rsj_pb2_grpc.WorkerNodeStub(channel)
            
            # Build Packet
            packet = rsj_pb2.EvidencePacket(
                hash=hash_id,
                raw_data=data,
                timestamp=time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())
            )
            
            # Transmit
            print(f"[EDGE] ⚡ Firing Packet: {hash_id}...")
            start = time.time()
            resp = stub.SubmitEvidence(packet)
            latency = (time.time() - start) * 1000
            
            # Report
            print(f"[RECV] ✅ Status: {resp.status}")
            print(f"       📝 Message: {resp.message}")
            print(f"       ☁️  Cloud Time: {resp.processing_time_ms}ms")
            print(f"       📶 Latency: {latency:.2f}ms")
            
    except grpc.RpcError as e:
        print(f"\n[FAIL] ❌ RPC Error: {e.code()}")
        print(f"       Details: {e.details()}")

if __name__ == "__main__":
    transmit("TERMUX-INIT-001", "RSJ-V3 Uplink Established")
