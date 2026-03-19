#!/usr/bin/env python3
"""Generate a markdown document from watch_later_details.json."""

import json
import re
import os
from collections import Counter, defaultdict
from datetime import datetime

INPUT = os.path.join(os.path.dirname(__file__), "..", "data", "youtube", "watch_later_details.json")
OUTPUT = os.path.join(os.path.dirname(__file__), "..", "data", "youtube", "watch_later_catalog.md")


def detect_lang(title):
    tamil = len(re.findall(r"[\u0B80-\u0BFF]", title))
    hindi = len(re.findall(r"[\u0900-\u097F]", title))
    if tamil > 3:
        return "Tamil"
    if hindi > 3:
        return "Hindi"
    return "English"


def categorize(title, channel):
    t = (title + " " + channel).lower()
    rules = [
        ("AI/ML", ["ai ", "llm", "gpt", "claude", "gemini", "machine learning", "deep learning",
                    "neural", "openai", "anthropic", "langchain", "rag ", "agentic", "agent",
                    "transformer", "chatgpt", "copilot", "diffusion", "embedding", "vector",
                    "fine-tun", "finetun", "prompt", "inference", "hugging", "lora", "cursor",
                    "coding agent", "vibe cod", "mcp ", "crewai", "autogen", "worldofai",
                    "fahd mirza", "aicodek"]),
        ("Software/Tech", ["kubernetes", "docker", "microservice", "api ", "cloud", "devops",
                           "aws ", "azure", "database", "sql", "rust ", "golang", "go ",
                           "python", "java", "react", "system design", "architect", "distribut",
                           "kafka", "redis", "grpc", "graphql", "observab", "monitor",
                           "bytebytego", "gaurav sen", "telcoma", "infra", "platform eng",
                           "software", "engineer", "program", "coding", "code review",
                           "tech field", "infoq", "thoughtworks"]),
        ("Finance/Investing", ["stock", "market", "invest", "mutual fund", "portfolio",
                               "trading", "nifty", "sensex", "wealth", "money pechu",
                               "finance", "ppfas", "sip ", "smallcap", "midcap", "largecap",
                               "equity", "debt fund", "prakala", "muthaleetuk"]),
        ("Astrology", ["astro", "jothid", "rasi", "zodiac", "horoscope", "nakshatra",
                       "sitthar", "raasi", "palangal", "jathag", "brindha", "raghavan"]),
        ("Startups/Business", ["startup", "founder", "y combinator", "venture",
                               "entrepreneur", "business", "saas", "growth hack"]),
        ("Science/Learning", ["physics", "math", "science", "biology", "chemistry",
                              "quantum", "big think", "ted ", "tedx", "freecodecamp",
                              "education", "lecture", "university"]),
        ("Health/Fitness", ["health", "workout", "exercise", "yoga", "meditat", "diet",
                           "nutrition", "doctor", "medical"]),
        ("Career/Productivity", ["career", "interview", "resume", "productiv", "habit",
                                 "tim ferriss", "skill", "learn", "study"]),
    ]
    for cat, keywords in rules:
        for kw in keywords:
            if kw in t:
                return cat
    return "Other"


def main():
    with open(INPUT) as f:
        data = json.load(f)

    ok = [d for d in data if d["status"] == "ok"]
    unavailable = [d for d in data if d["status"] != "ok"]

    # Enrich each entry
    for d in ok:
        d["language"] = detect_lang(d["title"])
        d["category"] = categorize(d["title"], d["channel"])
        d["added_date"] = d["added_at"][:10]
        d["added_month"] = d["added_at"][:7]

    # Stats
    total = len(data)
    ok_count = len(ok)
    lang_counts = Counter(d["language"] for d in ok)
    cat_counts = Counter(d["category"] for d in ok)
    channel_counts = Counter(d["channel"] for d in ok)
    months = defaultdict(int)
    for d in ok:
        months[d["added_month"]] += 1

    # Group by category then by month
    by_cat = defaultdict(list)
    for d in ok:
        by_cat[d["category"]].append(d)

    lines = []
    w = lines.append

    w("# YouTube Watch Later Catalog")
    w("")
    w(f"> Auto-generated from {total:,} videos in Watch Later playlist")
    w(f"> Date range: {ok[- 1]['added_date']} → {ok[0]['added_date']}")
    w(f"> Successfully resolved: {ok_count:,} | Unavailable: {len(unavailable)}")
    w("")

    # Summary stats
    w("## 📊 Summary")
    w("")
    w("### By Language")
    w("")
    w("| Language | Count | % |")
    w("|----------|------:|--:|")
    for lang, cnt in lang_counts.most_common():
        w(f"| {lang} | {cnt:,} | {cnt*100//ok_count}% |")
    w("")

    w("### By Category")
    w("")
    w("| Category | Count | % |")
    w("|----------|------:|--:|")
    for cat, cnt in sorted(cat_counts.items(), key=lambda x: -x[1]):
        w(f"| {cat} | {cnt:,} | {cnt*100//ok_count}% |")
    w("")

    w("### Top 50 Channels")
    w("")
    w("| # | Channel | Videos |")
    w("|--:|---------|-------:|")
    for i, (ch, cnt) in enumerate(channel_counts.most_common(50), 1):
        w(f"| {i} | {ch} | {cnt} |")
    w("")

    w("### Monthly Activity (Last 12 Months)")
    w("")
    w("| Month | Videos |")
    w("|-------|-------:|")
    for m in sorted(months.keys(), reverse=True)[:12]:
        w(f"| {m} | {months[m]} |")
    w("")

    # Full catalog by category
    w("---")
    w("")
    w("## 📚 Full Catalog")
    w("")

    cat_order = sorted(by_cat.keys(), key=lambda c: -len(by_cat[c]))
    for cat in cat_order:
        videos = by_cat[cat]
        w(f"### {cat} ({len(videos):,} videos)")
        w("")
        w("| # | Title | Channel | Language | Added | Link |")
        w("|--:|-------|---------|----------|-------|------|")
        for i, d in enumerate(videos, 1):
            title = d["title"].replace("|", "\\|")
            channel = d["channel"].replace("|", "\\|")
            link = f"[▶]({d['url']})"
            w(f"| {i} | {title} | {channel} | {d['language']} | {d['added_date']} | {link} |")
        w("")

    # Unavailable videos
    if unavailable:
        w("---")
        w("")
        w(f"## ⚠️ Unavailable Videos ({len(unavailable)})")
        w("")
        w("| # | Video ID | Added | Status | Link |")
        w("|--:|----------|-------|--------|------|")
        for i, d in enumerate(unavailable, 1):
            link = f"[▶]({d['url']})"
            w(f"| {i} | `{d['video_id']}` | {d['added_at'][:10]} | {d['status']} | {link} |")
        w("")

    os.makedirs(os.path.dirname(OUTPUT), exist_ok=True)
    with open(OUTPUT, "w", encoding="utf-8") as f:
        f.write("\n".join(lines))

    print(f"Generated {OUTPUT} ({len(lines)} lines)")


if __name__ == "__main__":
    main()
