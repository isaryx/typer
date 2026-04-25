#!/usr/bin/env python3
"""
Regenerate `assets/passages.txt` with ~N public-domain passages (UTF-8), separated by blank
lines, for the passage-mode embed (`assets/embed.go`).

We intentionally source from a Project Gutenberg *dictionary-of-proverbs* text where each entry is
usually a short sentence. For typing "passages", we bundle consecutive entries into a single passage
until it lands in a configurable word-count window (default: 15–50 words).
"""

from __future__ import annotations

import hashlib
import re
import urllib.request
from dataclasses import dataclass
from typing import List


@dataclass(frozen=True)
class GutenbergPlainText:
    title: str
    url: str
    ebook_id: int


BOOK: GutenbergPlainText = GutenbergPlainText(
    title='Thomas Preston — "Dictionary of English Proverbs and Proverbial Phrases" (eBook #39281)',
    url="https://www.gutenberg.org/files/39281/39281-0.txt",
    ebook_id=39281,
)

# Examples:
#   402. DEEPEST WATER. In the deepest water is the best fishing.
#   407. DEPTH. Never venture out of your depth until you can swim.
RE_ENTRY = re.compile(
    r"^\s*(\d+)\.\s+([A-Z0-9][A-Z0-9 &]{0,50})\.\s+(.+)\s*$"
)

RE_BAD = re.compile(r"(?i)(gutenberg|pgdp\.net|www\.gutenberg|project gutenberg|_+[a-z]+_+)")
RE_BYLINE = re.compile(r"(?i)^by\s+[^.]*\b(M\.?A|D\.?D|LL\.?D|Rev|Professor|translated)\b")
RE_TOPIC = re.compile(r"^[A-Z0-9][A-Z0-9 &]{0,50}$")
RE_TOPIC_JUNK = re.compile(
    r"(?i)(price|familiar|quotations|series|handbook|gutenberg|copyright|ebook|license|publishing)"
)


def is_ok_topic(topic: str) -> bool:
    t = topic.strip()
    if not t:
        return False
    if not RE_TOPIC.match(t):
        return False
    if RE_TOPIC_JUNK.search(t):
        return False
    words = [w for w in t.split() if w]
    if not (1 <= len(words) <= 4):
        return False
    for w in words:
        if w in {"&"}:
            continue
        if not w.isupper():
            return False
    return True


def fetch(url: str) -> str:
    req = urllib.request.Request(
        url,
        headers={"User-Agent": "typer/1.0 (passage builder; open source)"},
    )
    with urllib.request.urlopen(req, timeout=180) as resp:  # noqa: S310
        raw = resp.read()
    return raw.decode("utf-8", errors="replace").lstrip("\ufeff")


def gutenberg_core(text: str) -> str:
    # Tolerate both the modern and older delimiter wording.
    m_start = re.search(
        r"(?m)^\*\*\*\s*START OF (?:THIS|THE) PROJECT GUTENBERG EBOOK.*\*\*\*",
        text,
    )
    m_end = re.search(r"(?m)^\*\*\*\s*END OF (?:THIS|THE) PROJECT GUTENBERG EBOOK.*\*\*\*", text)
    if not m_start or not m_end or m_start.start() >= m_end.start():
        raise RuntimeError("Could not find Gutenberg START/END markers; refusing to parse header/license text")
    return text[m_start.end() : m_end.start()]


def clean_proverb(s: str) -> str:
    s = s.replace("\r\n", "\n")
    s = s.replace("\u00a0", " ")
    s = re.sub(r"\s+", " ", s).strip()
    s = s.strip("“”\"'")
    return s


def count_words(s: str) -> int:
    return len(re.findall(r"\b[’A-Za-z0-9'-]+\b", s))


def is_ok_atomic_proverb(s: str) -> bool:
    """A single dictionary proverb line, before bundling into a longer passage."""
    t = clean_proverb(s)
    if not t:
        return False
    if not (12 <= len(t) <= 240):
        return False
    if RE_BAD.search(t):
        return False
    if RE_BYLINE.search(t):
        return False
    if t.startswith(("_", "*", "#", "-", "—", "–")):
        return False
    if t.isupper():
        return False
    if not t[0].isalpha() and t[:1] not in {'"'}:
        return False
    wc = count_words(t)
    # Keep atoms short so 2–3 can be bundled into a 15–50 word passage without overshooting 50.
    if not (3 <= wc <= 22):
        return False
    if t[-1] not in ".?!":
        return False
    return True


