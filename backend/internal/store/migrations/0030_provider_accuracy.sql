-- Bayesian Provider Fusion: track each provider's historical accuracy per state.
-- Accuracy = successful / total_attempts. Cold start uses 0.5 prior when no rows.
CREATE TABLE IF NOT EXISTS provider_accuracy (
    provider_name  TEXT NOT NULL,
    state          TEXT NOT NULL,
    total_attempts INT NOT NULL DEFAULT 0,
    successful     INT NOT NULL DEFAULT 0,
    last_updated   TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (provider_name, state)
);
