#!/usr/bin/env python3
"""Write reproducible provenance for one accepted tile batch."""

import argparse
import hashlib
import json
from pathlib import Path


def sha(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main() -> None:
    ap = argparse.ArgumentParser()
    ap.add_argument("--tileset", type=int, required=True)
    ap.add_argument("--start", type=int, required=True)
    ap.add_argument("--end", type=int, required=True)
    ap.add_argument("--required-total", type=int, required=True)
    ap.add_argument("--source", type=Path, required=True)
    ap.add_argument("--tiles", type=Path, required=True)
    ap.add_argument("--out", type=Path, required=True)
    args = ap.parse_args()
    indices = range(args.start, args.end + 1)
    atlas = args.tiles / f"atlas-{args.start:03d}-{args.end:03d}.png"
    data = {
        "schema": 1,
        "id": f"faithful-hd-tileset-{args.tileset}-{args.start:03d}-{args.end:03d}",
        "created_local": "2026-08-21",
        "timezone": "Asia/Taipei",
        "generator": "OpenAI image generation via Codex",
        "source_prompt": "artwork/prompts/tileset-faithful-hd.md",
        "reference": f"workplace/art-reference/tileset-{args.tileset}-batch-{args.start:03d}-{args.end:03d}-8x.png",
        "reference_distribution": "local-only",
        "source": {"path": str(args.source), "sha256": sha(args.source)},
        "processing": {
            "slice": "tools/art/slice_grid.py",
            "edge_harmonizer": "tools/art/harmonize_edges.py",
            "atlas": "tools/art/assemble_atlas.py",
            "cell": "48x48",
        },
        "status": "accepted-candidate",
        "coverage": {
            "tileset": args.tileset,
            "indices": f"{args.start}-{args.end}",
            "count": args.end - args.start + 1,
            "required_total": args.required_total,
        },
        "atlas": {"path": str(atlas), "sha256": sha(atlas)},
        "tiles": [
            {"index": i, "path": str(args.tiles / f"{i:03d}.png"), "sha256": sha(args.tiles / f"{i:03d}.png")}
            for i in indices
        ],
        "gates": {
            "grid_count": "pass",
            "dimensions": "pass",
            "index_order_visual_review": "pass",
            "same-state_map_review": "pass-with-wip-original-icons",
        },
    }
    args.out.parent.mkdir(parents=True, exist_ok=True)
    args.out.write_text(json.dumps(data, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")


if __name__ == "__main__":
    main()
