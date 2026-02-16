#!/bin/bash
# RSJ-V3 Manual Trigger
# This "knocks" on the Juristic Agent's door to open the gate.
~/go/bin/grpcurl -plaintext -proto juristic.proto -d '{}' localhost:50052 juristic.v1.JuristicService/CommitState

