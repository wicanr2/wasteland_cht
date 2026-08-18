#!/usr/bin/env python3
"""把 docs/walkthrough/swm-005/ 的 markdown 產生成自包含的靜態 HTML。

只用標準函式庫，不需要任何套件、不需要 docker。用法：

    python3 site/build.py

產出寫到 site/ 底下，每份 markdown 對應一個 .html，另加一個 index.html。
CSS 內嵌在每一頁裡，不引用任何外部資源。
"""

import html
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
SRC = ROOT / "docs" / "walkthrough" / "swm-005"
OUT = Path(__file__).resolve().parent

# 頁面順序與顯示標題。index 由 README 產生，其餘照連載順序。
PAGES = [
    ("README.md", "index.html", "軟體世界第 5 期《沙漠傳奇》"),
    ("p56.md", "p56.html", "p.56 藏匿處的激戰"),
    ("p57.md", "p57.html", "p.57 Needles 城中心區"),
    ("p58.md", "p58.html", "p.58 廢坑與血神殿"),
    ("p59.md", "p59.html", "p.59 血池、血牧師與抵達 Vegas"),
    ("p60.md", "p60.html", "p.60 蠍式殺人機器與蕈狀雲神殿"),
    ("p61.md", "p61.html", "p.61 下水道之旅與 Max 的組裝"),
    ("glossary.md", "glossary.html", "用語與譯名"),
    ("mechanics-claims.md", "mechanics-claims.html", "機制斷言與逆向對照"),
]

# markdown 檔名 → 產出的 html 檔名，用來改寫內部連結
LINK_MAP = {md: out for md, out, _ in PAGES}

CSS = """
:root {
  --bg: #fbfaf7;
  --fg: #22201c;
  --muted: #6b665e;
  --rule: #ddd8cd;
  --accent: #8a4b2a;
  --card: #ffffff;
  --code-bg: #f1eee7;
}
@media (prefers-color-scheme: dark) {
  :root {
    --bg: #171614;
    --fg: #e6e2da;
    --muted: #9a948a;
    --rule: #34312c;
    --accent: #d99a6e;
    --card: #1e1d1a;
    --code-bg: #26241f;
  }
}
* { box-sizing: border-box; }
html { -webkit-text-size-adjust: 100%; }
body {
  margin: 0;
  background: var(--bg);
  color: var(--fg);
  font-family: "Noto Serif CJK TC", "Source Han Serif TC", "PMingLiU",
               "Songti TC", Georgia, "Times New Roman", serif;
  font-size: 17px;
  line-height: 1.9;
}
.wrap { max-width: 40em; margin: 0 auto; padding: 2.5rem 1.25rem 5rem; }
header.site {
  border-bottom: 1px solid var(--rule);
  margin-bottom: 2.5rem;
  padding-bottom: 1.25rem;
}
header.site .kicker {
  font-size: 0.8rem;
  letter-spacing: 0.18em;
  color: var(--muted);
  text-transform: uppercase;
  margin: 0 0 0.4rem;
}
header.site a { color: inherit; text-decoration: none; }
header.site .title { font-size: 1.15rem; margin: 0; font-weight: 600; }
h1, h2, h3 { line-height: 1.45; font-weight: 600; }
h1 { font-size: 1.65rem; margin: 0 0 1.5rem; }
h2 {
  font-size: 1.25rem;
  margin: 2.75rem 0 1rem;
  padding-bottom: 0.35rem;
  border-bottom: 1px solid var(--rule);
}
h3 { font-size: 1.05rem; margin: 2rem 0 0.75rem; color: var(--accent); }
p { margin: 0 0 1.15rem; }
a { color: var(--accent); text-underline-offset: 0.2em; }
ul, ol { margin: 0 0 1.15rem; padding-left: 1.5rem; }
li { margin-bottom: 0.4rem; }
blockquote {
  margin: 1.5rem 0;
  padding: 0.6rem 0 0.6rem 1.15rem;
  border-left: 3px solid var(--accent);
  color: var(--muted);
  font-style: normal;
}
blockquote p:last-child { margin-bottom: 0; }
code {
  background: var(--code-bg);
  padding: 0.12em 0.38em;
  border-radius: 3px;
  font-family: "DejaVu Sans Mono", "Menlo", "Consolas", monospace;
  font-size: 0.86em;
}
pre {
  background: var(--code-bg);
  padding: 1rem;
  border-radius: 4px;
  overflow-x: auto;
}
pre code { background: none; padding: 0; }
.table-wrap { overflow-x: auto; margin: 0 0 1.5rem; }
table { border-collapse: collapse; width: 100%; font-size: 0.92rem; }
th, td {
  border: 1px solid var(--rule);
  padding: 0.5rem 0.7rem;
  text-align: left;
  vertical-align: top;
}
th { background: var(--code-bg); font-weight: 600; }
hr { border: 0; border-top: 1px solid var(--rule); margin: 2.5rem 0; }
nav.pager {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  margin-top: 3.5rem;
  padding-top: 1.25rem;
  border-top: 1px solid var(--rule);
  font-size: 0.9rem;
}
nav.pager a { text-decoration: none; }
nav.pager .spacer { flex: 1; }
footer.site {
  margin-top: 3.5rem;
  padding-top: 1.25rem;
  border-top: 1px solid var(--rule);
  color: var(--muted);
  font-size: 0.82rem;
}
footer.site p { margin: 0 0 0.5rem; }
@media (max-width: 600px) {
  body { font-size: 16px; line-height: 1.85; }
  .wrap { padding: 1.75rem 1rem 3.5rem; }
}
"""

