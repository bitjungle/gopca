#!/usr/bin/env python3
"""
Verify and optionally extend the curated OWID dataset catalog at
cmd/gocsv/owid_catalog.json.

Usage (from repo root, no venv needed — only stdlib used):

    python scripts/generate_owid_catalog.py [--check]

With --check: validates that every path in the catalog returns HTTP 200.
Without arguments: prints the current catalog summary.

To add new datasets, probe candidate URLs manually:

    curl -I "https://catalog.ourworldindata.org/<path>.parquet"

Then add a new entry to cmd/gocsv/owid_catalog.json with:
  namespace, dataset, version, table, title, description, path

OWID URL pattern (from https://docs.owid.io/projects/etl/api/catalog-api/):
  https://catalog.ourworldindata.org/{channel}/{namespace}/{version}/{dataset}/{table}.parquet

Example:
  https://catalog.ourworldindata.org/garden/energy/2024-06-20/energy_mix/energy_mix.parquet
"""

import json
import sys
import urllib.request

CATALOG_PATH = "cmd/gocsv/owid_catalog.json"
BASE_URL = "https://catalog.ourworldindata.org/"

with open(CATALOG_PATH) as f:
    catalog = json.load(f)

print(f"Catalog: {len(catalog)} datasets\n")
for d in catalog:
    print(f"  [{d['namespace']}/{d['dataset']}] {d['title']}")

if "--check" in sys.argv:
    print("\nVerifying URLs...")
    ok = 0
    fail = 0
    for d in catalog:
        url = BASE_URL + d["path"] + ".parquet"
        try:
            req = urllib.request.Request(url, method="HEAD")
            with urllib.request.urlopen(req, timeout=30) as r:
                size = r.headers.get("Content-Length", "?")
                print(f"  200  {size:>12}B  {d['path']}")
                ok += 1
        except Exception as e:
            print(f"  ERR  {'?':>12}   {d['path']} — {e}")
            fail += 1
    print(f"\n{ok} OK, {fail} failed")
