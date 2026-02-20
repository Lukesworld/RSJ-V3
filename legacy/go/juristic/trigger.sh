#!/bin/bash
# Trigger a CommitState call to the Juristic Gate
grpcurl -plaintext -proto juristic.proto -d '{"new_state": "AUDIT_COMPLETED", "principal_id": "GEMINI_CLI_ARCHITECT"}' localhost:50053 rsj.v3.sovereign.SovereignNode/CommitState
