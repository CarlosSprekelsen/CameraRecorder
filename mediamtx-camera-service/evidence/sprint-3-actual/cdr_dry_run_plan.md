# MVP CDR Dry‑Run Test Sprint – Prompt‑Based Plan (Outside CDR)

**Purpose:** Run a short, test‑focused dry‑run to de‑risk the CDR. Tight scope on MVP only, managed by PM, executed by Developer and IV\&V. Outputs feed the CDR but are **not** part of it.

**Non‑Goals (to prevent scope creep):** No new features, no new camera models, no refactors. Only test harness fixes, configuration fixes, logging/observability tweaks, and defect fixes that unblock MVP tests.

**MVP Test Surface (reference IDs):**

* **SMK‑1**: Service starts from clean config; health check OK.
* **SMK‑2**: Auth: obtain/validate token; reject invalid/expired.
* **SMK‑3**: Discovery: at least one camera discovered in ≤ 10 s.
* **SMK‑4**: Snapshot path: `take_snapshot` returns file; EXIF/timestamp present.
* **SMK‑5**: Recording path: `start_recording` → file growing; `stop_recording` closes; playable.
* **SMK‑6**: Error handling: invalid RPC method produces documented error code.
* **SMK‑7**: Concurrency (light): two clients list cameras and one records; both remain responsive.

**Global Gates for Dry‑Run:**

```
Exit when: All SMK‑1..7 pass; flaky ≤ 5%; critical defects = 0; major defects ≤ 3 with workarounds documented.
Timebox: 3 working days (can compress to 2); 2 iterations max.
Change control: Fix-only unless PM grants a 1‑line scope waiver in writing.
```

---

## Phase A — Kickoff & Baseline (PM‑led)

### A1. Kickoff & Constraints

```
Role: Project Manager
Principles:
- Freeze scope to MVP tests (SMK‑1..7).
- Use a reproducible baseline tag to avoid moving targets.
Task:
- Announce dry‑run start, scope, gates, timebox, and comms channel.
Prompt to team:
"""
Goal: 3‑day MVP dry‑run to de‑risk CDR. Scope = SMK‑1..7 only. No features; fix‑only.
Baseline = vX.Y.Z-cdr.
Gates: All SMK pass; flaky ≤5%; no criticals.
Artifacts under: dry_run/.
Stand‑ups: 09:00 + 16:00.
Decision points: Iteration end of Day‑1 and Day‑2.
"""
Outputs:
- `dry_run/00_plan.md` (this plan + dates, owners)
- Slack/Teams post with the above prompt
```

### A2. Environment Freeze

```
Role: Developer (PM oversight)
Principles: Reproducible env; minimal variables.
Task:
- Check out baseline tag vX.Y.Z-cdr
- Export env matrix (OS, kernel, Python, MediaMTX, camera IDs)
- Produce `dry_run/env.sh` to set all needed env vars
Prompts:
PM→Dev: "Confirm env matrix and push `dry_run/env.sh` today."
Dev→PM reply (template): "Env frozen. SHA=<...>; OS=<...>; Python=<...>; MediaMTX=<...>; Cameras=<...>."
Outputs:
- `dry_run/01_env_matrix.md`
- `dry_run/env.sh`
```

---

## Phase B — Smoke Execution & Triage (IV\&V‑led)

### B1. Smoke Test Execution

```
Role: IV&V
Principles: Fast feedback, deterministic runs, logs preserved.
Task:
- Run smoke set: `pytest tests/smoke -q --maxfail=1 --ff --durations=10 | tee dry_run/02_smoke_run_1.log`
- Capture artifacts: logs, RPC traces, sample media files
Prompt:
IV&V→All: "Smoke run 1 started at <time>, baseline vX.Y.Z-cdr; log at dry_run/02_smoke_run_1.log."
Outputs:
- `dry_run/02_smoke_run_1.log`
- `dry_run/artifacts/` (snapshot.jpg, recording.mp4, rpc_trace.json)
```

### B2. Failure Triage (R/A/G + Timebox)

```
Role: PM (chair), Developer, IV&V
Principles: Decide fast; protect scope.
Task:
- Triage with R/A/G and timebox per issue (15 min each):
  - Red (Critical): blocks SMK; must fix now
  - Amber (Major): impacts quality; fix if < 0.5 day
  - Green (Minor): log for post‑CDR
Prompt:
PM→All: "Post results using template below by EOD. Reds get owners and ETAs."
Issue template:
- ID: DRY‑<nn>
- Affected SMK: <id>
- Symptom:
- Suspect area:
- Repro steps:
- Owner/ETA:
- Decision: (Fix now | Defer | Workaround)
Outputs:
- `dry_run/03_triage_register.md`
```

---

## Phase C — Focused Fix & Verify Loop (1–2 Iterations)

### C1. Fix Cycle (Red/Amber only)

```
Role: Developer
Principles: Small diffs; test first; logs > guesses.
Task:
- Add/adjust diagnostic logs only where needed
- Implement smallest fix; include regression test where trivial
- Push branch `fix/DRY-<nn>`; open PR with evidence
Prompt:
Dev→All: "Fix DRY‑<nn> pushed (diff <N> lines). Evidence at dry_run/fixes/DRY-<nn>/ . Ready for verify."
Outputs:
- `dry_run/fixes/DRY-<nn>/before_after.log`
- `dry_run/fixes/DRY-<nn>/patch.diff`
```

### C2. Verify Cycle