def join_two_proverbs(a: str, b: str) -> str:
    a = clean_proverb(a)
    b = clean_proverb(b)
    if a[-1] in ".?!":
        return f"{a} {b}"
    return f"{a}. {b}"


def join_passage(parts: List[str]) -> str:
    if not parts:
        return ""
    joined = parts[0]
    for nxt in parts[1:]:
        joined = join_two_proverbs(joined, nxt)
    return clean_proverb(joined)


def trim_buf_to_max(buf: List[str], min_words: int, max_words: int) -> None:
    while len(buf) > 1 and count_words(join_passage(buf)) > max_words:
        if count_words(join_passage(buf[:-1])) < min_words:
            break
        buf.pop()


def merge_short_tail(out: List[str], min_words: int, max_words: int) -> None:
    if not out or count_words(out[-1]) >= min_words or len(out) < 2:
        return
    tail = out.pop()
    prev = out.pop()
    merged = join_two_proverbs(prev, tail)
    if count_words(merged) <= max_words:
        out.append(merged)
        return
    out.append(prev)
    out.append(tail)


def bundle_atomic_proverbs(items: List[str], min_words: int, max_words: int) -> List[str]:
    atoms = [clean_proverb(x) for x in items if is_ok_atomic_proverb(x)]
    out: List[str] = []
    buf: List[str] = []

    def buf_wc() -> int:
        return count_words(join_passage(buf)) if buf else 0

    def emit_buf() -> None:
        nonlocal buf
        if not buf:
            return
        trim_buf_to_max(buf, min_words=min_words, max_words=max_words)
        out.append(join_passage(buf))
        buf = []

    for a in atoms:
        if not buf:
            buf = [a]
            continue

        # If we're still short of the minimum, keep growing even if it exceeds max_words.
        if buf_wc() < min_words:
            buf.append(a)
            continue

        trial_wc = count_words(join_passage(buf + [a]))
        if trial_wc <= max_words:
            buf.append(a)
            continue

        emit_buf()
        buf = [a]

    emit_buf()
    merge_short_tail(out, min_words=min_words, max_words=max_words)
    return out


def parse_preston_atomic_proverbs(core: str) -> List[str]:
    class _PrestonParser:
        def __init__(self) -> None:
            self.out: List[str] = []
            self.buf: str | None = None

        def flush(self) -> None:
            if not self.buf:
                return
            if is_ok_atomic_proverb(self.buf):
                self.out.append(clean_proverb(self.buf))
            self.buf = None

        def start_entry(self, m: re.Match[str]) -> None:
            if not is_ok_topic(m.group(2)):
                self.buf = None
                return
            self.buf = m.group(3).strip()

        def continue_entry(self, line: str) -> None:
            if self.buf is None:
                return
            nxt = line.strip()
            if not nxt:
                return
            if self.buf.endswith(("-", "--")):
                self.buf = self.buf[:-1].rstrip()
            self.buf = (self.buf + " " + nxt).strip()

    p = _PrestonParser()
    for line in (ln.rstrip() for ln in core.splitlines()):
        if not line.strip():
            p.flush()
            continue

        m = RE_ENTRY.match(line)
        if m:
            p.flush()
            p.start_entry(m)
            continue

        p.continue_entry(line)

    p.flush()
    return p.out


def pick_n(items: List[str], n: int, mix_key: str) -> List[str]:
    if len(items) < n:
        raise RuntimeError(f"not enough passages: need {n}, have {len(items)}")

    def k(s: str) -> str:
        return hashlib.sha256(f"{mix_key}::{s}".encode("utf-8")).hexdigest()

    return sorted(items, key=k)[:n]


def main() -> None:
    n = 200
    out_path = "assets/passages.txt"
    mix_key = "typer-passages-preston-39281-v2-bundled"
    min_words = 15
    max_words = 50

    text = fetch(BOOK.url)
    core = gutenberg_core(text)
    atomic = parse_preston_atomic_proverbs(core)
    bundled = bundle_atomic_proverbs(atomic, min_words=min_words, max_words=max_words)
    chosen = pick_n(bundled, n, mix_key=mix_key)
    with open(out_path, "w", encoding="utf-8", newline="\n") as f:
        f.write("\n\n".join(chosen) + "\n")

    print(
        f"Wrote {n} passages to {out_path} "
        f"(atomic lines: {len(atomic)}, bundled passages: {len(bundled)}, window: {min_words}-{max_words} words)"
    )


if __name__ == "__main__":
    main()
