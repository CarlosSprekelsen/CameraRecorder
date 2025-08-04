#!/usr/bin/env python3
import pathlib
import textwrap
import sys
from typing import List, Iterable

MAX_WIDTH = 88  # must match flake8 / black config

def reflow_comment_block(lines: List[str], indent: str) -> List[str]:
    stripped = [line.lstrip()[1:].lstrip() for line in lines]
    joined = " ".join(s.strip() for s in stripped)
    wrapped = textwrap.wrap(joined, width=MAX_WIDTH - len(indent) - 2)
    return [f"{indent}# {w}" for w in wrapped]

def process_file(path: pathlib.Path):
    text = path.read_text(encoding="utf-8")
    new_lines = []
    i = 0
    lines = text.splitlines()
    while i < len(lines):
        line = lines[i]
        if line.lstrip().startswith("#"):
            indent = line[: line.index("#")]
            block = []
            while i < len(lines) and lines[i].lstrip().startswith("#"):
                block.append(lines[i])
                i += 1
            if any(len(l) > MAX_WIDTH for l in block):
                new_block = reflow_comment_block(block, indent)
                new_lines.extend(new_block)
            else:
                new_lines.extend(block)
        else:
            if i == 0 and (line.strip().startswith('"""') or line.strip().startswith("'''")):
                quote = line.strip()[:3]
                doc_lines = [line]
                i += 1
                while i < len(lines):
                    doc_lines.append(lines[i])
                    if lines[i].strip().endswith(quote):
                        i += 1
                        break
                    i += 1
                if len(doc_lines) >= 3:
                    start = doc_lines[0]
                    end = doc_lines[-1]
                    inner = doc_lines[1:-1]
                    inner_text = " ".join(l.strip() for l in inner)
                    wrapped_inner = textwrap.wrap(inner_text, width=MAX_WIDTH)
                    new_doc = [start] + wrapped_inner + [end]
                    new_lines.extend(new_doc)
                else:
                    new_lines.extend(doc_lines)
            else:
                new_lines.append(line)
                i += 1
    new_content = "\n".join(new_lines).rstrip() + "\n"  # ensure single trailing newline
    if new_content != text:
        path.write_text(new_content, encoding="utf-8")
        print(f"Reflowed comments in {path}")
    else:
        print(f"No change needed for {path}")

def gather_targets(paths: Iterable[str]):
    for p in paths:
        path = pathlib.Path(p)
        if path.is_dir():
            for py in path.rglob("*.py"):
                yield py
        elif path.is_file() and path.suffix == ".py":
            yield path
        else:
            # allow globs already expanded by shell
            continue

def main():
    if len(sys.argv) < 2:
        print("Usage: reflow_comments.py <file_or_dir> [more...]")
        sys.exit(1)
    for p in gather_targets(sys.argv[1:]):
        process_file(p)

if __name__ == "__main__":
    main()
