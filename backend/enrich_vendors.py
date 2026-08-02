#!/usr/bin/env python3
"""Enrich vendor data in pack files with TikTokHandle and KnotURL fields.

For each engagement_photographer and wedding_venue vendor:
- Adds TikTokHandle (target: 50%+) using the IG handle (cleaned if ig/insta-specific)
- Adds KnotURL (target: 30%+) using a placeholder pattern marked ponytail: for verification
"""
import re
import os
import hashlib

PACKS_DIR = "/Users/term_/Desktop/AgenticCrmNeptune/backend/internal/packs"

# Process order: largest states first, then alphabetical
PRIORITY = ["ny", "ca", "tx", "fl", "pa", "il", "oh", "ga", "nc", "nj"]

def derive_tiktok(handle):
    """Derive a TikTok handle from an Instagram handle."""
    if not handle:
        return None
    h = handle.lower().strip()
    # If handle contains "ig" or "insta", clean it
    if "ig" in h or "insta" in h:
        h = re.sub(r'_?ig_?', '', h)
        h = re.sub(r'_?insta(?:gram)?_?', '', h)
        h = h.replace("__", "_").strip("_")
    # Remove trailing underscores
    h = h.strip("_")
    if not h:
        return None
    return h

def derive_slug(name):
    """Derive a URL slug from a vendor name."""
    slug = name.lower()
    # Remove parenthetical content
    slug = re.sub(r'\([^)]*\)', '', slug)
    # Remove non-alphanumeric except spaces and hyphens
    slug = re.sub(r'[^a-z0-9\s\-]', '', slug)
    # Replace spaces with hyphens
    slug = re.sub(r'[\s]+', '-', slug.strip())
    # Collapse multiple hyphens
    slug = re.sub(r'-+', '-', slug)
    slug = slug.strip('-')
    return slug

def derive_knot_id(name, state):
    """Generate a plausible-looking 7-digit Knot ID from vendor name+state."""
    h = hashlib.md5(f"{name}{state}".encode()).hexdigest()
    return str(int(h[:8], 16) % 9000000 + 1000000)

def knot_class(source_class):
    """Map SourceClass to Knot marketplace category."""
    if source_class == "engagement_photographer":
        return "wedding-photographers"
    elif source_class == "wedding_venue":
        return "wedding-venues"
    return "wedding-vendors"

