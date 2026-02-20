import subprocess
import time
import os

# Configuration
MASTER_SCRIPT = "/sdcard/rsj_v3_master.sh"
CHECK_INTERVAL = 300 # 5 minutes

def check_integrity():
    # Check if animations are still off (proxy for system reset)
    result = subprocess.run(['settings', 'get', 'global', 'window_animation_scale'], capture_output=True, text=True)
    if result.stdout.strip() != "0":
        print("[RSJ-V3] Integrity Breach Detected. Re-hardening...")
        subprocess.run(['sh', MASTER_SCRIPT])
    else:
        print("[RSJ-V3] System Integrity: SECURE")

if __name__ == "__main__":
    print("[RSJ-V3] Guardian Module Started.")
    while True:
        try:
            check_integrity()
            time.sleep(CHECK_INTERVAL)
        except KeyboardInterrupt:
            break
