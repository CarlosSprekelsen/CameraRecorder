#!/usr/bin/env python3
import os
import re

BASE_DIR = "/home/dts/CameraRecorder/dry_run"
LOG_PATH = os.path.join(BASE_DIR, "07_smoke_run_final.log")


def read_text(path: str) -> str:
    try:
        with open(path, "r", errors="ignore") as f:
            return f.read().strip()
    except Exception:
        return ""


def main() -> None:
    start_iso = read_text(os.path.join(BASE_DIR, ".start_iso"))
    end_iso = read_text(os.path.join(BASE_DIR, ".end_iso"))
    try:
        start_epoch = int(read_text(os.path.join(BASE_DIR, ".start_epoch")) or "0")
        end_epoch = int(read_text(os.path.join(BASE_DIR, ".end_epoch")) or "0")
        wall = end_epoch - start_epoch if end_epoch >= start_epoch else 0
    except Exception:
        wall = 0

    passed = failed = errors = skipped = xfailed = xpassed = total = 0

    lines: list[str] = []
    if os.path.exists(LOG_PATH):
        with open(LOG_PATH, "r", errors="ignore") as f:
            lines = f.read().splitlines()

    # Find last line with a percentage marker and extract progress token
    token: str | None = None
    for s in reversed(lines):
        s = s.strip()
        if "[" in s and "]" in s and "%" in s:
            first = s.split()[0]
            if re.fullmatch(r"[\.FEsxX]+", first or ""):
                token = first
                break

    if token:
        passed = token.count('.')
        failed = token.count('F')
        errors = token.count('E')
        skipped = token.count('s')
        xfailed = token.count('x')
        xpassed = token.count('X')
        total = passed + failed + errors + skipped + xfailed + xpassed

    # Extract slow block
    slow: list[str] = []
    if lines:
        data = "\n".join(lines)
        m = re.search(r"=+\s*slowest 10 durations\s*=+\n([\s\S]*?)\n=+", data)
        if m:
            slow = [ln.strip() for ln in m.group(1).strip().splitlines() if ln.strip()][:12]
        else:
            m2 = re.search(r"=+\s*slowest 10 durations\s*=+\n([\s\S]*)$", data)
            if m2:
                slow = [ln.strip() for ln in m2.group(1).strip().splitlines() if ln.strip()][:12]

    denom = passed + failed + errors
    if denom > 0:
        pass_rate = (passed / denom) * 100.0
    else:
        pass_rate = 100.0 if total > 0 else 0.0

    flaky_pct = 0.0  # single run; flakiness not sampled here

    out_path = os.path.join(BASE_DIR, "08_metrics.md")
    with open(out_path, "w") as f:
        f.write("# Smoke Run Metrics (Final)\n\n")
        f.write(f"- Start: {start_iso}\n")
        f.write(f"- End: {end_iso}\n")
        f.write(f"- Wall duration: {wall}s\n\n")
        f.write("## Results\n")
        f.write(f"- Total tests (incl. skipped): {total}\n")
        f.write(f"- Passed: {passed}\n")
        f.write(f"- Failed: {failed}\n")
        f.write(f"- Errors: {errors}\n")
        f.write(f"- Skipped: {skipped}\n")
        f.write(f"- XFailed: {xfailed}\n")
        f.write(f"- XPassed: {xpassed}\n")
        f.write(f"- Pass rate: {pass_rate:.2f}%\n")
        f.write(f"- Flaky: {flaky_pct:.2f}%\n\n")
        f.write("## Top slow tests\n")
        if slow:
            for ln in slow:
                f.write(f"- {ln}\n")
        else:
            f.write("- (no slow block reported)\n")


if __name__ == "__main__":
    main()
