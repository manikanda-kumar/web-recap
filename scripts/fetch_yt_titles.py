#!/usr/bin/env python3
"""Fetch YouTube video titles and channel names via oEmbed API."""

import csv
import json
import sys
import os
import time
from concurrent.futures import ThreadPoolExecutor, as_completed
from urllib.request import urlopen, Request
from urllib.error import HTTPError, URLError

INPUT_CSV = os.path.join(os.path.dirname(__file__), "..", "data", "watch_later_videos.csv")
OUTPUT_JSON = os.path.join(os.path.dirname(__file__), "..", "data", "watch_later_details.json")
PROGRESS_FILE = os.path.join(os.path.dirname(__file__), "..", "tmp", "yt_fetch_progress.json")
MAX_WORKERS = 20
OEMBED_URL = "https://www.youtube.com/oembed?url=https://www.youtube.com/watch?v={vid}&format=json"


def fetch_one(vid: str, timestamp: str) -> dict:
    """Fetch oEmbed data for a single video ID."""
    url = OEMBED_URL.format(vid=vid.strip())
    try:
        req = Request(url, headers={"User-Agent": "Mozilla/5.0"})
        with urlopen(req, timeout=10) as resp:
            data = json.loads(resp.read().decode())
            return {
                "video_id": vid.strip(),
                "title": data.get("title", ""),
                "channel": data.get("author_name", ""),
                "channel_url": data.get("author_url", ""),
                "added_at": timestamp,
                "url": f"https://www.youtube.com/watch?v={vid.strip()}",
                "status": "ok",
            }
    except HTTPError as e:
        status = "private_or_deleted" if e.code in (401, 403, 404) else f"http_{e.code}"
        return {
            "video_id": vid.strip(),
            "title": "",
            "channel": "",
            "channel_url": "",
            "added_at": timestamp,
            "url": f"https://www.youtube.com/watch?v={vid.strip()}",
            "status": status,
        }
    except (URLError, TimeoutError, Exception) as e:
        return {
            "video_id": vid.strip(),
            "title": "",
            "channel": "",
            "channel_url": "",
            "added_at": timestamp,
            "url": f"https://www.youtube.com/watch?v={vid.strip()}",
            "status": f"error: {str(e)[:80]}",
        }


def load_progress() -> dict:
    """Load already-fetched results to support resuming."""
    if os.path.exists(PROGRESS_FILE):
        with open(PROGRESS_FILE, "r") as f:
            data = json.load(f)
            return {item["video_id"]: item for item in data}
    return {}


def save_progress(results: list):
    os.makedirs(os.path.dirname(PROGRESS_FILE), exist_ok=True)
    with open(PROGRESS_FILE, "w") as f:
        json.dump(results, f)


def main():
    # Read CSV
    videos = []
    with open(INPUT_CSV, "r") as f:
        reader = csv.reader(f)
        next(reader)  # skip header
        for row in reader:
            if len(row) >= 2:
                vid = row[0].strip()
                ts = row[1].strip()
                if vid:
                    videos.append((vid, ts))

    print(f"Total videos in CSV: {len(videos)}")

    # Load progress
    done = load_progress()
    print(f"Already fetched: {len(done)}")

    # Filter remaining
    remaining = [(vid, ts) for vid, ts in videos if vid not in done]
    print(f"Remaining to fetch: {len(remaining)}")

    if not remaining:
        print("All done! Generating output...")
    else:
        # Fetch in batches
        results_list = list(done.values())
        batch_size = 100
        for i in range(0, len(remaining), batch_size):
            batch = remaining[i : i + batch_size]
            with ThreadPoolExecutor(max_workers=MAX_WORKERS) as executor:
                futures = {
                    executor.submit(fetch_one, vid, ts): vid for vid, ts in batch
                }
                for future in as_completed(futures):
                    result = future.result()
                    results_list.append(result)
                    done[result["video_id"]] = result

            fetched_so_far = len(done)
            ok_count = sum(1 for r in done.values() if r["status"] == "ok")
            print(
                f"  Progress: {fetched_so_far}/{len(videos)} "
                f"(ok: {ok_count}, errors: {fetched_so_far - ok_count})"
            )
            save_progress(results_list)
            time.sleep(0.2)  # small delay between batches

    # Final output - ordered by original CSV order
    all_results = []
    for vid, ts in videos:
        if vid in done:
            all_results.append(done[vid])

    # Save JSON
    with open(OUTPUT_JSON, "w") as f:
        json.dump(all_results, f, indent=2, ensure_ascii=False)
    print(f"\nSaved {len(all_results)} entries to {OUTPUT_JSON}")

    # Stats
    ok = sum(1 for r in all_results if r["status"] == "ok")
    err = len(all_results) - ok
    print(f"OK: {ok}, Unavailable/Error: {err}")


if __name__ == "__main__":
    main()
