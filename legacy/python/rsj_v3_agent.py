import grpc
from concurrent import futures
import time
import logging
import threading
import random
import os
import sys

# Attempt to import generated gRPC modules
try:
    import juristic_pb2
    import juristic_pb2_grpc
except ImportError:
    # Fallback for when protobuf files aren't generated yet
    logging.warning("Protobuf modules not found. Using MOCK fallback.")
    
    class juristic_pb2:
        class DirectiveAck:
            def __init__(self, directive_id, accepted, reason): pass
        class HeartbeatRequest:
            def __init__(self, agent_id, load_average, uptime_seconds):
                self.agent_id = agent_id
                self.load_average = load_average
                self.uptime_seconds = uptime_seconds

    class juristic_pb2_grpc:
        class SovereignNodeServicer:
            pass
        
        class SovereignNodeStub:
            def __init__(self, channel):
                pass
            def SendHeartbeat(self, request, timeout=None):
                logging.info(f"[MOCK] Sending Heartbeat for {request.agent_id}")
                class MockResponse:
                    pending_directives = False
                return MockResponse()

        @staticmethod
        def add_SovereignNodeServicer_to_server(servicer, server):
            pass

# Configure Logging
LOG_FILE = os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))), "logs", "agent.log")
os.makedirs(os.path.dirname(LOG_FILE), exist_ok=True)

# Load .env manually if available
env_path = os.path.join(os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__)))), '.env')
if os.path.exists(env_path):
    try:
        with open(env_path) as f:
            for line in f:
                if line.strip() and not line.startswith('#'):
                    parts = line.strip().split('=', 1)
                    if len(parts) == 2:
                        k, v = parts
                        if k.startswith("export "):
                            k = k[7:]
                        os.environ[k] = v.strip('"\'')
    except Exception as e:
        logging.warning(f"Failed to load .env: {e}")

JURISTIC_PORT = os.getenv("JURISTIC_PORT", "50053")
AGENT_PORT = os.getenv("AGENT_PORT", "35507")

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - [AGENT] %(message)s',
    handlers=[
        logging.FileHandler(LOG_FILE),
        logging.StreamHandler(sys.stdout)
    ]
)

class AgentState:
    def __init__(self, agent_id):
        self.agent_id = agent_id
        self.start_time = time.time()
        self.resource_allocation = 0.0

    def update_resources(self):
        self.resource_allocation = random.uniform(10.0, 90.0)

    def get_uptime(self):
        return int(time.time() - self.start_time)

class TaskHandler:
    def process_directive(self, directive):
        logging.info(f"TaskHandler: Received Directive ID: {directive.directive_id}")
        return True, "Executed successfully"

class RSJV3Agent(juristic_pb2_grpc.SovereignNodeServicer):
    def __init__(self, state, task_handler):
        self.state = state
        self.task_handler = task_handler

    def DispatchDirective(self, request, context):
        logging.info(f"gRPC: DispatchDirective received for {request.directive_id}")
        return juristic_pb2.DirectiveAck(directive_id=request.directive_id, accepted=True, reason="Executed")

def report_state_loop(state, stub, stop_event):
    logging.info("State Reporting Loop Started.")
    while not stop_event.is_set():
        try:
            state.update_resources()
            uptime = state.get_uptime()
            req = juristic_pb2.HeartbeatRequest(
                agent_id=state.agent_id,
                load_average=state.resource_allocation,
                uptime_seconds=uptime
            )
            logging.debug(f"Sending Heartbeat...")
            stub.SendHeartbeat(req, timeout=5)
        except Exception as e:
            logging.error(f"Unexpected error in reporting loop: {e}")
        stop_event.wait(60)

def serve():
    agent_id = f"agent-{random.randint(1000,9999)}"
    logging.info(f"Initializing RSJ-V3 Agent: {agent_id}")
    
    state = AgentState(agent_id)
    task_handler = TaskHandler()
    
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    if hasattr(juristic_pb2_grpc, 'add_SovereignNodeServicer_to_server'):
        juristic_pb2_grpc.add_SovereignNodeServicer_to_server(RSJV3Agent(state, task_handler), server)
    
    server.add_insecure_port(f'[::]:{AGENT_PORT}')
    server.start()
    logging.info(f"Agent Listener Active on Port {AGENT_PORT}")

    channel = grpc.insecure_channel(f'localhost:{JURISTIC_PORT}')
    stub = juristic_pb2_grpc.SovereignNodeStub(channel)

    stop_event = threading.Event()
    reporter = threading.Thread(target=report_state_loop, args=(state, stub, stop_event), daemon=True)
    reporter.start()

    try:
        while True:
            time.sleep(86400)
    except KeyboardInterrupt:
        stop_event.set()
        server.stop(0)

if __name__ == "__main__":
    serve()