#!/usr/bin/env python3
import subprocess
import json

def tor_request(command):
    """Run any command through Tor"""
    result = subprocess.run(
        ["torsocks"] + command.split(),
        capture_output=True,
        text=True
    )
    return result.stdout

def rsj_anonymous_operation():
    """RSJ V3 Framework running through Tor"""
    print("="*60)
    print("RSJ-V3 ANONYMOUS OPERATION")
    print("="*60)
    
    # Check Tor status
    tor_check = tor_request("curl -s https://check.torproject.org/api/ip")
    tor_data = json.loads(tor_check)
    
    if tor_data.get("IsTor"):
        print(f"✓ Operating through Tor")
        print(f"✓ Exit Node: {tor_data.get('IP')}")
        print(f"✓ Jurisdiction: UNRECOGNIZED (Sovereign Mode Active)")
        print("="*60)
        return True
    else:
        print("✗ Tor connection failed. RSJ-V3 requires anonymity.")
        return False

if __name__ == "__main__":
    if rsj_anonymous_operation():
        print("\nRSJ-V3 ready for anonymous operations.")
        # Your RSJ V3 logic here
    else:
        print("\nCannot proceed without Tor.")
