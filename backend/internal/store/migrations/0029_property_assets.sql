-- Asset Discovery from Property Records: store parsed county auditor financial data.
-- asset_json holds PropertyAsset (assessed value, sqft, year built, lot size, tax).
-- estimated_home_value is the computed market estimate (internal operator use only).
ALTER TABLE congratulate_kits
  ADD COLUMN IF NOT EXISTS asset_json TEXT,
  ADD COLUMN IF NOT EXISTS estimated_home_value BIGINT NOT NULL DEFAULT 0;