PAGE = """<!DOCTYPE html>
<html lang="zh-Hant">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{title}</title>
<style>{css}</style>
</head>
<body>
<div class="wrap">
<header class="site">
<p class="kicker">軟體世界 第 5 期 ‧ 攻略</p>
<p class="title"><a href="index.html">沙漠傳奇——荒野遊俠的故事</a></p>
</header>
<main>
{body}
</main>
{pager}
<footer class="site">
<p>《軟體世界》第 5 期，Y.M.J. 著。本頁為文字整理，掃描件不隨本站散布。</p>
<p>荒野遊俠繁體中文化專案的一部分。</p>
</footer>
</div>
</body>
</html>
"""


def inline(text: str) -> str:
    """處理行內語法：先抽出 code span，再處理連結、粗體。"""
    spans: list[str] = []

    def stash(m: re.Match) -> str:
        spans.append(m.group(1))
        return f"\x00{len(spans) - 1}\x00"

    text = re.sub(r"`([^`]+)`", stash, text)
    text = html.escape(text, quote=False)

    def link(m: re.Match) -> str:
        label, href = m.group(1), m.group(2)
        # 內部 markdown 連結改寫成對應的 html；指到 repo 其他地方的連結拿掉。
        if href in LINK_MAP:
            href = LINK_MAP[href]
        elif href.startswith(("http://", "https://", "#")):
            pass
        elif href.endswith(".md") or "/" in href:
            return label  # 站上沒有這個檔，只留文字
        return f'<a href="{html.escape(href, quote=True)}">{label}</a>'

    text = re.sub(r"\[([^\]]+)\]\(([^)]+)\)", link, text)
    text = re.sub(r"\*\*([^*]+)\*\*", r"<strong>\1</strong>", text)

    def pop(m: re.Match) -> str:
        return f"<code>{html.escape(spans[int(m.group(1))], quote=False)}</code>"

    return re.sub(r"\x00(\d+)\x00", pop, text)


def is_table_sep(line: str) -> bool:
    return bool(re.fullmatch(r"\|[\s:|-]+\|", line.strip()))


def cells(line: str) -> list[str]:
    return [c.strip() for c in line.strip().strip("|").split("|")]


