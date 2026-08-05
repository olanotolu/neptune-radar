-- QR scan tracking: physical-digital attribution layer.
-- When a postcard QR is scanned, /r/{code} logs the scan then 302s to the
-- celebrate deep link. These columns on congratulate_kits hold the running
-- count and last-scan timestamp; individual scan events live in audit_events.

ALTER TABLE congratulate_kits
  ADD COLUMN IF NOT EXISTS qr_scan_count INT NOT NULL DEFAULT 0;
ALTER TABLE congratulate_kits
  ADD COLUMN IF NOT EXISTS last_qr_scan_at TIMESTAMPTZ;
