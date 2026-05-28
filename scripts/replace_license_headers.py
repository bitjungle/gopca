#!/usr/bin/env python3
"""
Replace or add license headers in all .go, .ts, .tsx source files.

Exclusions: node_modules/, dist/, wailsjs/, .venv/
"""

import os
import sys
from pathlib import Path

ROOT = Path(__file__).parent.parent

# Full 5-line header (the canonical old form used in most files)
OLD_HEADER = """\
// Copyright 2025-2026 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications."""

# Shorter 3-line header variant found in some files (without the military line)
OLD_HEADER_SHORT = """\
// Copyright 2025-2026 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file."""

# Single-line copyright comment found as a standalone comment in some files
OLD_HEADER_ONELINE = "// Copyright 2025-2026 bitjungle - Rune Mathisen. All rights reserved."

NEW_HEADER = """\
// GoPCA Suite
//
// Copyright © 2025-2026 Rune Mathisen <devel@bitjungle.com>
//
// This file is part of GoPCA Suite.
//
// GoPCA Suite is source-available software with free binary redistribution.
// Official compiled binary releases may be used and redistributed free of charge
// under the GoPCA Suite Source-Available Freeware License.
//
// The source code is provided for viewing, review, education, security analysis,
// research, interoperability analysis, and evaluation only.
//
// Modification, redistribution, publication, sublicensing, reuse, incorporation
// into another project, or creation of derivative works based on the source code
// is not permitted without prior written permission from the copyright holder.
//
// Usage Restriction: GoPCA Suite may not be used, directly or indirectly, for
// military, warfare, weapons, intelligence, surveillance, targeting, or
// law-enforcement surveillance applications.
//
// See LICENSE for the full license terms."""

EXCLUDED_DIRS = {"node_modules", "dist", "wailsjs", ".venv", "worktrees"}

EXTENSIONS = {".go", ".ts", ".tsx"}


def should_exclude(path: Path) -> bool:
    """Return True if path contains any excluded directory segment."""
    parts = path.parts
    for part in parts:
        if part in EXCLUDED_DIRS:
            return True
    return False


def find_source_files():
    """Yield all eligible source files under ROOT."""
    for root, dirs, files in os.walk(ROOT):
        # Prune excluded directories in-place so os.walk doesn't descend
        dirs[:] = [d for d in dirs if d not in EXCLUDED_DIRS]
        for fname in files:
            p = Path(root) / fname
            if p.suffix in EXTENSIONS:
                yield p


def process_file(path: Path, dry_run: bool = False) -> str:
    """
    Process a single file: replace old header or prepend new header.
    Returns a status string: 'replaced', 'added', or 'skipped'.
    """
    content = path.read_text(encoding="utf-8")

    if OLD_HEADER in content:
        new_content = content.replace(OLD_HEADER, NEW_HEADER, 1)
        if not dry_run:
            path.write_text(new_content, encoding="utf-8")
        return "replaced"

    if OLD_HEADER_SHORT in content:
        new_content = content.replace(OLD_HEADER_SHORT, NEW_HEADER, 1)
        if not dry_run:
            path.write_text(new_content, encoding="utf-8")
        return "replaced"

    # Single-line copyright comment — remove it (new header already prepended above it)
    if OLD_HEADER_ONELINE in content:
        # Remove the line including its trailing newline
        new_content = content.replace(OLD_HEADER_ONELINE + "\n", "", 1)
        if not dry_run:
            path.write_text(new_content, encoding="utf-8")
        return "replaced"

    # File doesn't have the old header — check if it already has new header
    if NEW_HEADER in content:
        return "skipped"

    # Need to prepend the new header.
    # For Go files with build constraints, the header must go BEFORE the build tag.
    # Go convention (pre-1.17 // +build, post-1.17 //go:build) requires build
    # tags before the package declaration, but copyright headers are expected
    # to precede build tags per gofmt conventions.
    # We place the new header at the very top (before build constraints).
    new_content = NEW_HEADER + "\n\n" + content
    if not dry_run:
        path.write_text(new_content, encoding="utf-8")
    return "added"


def main():
    """Scan all eligible source files and replace or add the license header."""
    dry_run = "--dry-run" in sys.argv

    replaced = []
    added = []
    skipped = []
    errors = []

    for path in sorted(find_source_files()):
        try:
            status = process_file(path, dry_run=dry_run)
            rel = path.relative_to(ROOT)
            if status == "replaced":
                replaced.append(str(rel))
            elif status == "added":
                added.append(str(rel))
            else:
                skipped.append(str(rel))
        except OSError as e:
            errors.append(f"{path}: {e}")

    print(f"\n=== License Header Update {'(DRY RUN) ' if dry_run else ''}===")
    print(f"  Replaced old header : {len(replaced)}")
    print(f"  Added new header    : {len(added)}")
    print(f"  Already up to date  : {len(skipped)}")
    print(f"  Errors              : {len(errors)}")

    if added:
        print("\nFiles where header was ADDED:")
        for f in added:
            print(f"  + {f}")

    if errors:
        print("\nERRORS:")
        for e in errors:
            print(f"  !! {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()
