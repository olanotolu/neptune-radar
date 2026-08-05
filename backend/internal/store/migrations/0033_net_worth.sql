-- Net Worth Estimation: store couple net-worth estimate from property + Instagram signals.
-- net_worth_estimate is the computed total (internal operator use only — never on postcards).
-- net_worth_tier is the confidence tier (high|medium|low).
-- net_worth_breakdown_json holds the category→amount map for the dashboard.
ALTER TABLE congratulate_kits
  ADD COLUMN IF NOT EXISTS net_worth_estimate BIGINT NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS net_worth_tier TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS net_worth_breakdown_json TEXT NOT NULL DEFAULT '';