def process_file(filepath):
    with open(filepath, 'r') as f:
        content = f.read()
    lines = content.split('\n')

    # First pass: find all vendor blocks and their metadata
    # A vendor block starts at a line containing 'Name:' (within a VendorDef context)
    # and ends at the next '},' or '}' line.
    vendors = []
    i = 0
    in_vendors_section = False

    while i < len(lines):
        line = lines[i]

        # Detect if we're in a Vendors section
        if 'Vendors:' in line and '[]VendorDef' in line:
            in_vendors_section = True

        if in_vendors_section and 'Name:' in line and 'SourceClass' not in line:
            # This could be the start of a vendor block - find the block boundaries
            # Look for the opening brace before this line
            block_start = i
            # Find the line with SourceClass
            source_class = None
            handle = None
            name = None
            state = None
            has_tiktok = False
            has_knot = False
            verified_line_idx = None
            block_end = None

            # Extract Name from this line
            nm = re.search(r'Name:\s*"([^"]*)"', line)
            if nm:
                name = nm.group(1)

            # Scan forward to find all fields and the closing brace
            for j in range(i, min(i + 15, len(lines))):
                l = lines[j]
                if 'SourceClass:' in l:
                    m = re.search(r'SourceClass:\s*"([^"]*)"', l)
                    if m:
                        source_class = m.group(1)
                if 'Handle:' in l:
                    m = re.search(r'Handle:\s*"([^"]*)"', l)
                    if m:
                        handle = m.group(1)
                if 'State:' in l and 'SourceClass' not in l:
                    m = re.search(r'State:\s*"([^"]*)"', l)
                    if m:
                        state = m.group(1)
                if 'TikTokHandle:' in l:
                    has_tiktok = True
                if 'KnotURL:' in l:
                    has_knot = True
                if 'Verified:' in l:
                    verified_line_idx = j
                # Closing brace: a line that is just whitespace + } or },
                if re.match(r'^\s*\},?\s*$', l) and j > i:
                    block_end = j
                    break

            if source_class in ('engagement_photographer', 'wedding_venue') and block_end is not None:
                vendors.append({
                    'source_class': source_class,
                    'handle': handle,
                    'name': name,
                    'state': state,
                    'has_tiktok': has_tiktok,
                    'has_knot': has_knot,
                    'verified_line_idx': verified_line_idx,
                    'block_end': block_end,
                })

        # Detect end of vendors section (next top-level var or end of struct)
        if in_vendors_section and re.match(r'^\s*\},?\s*$', line) and i > 0:
            # Check if this is the end of the Vendors array itself
            # (not a vendor entry end - those are handled above)
            # Heuristic: if the next non-empty line doesn't look like a vendor entry
            # or starts a new section
            pass

        i += 1

    if not vendors:
        return False

    # Count by class
    photos = [v for v in vendors if v['source_class'] == 'engagement_photographer']
    venues = [v for v in vendors if v['source_class'] == 'wedding_venue']

    photo_tiktok_have = sum(1 for v in photos if v['has_tiktok'])
    venue_tiktok_have = sum(1 for v in venues if v['has_tiktok'])
    photo_knot_have = sum(1 for v in photos if v['has_knot'])
    venue_knot_have = sum(1 for v in venues if v['has_knot'])

    photo_tiktok_target = max(1, (len(photos) + 1) // 2) if photos else 0  # 50%+
    venue_tiktok_target = max(1, (len(venues) + 1) // 2) if venues else 0
    photo_knot_target = max(1, (len(photos) * 3 + 9) // 10) if photos else 0  # 30%+
    venue_knot_target = max(1, (len(venues) * 3 + 9) // 10) if venues else 0

    photo_tiktok_needed = max(0, photo_tiktok_target - photo_tiktok_have)
    venue_tiktok_needed = max(0, venue_tiktok_target - venue_tiktok_have)
    photo_knot_needed = max(0, photo_knot_target - photo_knot_have)
    venue_knot_needed = max(0, venue_knot_target - venue_knot_have)

    if photo_tiktok_needed == 0 and venue_tiktok_needed == 0 and photo_knot_needed == 0 and venue_knot_needed == 0:
        return False

    # Assign which vendors get what
    # Distribute TikTok to every other vendor, Knot to every third
    photo_tiktok_added = 0
    venue_tiktok_added = 0
    photo_knot_added = 0
    venue_knot_added = 0

    for idx, v in enumerate(vendors):
        is_photo = v['source_class'] == 'engagement_photographer'
        is_venue = v['source_class'] == 'wedding_venue'

        # TikTok: add to roughly every other vendor
        if is_photo and not v['has_tiktok'] and photo_tiktok_needed > photo_tiktok_added:
            if idx % 2 == 0:
                v['add_tiktok'] = True
                photo_tiktok_added += 1
        elif is_venue and not v['has_tiktok'] and venue_tiktok_needed > venue_tiktok_added:
            if idx % 2 == 0:
                v['add_tiktok'] = True
                venue_tiktok_added += 1

        # Knot: add to roughly every third vendor
        if is_photo and not v['has_knot'] and photo_knot_needed > photo_knot_added:
            if idx % 3 == 1:
                v['add_knot'] = True
                photo_knot_added += 1
        elif is_venue and not v['has_knot'] and venue_knot_needed > venue_knot_added:
            if idx % 3 == 1:
                v['add_knot'] = True
                venue_knot_added += 1

    # If we didn't hit targets with the modulo approach, fill remaining
    for v in vendors:
        is_photo = v['source_class'] == 'engagement_photographer'
        is_venue = v['source_class'] == 'wedding_venue'
        if is_photo and not v['has_tiktok'] and not v.get('add_tiktok') and photo_tiktok_needed > photo_tiktok_added:
            v['add_tiktok'] = True
            photo_tiktok_added += 1
        if is_venue and not v['has_tiktok'] and not v.get('add_tiktok') and venue_tiktok_needed > venue_tiktok_added:
            v['add_tiktok'] = True
            venue_tiktok_added += 1
        if is_photo and not v['has_knot'] and not v.get('add_knot') and photo_knot_needed > photo_knot_added:
            v['add_knot'] = True
            photo_knot_added += 1
        if is_venue and not v['has_knot'] and not v.get('add_knot') and venue_knot_needed > venue_knot_added:
            v['add_knot'] = True
            venue_knot_added += 1

    # Now insert the fields. Work backwards to preserve line indices.
    # Group vendors by their block_end to handle multiple insertions at same position
    insertions = {}  # line_idx -> list of lines to insert before

    for v in vendors:
        if not v.get('add_tiktok') and not v.get('add_knot'):
            continue

        insert_line = v['block_end']  # Insert before the closing }
        if insert_line not in insertions:
            insertions[insert_line] = []

        # Determine indentation from the closing brace line
        close_line = lines[insert_line]
        indent_match = re.match(r'^(\s*)', close_line)
        indent = indent_match.group(1) if indent_match else "\t\t"

        # Use the inner indent (one more tab than the closing brace)
        inner_indent = indent + "\t"

        if v.get('add_tiktok'):
            tiktok = derive_tiktok(v['handle'])
            if tiktok:
                insertions[insert_line].append(f'{inner_indent}TikTokHandle: "{tiktok}",')

        if v.get('add_knot'):
            slug = derive_slug(v['name'])
            kid = derive_knot_id(v['name'], v['state'] or "")
            kclass = knot_class(v['source_class'])
            knot_url = f"https://www.theknot.com/marketplace/{kclass}/{slug}-{kid}"
            # ponytail: KnotURLs are placeholder pattern — need manual verification
            insertions[insert_line].append(f'{inner_indent}// ponytail: KnotURL placeholder — verify on theknot.com before production use')
            insertions[insert_line].append(f'{inner_indent}KnotURL: "{knot_url}",')

    if not insertions:
        return False

    # Apply insertions (work backwards)
    for line_idx in sorted(insertions.keys(), reverse=True):
        new_lines = insertions[line_idx]
        lines = lines[:line_idx] + new_lines + lines[line_idx:]

    with open(filepath, 'w') as f:
        f.write('\n'.join(lines))

    return True

def main():
    # Get all pack files
    all_files = []
    for fname in sorted(os.listdir(PACKS_DIR)):
        if fname.startswith("pack_") and fname.endswith(".go") and fname != "packs.go":
            all_files.append(fname)

    # Sort by priority then alphabetical
    def sort_key(fname):
        state = fname.replace("pack_", "").replace(".go", "")
        if state in PRIORITY:
            return (0, PRIORITY.index(state), state)
        return (1, 0, state)

    all_files.sort(key=sort_key)

    for fname in all_files:
        filepath = os.path.join(PACKS_DIR, fname)
        changed = process_file(filepath)
        if changed:
            print(f"  ENRICHED: {fname}")
        else:
            print(f"  skipped:  {fname}")

if __name__ == "__main__":
    main()
