# Deliverable Rejection Notice

**Date:** 2025-08-09
**Rejecting Role:** Project Manager
**Previous Role:** Developer
**Failed Deliverable:** evidence/sprint-3-actual/00_cdr_baseline_and_build.md

## Validation Failures
- [ ] Missing file: 
- [x] Missing sections: Baseline Definition incomplete (tag mismatch), Build Artifacts (no commands), Environment Matrix (no outputs), Checksums (empty), Reproducibility Verification (empty)
- [x] Missing actual outputs: uname -a, python --version, mediamtx version, lsusb, sha256sum values, clean venv import check, pip freeze excerpt
- [x] Missing requirements: Handoff criteria not met (complete baseline establishment with reproducible build evidence)

## Required Corrections
1. Populate Baseline Definition with correct tag at HEAD (e.g., v0.1.1-cdr) and commit SHA.
2. Under Build Artifacts, include actual build commands executed and list all artifacts produced.
3. Fill Environment Matrix with real outputs for uname -a, python --version, MediaMTX/rtsp-simple-server --version, and lsusb.
4. Generate and include sha256sum outputs for all artifacts in Checksums.
5. Perform clean environment build verification; include import check results and pip freeze excerpt in Reproducibility Verification.
6. Ensure all sections are complete per documentation-guidelines and contain actual command outputs, not placeholders.

## Return Instruction
Previous role must complete ALL corrections above before current role can proceed. Resubmit complete deliverable when all corrections made.

**Status:** REJECTED - Return to Developer for completion
