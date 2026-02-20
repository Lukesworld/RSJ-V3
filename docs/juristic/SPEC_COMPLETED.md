# Product Specification: Juristic Gatekeeper Stabilization

## 1. Objective
Stabilize the Juristic Gatekeeper (v4.1) within the RSJ-V3 monorepo to ensure sovereign authorization for all state-changing gRPC calls.

## 2. Technical Scope
*   **Component:** Juristic Gatekeeper (Sovereign Node)
*   **Protocols:** gRPC (Port 50053), Unix Socket (TOTP Verification)
*   **Key Files:**
    *   `projects/web/rsj-v3-monorepo/juristic/main.go`
    *   `projects/web/rsj-v3-monorepo/juristic/juristic_agent`
    *   `scripts/start_services.sh`

## 3. Functional Requirements
*   [x] Juristic Gatekeeper binary is functional.
*   [ ] Juristic Gatekeeper starts automatically via `start_services.sh`.
*   [ ] `RSJV3_SECRET` environment variable is securely managed or verified before startup.
*   [ ] Interceptor correctly blocks `CommitState` until TOTP is provided via `/data/data/com.termux/files/usr/tmp/juristic_gate.sock`.

## 4. Constraints & Compliance
*   **Hallucination Check:** No imaginary imports.
*   **Sovereignty:** Must run locally on Termux.
*   **Security:** Socket must have 0600 permissions (enforced in code).

## 5. Acceptance Criteria (Auditor Checklist)
*   [ ] `juristic_agent` runs in the background.
*   [ ] `start_services.sh` is updated and verified.
*   [ ] `CommitState` gRPC call is successfully intercepted and released upon valid TOTP.
*   [ ] `map.json` correctly reflects the `juristic` service and its dependencies.
