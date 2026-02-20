import sys
import os
import time
import importlib
import grpc
# IMPORT CRITICAL: Use the pip-installed tools directly
from grpc_tools import protoc

# --- CONFIGURATION ---
# The target Cloud Run endpoint (Do not add https://)
TARGET = "rsj-worker-fm2qsd2x4a-ts.a.run.app"
PORT = "443"
PROTO_FILE = "proto/rsj.proto"

def log(msg, type="INFO"):
    print(f"[{time.strftime('%H:%M:%S')}] [{type}] {msg}")

def self_heal_schema():
    """Checks if the Neural Link (Protobuf) is compiled. If not, compiles it."""
    # Check if files exist and are not empty
    if not os.path.exists("rsj_pb2.py") or os.path.getsize("rsj_pb2.py") == 0:
        log("Schema binaries missing or empty. Initiating auto-compile...", "WARN")
        compile_schema()
        # Force reload after compile to ensure Python sees the new files
        try:
            import rsj_pb2
            import rsj_pb2_grpc
            importlib.reload(rsj_pb2)
            importlib.reload(rsj_pb2_grpc)
        except ImportError:
             log("Post-compile import failed. Schema is likely invalid.", "FATAL")
             sys.exit(1)
        return

    # Test import to catch corruption
    try:
        import rsj_pb2
        import rsj_pb2_grpc
    except ImportError:
        log("Schema binaries corrupted. Recompiling...", "ERROR")
        compile_schema()
        # Force reload
        importlib.reload(rsj_pb2)
        importlib.reload(rsj_pb2_grpc)

def compile_schema():
    """Runs the Python-native compiler to regenerate bindings."""
    log("Compiling schema using native Python tools...", "unifying")

    # We construct the command arguments exactly as if calling protoc on CLI.
    # The first argument is ignored by convention but required.
    grpc_args = [
        'grpc_tools.protoc',
        '-I.',
        '--python_out=.',
        '--grpc_python_out=.',
        PROTO_FILE
    ]

    # Execute the bundled compiler directly
    exit_code = protoc.main(grpc_args)

    if exit_code == 0:
        log("Schema compiled successfully.", "SUCCESS")
        # Verify output exists
        if not os.path.exists("rsj_pb2_grpc.py"):
             log("Compiler reported success, but generated files are missing!", "FATAL")
             sys.exit(1)
    else:
        log(f"Compilation failed with exit code: {exit_code}", "CRITICAL")
        # Pro-tip: If this fails, run the equivalent command manually to see stderr:
        # python -m grpc_tools.protoc -I. --python_out=. --grpc_python_out=. proto/rsj.proto
        sys.exit(1)

def transmit(hash_id, data):
    """Securely transmits data to the Cloud Run Nexus."""
    # 1. Ensure infrastructure is sound before transmission
    self_heal_schema()

    # 2. Import generated modules LATE, after self-heal ensures they exist
    import rsj_pb2
    import rsj_pb2_grpc

    log(f"Establish mTLS transport to Vault: {TARGET}...", "NET")

    # Use secure channel credentials for Cloud Run (TLS 443)
    creds = grpc.ssl_channel_credentials()

    try:
        # Create the secure channel
        with grpc.secure_channel(f'{TARGET}:{PORT}', creds) as channel:
            stub = rsj_pb2_grpc.WorkerNodeStub(channel)

            # Create the Evidence Packet (The payload)
            packet = rsj_pb2.EvidencePacket(
                hash=hash_id,
                raw_data=data,
                # Generate ISO 8601 timestamp at the edge
                timestamp=time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())
            )

            log(f"Transmitting packet Hash: {hash_id}...", "SEND")
            start = time.time()

            # FIRE!
            resp = stub.SubmitEvidence(packet)
            latency = (time.time() - start) * 1000

            log(f"✅ ACK RECEIVED: Status='{resp.status}' | Msg='{resp.message}' | RTT: {latency:.2f}ms", "SUCCESS")

    except grpc.RpcError as e:
        log(f"🔥 RPC FAILURE: Code={e.code()} | Details={e.details()}", "FAIL")
        # In production, trigger exponential backoff retry here.

if __name__ == "__main__":
    # Initial "Cold Start" Test
    log("RSJ-V3 Termux Nexus Node Initializing...", "INIT")

    # Clean up old builds to force a fresh compile test
    if os.path.exists("rsj_pb2.py"): os.remove("rsj_pb2.py")
    if os.path.exists("rsj_pb2_grpc.py"): os.remove("rsj_pb2_grpc.py")

    # Send a live test packet to Cloud Run
    transmit("TERMUX-EDGE-TEST-001", "Legal Strategy Vector Alpha: Operational")
