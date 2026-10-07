#!/usr/bin/env python3

import json
import sys
from pathlib import Path
from urllib.parse import urlsplit


def load_metadata(path):
    if not path or not Path(path).is_file():
        return None
    return json.loads(Path(path).read_text(encoding="utf-8"))


def prefix_delta(previous, current):
    if previous is None:
        return "—", f"{current:,}", "—"
    delta = current - previous
    sign = "+" if delta > 0 else ""
    return f"{previous:,}", f"{current:,}", f"{sign}{delta:,}"


def source_label(source):
    group = source.get("group", "unknown")
    url = source.get("url", "")
    parsed = urlsplit(url)
    target = f"{parsed.netloc}{parsed.path}" if parsed.netloc else url
    return f"[{group} — {target}]({url})" if url else group


def source_rows(previous_sources, current_sources):
    previous_by_key = {
        (source.get("group", ""), source.get("url", "")): source
        for source in previous_sources
    }
    current_by_key = {
        (source.get("group", ""), source.get("url", "")): source
        for source in current_sources
    }
    rows = []

    for key in sorted(previous_by_key.keys() | current_by_key.keys()):
        old = previous_by_key.get(key)
        new = current_by_key.get(key)
        if old is None:
            status = "Added"
        elif new is None:
            status = "Removed"
        elif any(
            old.get(field) != new.get(field)
            for field in (
                "type",
                "sha256",
                "ipv4_count",
                "ipv6_count",
                "rejected_cidr_count",
            )
        ):
            status = "Changed"
        else:
            continue

        old_v4 = old.get("ipv4_count") if old else None
        new_v4 = new.get("ipv4_count") if new else 0
        old_v6 = old.get("ipv6_count") if old else None
        new_v6 = new.get("ipv6_count") if new else 0
        old_rejected = old.get("rejected_cidr_count") if old else None
        new_rejected = new.get("rejected_cidr_count", 0) if new else 0
        v4_before, v4_after, v4_delta = prefix_delta(old_v4, new_v4)
        v6_before, v6_after, v6_delta = prefix_delta(old_v6, new_v6)
        rejected_before, rejected_after, rejected_delta = prefix_delta(old_rejected, new_rejected)

        old_hash = (old or {}).get("sha256", "")
        new_hash = (new or {}).get("sha256", "")
        hash_change = ""
        if old_hash != new_hash:
            hash_change = f"{old_hash[:12] or '—'} → {new_hash[:12] or '—'}"

        source = new or old
        rows.append(
            "| {source} | {status} | {v4_before} → {v4_after} ({v4_delta}) "
            "| {v6_before} → {v6_after} ({v6_delta}) "
            "| {rejected_before} → {rejected_after} ({rejected_delta}) | {hash_change} |".format(
                source=source_label(source),
                status=status,
                v4_before=v4_before,
                v4_after=v4_after,
                v4_delta=v4_delta,
                v6_before=v6_before,
                v6_after=v6_after,
                v6_delta=v6_delta,
                rejected_before=rejected_before,
                rejected_after=rejected_after,
                rejected_delta=rejected_delta,
                hash_change=hash_change,
            )
        )
    return rows


def main():
    previous_path = sys.argv[1] if len(sys.argv) > 1 else None
    current_path = sys.argv[2]
    previous = load_metadata(previous_path)
    current = load_metadata(current_path)

    print("## CIDR data diff\n")
    previous_hash = (previous or {}).get("content_hash", "")
    current_hash = current.get("content_hash", "")
    print(f"Content hash: `{previous_hash or 'no previous release'}` → `{current_hash}`\n")

    print("### Aggregate prefix counts\n")
    print("| Family | Previous | Current | Δ |\n| --- | ---: | ---: | ---: |")
    for field, label in (("ipv4_count", "IPv4"), ("ipv6_count", "IPv6")):
        old_count = (previous or {}).get(field)
        new_count = current.get(field, 0)
        before, after, delta = prefix_delta(old_count, new_count)
        print(f"| {label} | {before} | {after} | {delta} |")

    old_sources = (previous or {}).get("sources")
    new_sources = current.get("sources", [])
    print("\n### Source changes\n")
    if old_sources is None:
        print("No per-source records are available in the previous release; current source status is the baseline.\n")
        print("| Source | Type | IPv4 | IPv6 | Rejected CIDRs | SHA-256 |\n| --- | --- | ---: | ---: | ---: | --- |")
        for source in new_sources:
            print(
                f"| {source_label(source)} | {source.get('type', '')} | "
                f"{source.get('ipv4_count', 0):,} | {source.get('ipv6_count', 0):,} | "
                f"{source.get('rejected_cidr_count', 0):,} | "
                f"{source.get('sha256', '')[:12]} |"
            )
        return

    rows = source_rows(old_sources, new_sources)
    if not rows:
        print("No configured source records changed.\n")
        return
    print("| Source | Status | IPv4 prefixes | IPv6 prefixes | Rejected CIDRs | SHA-256 |\n| --- | --- | ---: | ---: | ---: | --- |")
    print("\n".join(rows))


if __name__ == "__main__":
    main()
