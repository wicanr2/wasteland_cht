#!/usr/bin/env python3
"""Derive editable Simplified Chinese TSV from the approved Traditional text.

Run inside Docker. Supply the fixed wheel recorded in localization-inputs.json.
Only display values are converted; keys and original-language shared text stay.
The output TSV is the runtime catalogue's source, not a runtime transliterator.
"""
import argparse
import hashlib
import json
from pathlib import Path
import sys
import tempfile
import zipfile

from build_paragraphs import parse, to_text
from extract_strings import escape, unescape


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("wheel", type=Path)
    ap.add_argument("--root", type=Path, default=Path("translations"))
    args = ap.parse_args()
    record = json.loads(Path(__file__).with_name("localization-inputs.json").read_text())
    assert hashlib.sha256(args.wheel.read_bytes()).hexdigest() == record["sha256"]
    dest = args.root / "zh-Hans"
    dest.mkdir(exist_ok=True)
    with tempfile.TemporaryDirectory() as temporary:
        with zipfile.ZipFile(args.wheel) as archive:
            assert all(not Path(n).is_absolute() and ".." not in Path(n).parts for n in archive.namelist())
            archive.extractall(temporary)
        sys.path.insert(0, temporary)
        from opencc import OpenCC
        converter = OpenCC(record["configuration"])
        for p in sorted((args.root / "zh-Hant").glob("*.tsv")):
            lines = []
            for line in p.read_text().splitlines():
                if not line.strip() or line.startswith("#"):
                    lines.append(line)
                    continue
                key, value = line.split("\t", 1)
                lines.append(key + "\t" + escape(converter.convert(unescape(value))))
            (dest / p.name).write_text("\n".join(lines) + "\n")
        paragraphs = {}
        for p in sorted((args.root / "zh-Hant" / "paragraphs").glob("*.md")):
            for number, body in parse(p).items():
                assert number not in paragraphs
                paragraphs[number] = converter.convert(to_text(body))
        assert set(paragraphs) == set(range(1, 163))
        (dest / "paragraphs.tsv").write_text(
            "# 簡體手札正文；可直接維護。保留 para key 與換行控制碼。\n" +
            "\n".join(f"para:{n}\t{escape(text)}" for n, text in sorted(paragraphs.items())) + "\n")
    print(f"Derived {len(list(dest.glob('*.tsv')))} TSV files and 162 paragraphs; review before packaging.")


if __name__ == "__main__":
    main()
