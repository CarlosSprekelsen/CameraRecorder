#!/usr/bin/env python3
"""
Strip UTF-8 BOM and zero-width no-break spaces from Python source files.

Targets only .py files under src/ and tests/.
"""
from __future__ import annotations

import sys
from pathlib import Path


def strip_bom_from_file(path: Path) -> bool:
    data = path.read_bytes()
    original = data
    # Remove UTF-8 BOM if present
    if data.startswith(b"\xef\xbb\xbf"):
        data = data[3:]
    try:
        text = data.decode("utf-8", errors="replace")
    except Exception:
        return False
    # Remove any stray U+FEFF characters inside the file
    if "\ufeff" in text:
        text = text.replace("\ufeff", "")
        data = text.encode("utf-8")
    if data != original:
        path.write_bytes(data)
        return True
    return False


def main() -> int:
    repo_root = Path(__file__).resolve().parent.parent
    targets = [repo_root / "src", repo_root / "tests"]
    changed = 0
    scanned = 0
    for base in targets:
        if not base.exists():
            continue
        for path in base.rglob("*.py"):
            scanned += 1
            try:
                if strip_bom_from_file(path):
                    changed += 1
                    print(f"Fixed BOM: {path.relative_to(repo_root)}")
            except Exception as e:
                print(f"WARN: failed to process {path}: {e}")
    print(f"Scanned {scanned} files; fixed {changed} files.")
    return 0


if __name__ == "__main__":
    sys.exit(main())


