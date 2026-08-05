-- Ring detection + proposal photo classification upgrade.
-- Extends the vision_classifications log with YOLOv8 ring confidence and
-- CLIP zero-shot photo label so the dashboard can show per-post vision analysis.

ALTER TABLE vision_classifications
  ADD COLUMN IF NOT EXISTS ring_confidence REAL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS photo_label TEXT DEFAULT '',
  ADD COLUMN IF NOT EXISTS photo_confidence REAL DEFAULT 0;