def render(md: str) -> str:
    lines = md.split("\n")
    out: list[str] = []
    i = 0
    while i < len(lines):
        line = lines[i]
        stripped = line.strip()

        if not stripped:
            i += 1
            continue

        # 表格：需要「表頭 + 分隔列」兩行才算
        if (
            stripped.startswith("|")
            and i + 1 < len(lines)
            and is_table_sep(lines[i + 1])
        ):
            head = cells(line)
            i += 2
            rows = []
            while i < len(lines) and lines[i].strip().startswith("|"):
                rows.append(cells(lines[i]))
                i += 1
            out.append('<div class="table-wrap"><table>')
            out.append(
                "<thead><tr>"
                + "".join(f"<th>{inline(c)}</th>" for c in head)
                + "</tr></thead>"
            )
            out.append("<tbody>")
            for row in rows:
                out.append(
                    "<tr>" + "".join(f"<td>{inline(c)}</td>" for c in row) + "</tr>"
                )
            out.append("</tbody></table></div>")
            continue

        # 圍籬程式碼區塊
        if stripped.startswith("```"):
            i += 1
            buf = []
            while i < len(lines) and not lines[i].strip().startswith("```"):
                buf.append(lines[i])
                i += 1
            i += 1
            body = html.escape("\n".join(buf), quote=False)
            out.append(f"<pre><code>{body}</code></pre>")
            continue

        # 標題
        m = re.match(r"^(#{1,6})\s+(.*)$", stripped)
        if m:
            level = len(m.group(1))
            out.append(f"<h{level}>{inline(m.group(2))}</h{level}>")
            i += 1
            continue

        # 水平線
        if re.fullmatch(r"-{3,}", stripped):
            out.append("<hr>")
            i += 1
            continue

        # 引用
        if stripped.startswith(">"):
            buf = []
            while i < len(lines) and lines[i].strip().startswith(">"):
                buf.append(lines[i].strip().lstrip(">").strip())
                i += 1
            text = " ".join(x for x in buf if x)
            out.append(f"<blockquote><p>{inline(text)}</p></blockquote>")
            continue

        # 清單
        if re.match(r"^[-*]\s+", stripped) or re.match(r"^\d+\.\s+", stripped):
            ordered = bool(re.match(r"^\d+\.\s+", stripped))
            tag = "ol" if ordered else "ul"
            out.append(f"<{tag}>")
            while i < len(lines):
                s = lines[i].strip()
                m2 = re.match(r"^(?:[-*]|\d+\.)\s+(.*)$", s)
                if not m2:
                    break
                out.append(f"<li>{inline(m2.group(1))}</li>")
                i += 1
            out.append(f"</{tag}>")
            continue

        # 段落：吃到空行為止
        buf = []
        while i < len(lines) and lines[i].strip():
            s = lines[i].strip()
            if s.startswith(("#", ">", "|", "```")) or re.match(
                r"^(?:[-*]\s+|\d+\.\s+|-{3,}$)", s
            ):
                break
            buf.append(s)
            i += 1
        if buf:
            out.append(f"<p>{inline(' '.join(buf))}</p>")
        else:
            i += 1

    return "\n".join(out)


def pager(idx: int) -> str:
    prev_link = next_link = ""
    if idx > 0:
        _, href, label = PAGES[idx - 1]
        prev_link = f'<a href="{href}">← {html.escape(label)}</a>'
    if idx < len(PAGES) - 1:
        _, href, label = PAGES[idx + 1]
        next_link = f'<a href="{href}">{html.escape(label)} →</a>'
    if not prev_link and not next_link:
        return ""
    return (
        '<nav class="pager">'
        f"{prev_link}<span class=\"spacer\"></span>{next_link}"
        "</nav>"
    )


def main() -> int:
    if not SRC.is_dir():
        print(f"找不到來源目錄：{SRC}", file=sys.stderr)
        return 1

    written = []
    for idx, (name, outname, label) in enumerate(PAGES):
        path = SRC / name
        if not path.is_file():
            print(f"略過（不存在）：{path}", file=sys.stderr)
            continue
        body = render(path.read_text(encoding="utf-8"))
        title = "沙漠傳奇——荒野遊俠的故事"
        if name != "README.md":
            title = f"{label}｜{title}"
        page = PAGE.format(title=html.escape(title), css=CSS, body=body, pager=pager(idx))
        (OUT / outname).write_text(page, encoding="utf-8")
        written.append(outname)

    print(f"產出 {len(written)} 頁：{', '.join(written)}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
