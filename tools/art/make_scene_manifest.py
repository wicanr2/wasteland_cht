#!/usr/bin/env python3
"""Write a provenance manifest for one accepted faithful-HD scene batch."""

import argparse
import hashlib
import json
from pathlib import Path


def sha256(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as src:
        for chunk in iter(lambda: src.read(1024 * 1024), b""):
            digest.update(chunk)
    return digest.hexdigest()


def main() -> None:
    ap = argparse.ArgumentParser()
    ap.add_argument("--start", type=int, required=True)
    ap.add_argument("--end", type=int, required=True)
    ap.add_argument("--output", type=Path, required=True)
    ap.add_argument("--root", type=Path, default=Path("."))
    args = ap.parse_args()

    scenes = []
    for index in range(args.start, args.end + 1):
        source = Path(f"artwork/source/accepted/scenes/{index:03d}-source.png")
        result = Path(f"artpacks/faithful-hd/assets/scenes/{index:03d}.png")
        reference = Path(f"workplace/art-reference/scenes-8x/{index:03d}.png")
        for path in (source, result, reference):
            if not (args.root / path).is_file():
                raise SystemExit(f"missing: {path}")
        scenes.append({
            "index": index,
            "reference": str(reference),
            "reference_sha256": sha256(args.root / reference),
            "source": str(source),
            "source_sha256": sha256(args.root / source),
            "path": str(result),
            "sha256": sha256(args.root / result),
        })

    manifest = {
        "schema": 1,
        "id": f"faithful-hd-scenes-{args.start:03d}-{args.end:03d}",
        "created_local": "2026-08-21",
        "timezone": "Asia/Taipei",
        "generator": "OpenAI image generation via Codex",
        "source_prompt": "artwork/prompts/scenes-faithful-hd.md",
        "reference_distribution": "local-only",
        "processing": {"fit": "tools/art/fit_scene.py", "size": "288x252"},
        "status": "accepted-candidate",
        "scenes": scenes,
        "gates": {
            "dimensions": "pass",
            "semantic_reference_review": "pass",
            "false_affordance_review": "pass",
            "in_game_review": "pending",
        },
    }
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(json.dumps(manifest, ensure_ascii=False, indent=2) + "\n")


if __name__ == "__main__":
    main()