```
Role: IV&V
Principles: Re‑run only affected + smoke; keep it short.
Task:
- Targeted re‑run: `pytest -q -k "SMK_<id> or dependent" --maxfail=1 | tee dry_run/04_verify_DRY-<nn>.log`
- If pass, add to pass ledger; else bounce back once
Prompt:
IV&V→All: "Verify DRY‑<nn>: PASS/FAIL. Log at dry_run/04_verify_DRY-<nn>.log."
Outputs:
- `dry_run/04_verify_DRY-<nn>.log`
- Update `dry_run/05_pass_ledger.md`
```

### C3. Flake Hunt (if failures intermittent)

```
Role: Developer + IV&V
Principles: Prove flakiness; quarantine if needed.
Task:
- Repeat test N=10: `pytest -q -k "SMK_<id>" --count=10 --maxfail=1`
- If flake confirmed, add to quarantine; document cause/workaround
Outputs:
- `dry_run/06_flake_evidence.md`
- `dry_run/quarantine.txt` (tests skipped with reason)
```

---

## Phase D — Stabilization & Exit

### D1. Consolidated Re‑Run & Metrics

```
Role: IV&V
Principles: One truth pass; capture metrics.
Task:
- Full smoke re‑run: `pytest tests/smoke -q --maxfail=1 --ff --durations=10 | tee dry_run/07_smoke_run_final.log`
- Compute flaky %, pass rate, duration
Outputs:
- `dry_run/07_smoke_run_final.log`
- `dry_run/08_metrics.md` (pass rate, flaky %, top slow tests)
```

### D2. Dry‑Run Closeout

```
Role: Project Manager
Principles: Decision with evidence; channel learnings into CDR.
Task:
- Confirm gates met; if not, authorize a second (final) iteration or stop
- Publish closeout note and inputs that flow into CDR
Prompt:
PM→All: "Dry‑run decision: (ACCEPT | 2nd Iteration | STOP). Gates: pass=<x%>, flaky=<y%>, criticals=<n>. Next steps: <...>."
Outputs:
- `dry_run/09_closeout.md` (status, residual risks, handoffs to CDR steps 4–6/8/9)
```

---

## Working Prompts (Copy/Paste)

**PM Kickoff:**

```
Goal: 3‑day MVP dry‑run to de‑risk CDR. Scope = SMK‑1..7 only. No features; fix‑only.
Baseline = vX.Y.Z-cdr. Artifacts under dry_run/.
Gates: All SMK pass; flaky ≤5%; no criticals; ≤3 majors with workarounds.
Stand‑ups at 09:00/16:00. Decision points end of Day‑1/Day‑2.
Owners: Dev=<name>, IV&V=<name>. Comms: #dry-run.
```

**PM→Dev (Env Freeze):**

```
Please confirm environment on vX.Y.Z-cdr and commit `dry_run/env.sh` + `01_env_matrix.md` today. Include OS, kernel, Python, MediaMTX, camera IDs.
```

**IV\&V (Start Smoke):**

```
Starting Smoke Run 1 on vX.Y.Z-cdr now. Log -> dry_run/02_smoke_run_1.log. Will post R/A/G triage within 2h.
```

**PM (Triage Call):**

```
Post issues using template (DRY-<nn>, SMK id, Symptom, Suspect, Repro, Owner/ETA, Decision). Reds must have owners before EOD. Timebox 15 min/issue.
```

**Dev (Fix Notification):**

```
DRY-<nn> fix pushed (diff <N> lines). Evidence: dry_run/fixes/DRY-<nn>/. Ready for IV&V verify.
```

**IV\&V (Verify Result):**

```
Verify DRY-<nn>: PASS/FAIL. Log: dry_run/04_verify_DRY-<nn>.log. Re‑running affected smoke subset now.
```

**PM (Closeout):**

```
Dry‑run outcome: (ACCEPT | 2nd Iteration | STOP). Metrics: pass=<x%>, flaky=<y%>, criticals=<n>, majors=<m>.
Residual risks + owners captured in dry_run/09_closeout.md. Feeding forward to CDR Steps 4–6/8/9.
```

---

## File/Folder Skeleton (add to repo)

```
dry_run/
  00_plan.md
  01_env_matrix.md
  env.sh
  02_smoke_run_1.log
  03_triage_register.md
  04_verify_DRY-<nn>.log
  05_pass_ledger.md
  06_flake_evidence.md
  07_smoke_run_final.log
  08_metrics.md
  09_closeout.md
  artifacts/
  fixes/DRY-<nn>/{before_after.log, patch.diff}
  quarantine.txt
```

## Command Cheat‑Sheet

```
# Env
source dry_run/env.sh

# First smoke
pytest tests/smoke -q --maxfail=1 --ff --durations=10 | tee dry_run/02_smoke_run_1.log

# Targeted verify
pytest -q -k "SMK_<(id)>" --maxfail=1 | tee dry_run/04_verify_DRY-<nn>.log

# Flake check (requires pytest-xdist/pytest-rerunfailures or pytest-repeat)
pytest -q -k "SMK_<(id)>" --count=10 --maxfail=1

# Final smoke
pytest tests/smoke -q --maxfail=1 --ff --durations=10 | tee dry_run/07_smoke_run_final.log
```

**That’s it:** tightly scoped, prompt‑driven, and timeboxed. This dry‑run discovers and burns down the riskiest test failures without expanding scope, feeding cleanly into CDR execution steps.
