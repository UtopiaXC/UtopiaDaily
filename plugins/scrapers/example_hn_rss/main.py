import datetime
import email.utils
import json
import sys
import urllib.request
import xml.etree.ElementTree as ET


FEED_URL = "https://news.ycombinator.com/rss"
HEADERS = {"User-Agent": "Mozilla/5.0 (compatible; UtopiaDaily/0.2.0)"}


def fetch_rss():
    req = urllib.request.Request(FEED_URL, headers=HEADERS)
    with urllib.request.urlopen(req, timeout=15) as resp:
        return resp.read()


def parse_pub_date(raw_date):
    if not raw_date:
        return datetime.datetime.now(datetime.timezone.utc).isoformat()
    try:
        dt = email.utils.parsedate_to_datetime(raw_date)
        if dt.tzinfo is None:
            dt = dt.replace(tzinfo=datetime.timezone.utc)
        return dt.astimezone(datetime.timezone.utc).isoformat()
    except Exception:
        return datetime.datetime.now(datetime.timezone.utc).isoformat()


def run():
    try:
        input_data = json.load(sys.stdin)
    except Exception as exc:
        sys.stderr.write(f"Failed to read stdin JSON: {exc}\n")
        sys.stdout.write(json.dumps({"status": "failed", "message": str(exc)}))
        sys.exit(1)

    task_id = input_data.get("task_id", "")
    config = input_data.get("config", {})

    if task_id == "test":
        try:
            fetch_rss()
            sys.stdout.write(json.dumps({"status": "success", "message": "ok"}))
            sys.exit(0)
        except Exception as exc:
            sys.stderr.write(f"Test failed: {exc}\n")
            sys.stdout.write(json.dumps({"status": "failed", "message": str(exc)}))
            sys.exit(0)

    try:
        limit = int(config.get("limit", 10))
    except Exception:
        limit = 10
    if limit <= 0:
        limit = 10

    try:
        raw_xml = fetch_rss()
        root = ET.fromstring(raw_xml)
    except Exception as exc:
        sys.stderr.write(f"Fetch RSS failed: {exc}\n")
        sys.stdout.write(json.dumps({"status": "failed", "message": str(exc)}))
        sys.exit(1)

    results = []
    for item in root.findall("./channel/item")[:limit]:
        title = (item.findtext("title") or "").strip()
        link = (item.findtext("link") or "").strip()
        comments = (item.findtext("comments") or "").strip()
        pub_date = parse_pub_date((item.findtext("pubDate") or "").strip())

        lines = []
        if comments:
            lines.append(f"[Discussion]({comments})")
        content = "\n\n".join(lines) if lines else title

        results.append({
            "title": title,
            "summary": title,
            "content": content,
            "content_type": "markdown",
            "render_type": "raw",
            "source_name": "Hacker News",
            "external_url": link or comments,
            "published_at": pub_date,
        })

    sys.stderr.write(f"Collected {len(results)} Hacker News stories\n")
    sys.stdout.write(json.dumps({"status": "success", "data": results}, ensure_ascii=False))
    sys.exit(0)


if __name__ == "__main__":
    run()
