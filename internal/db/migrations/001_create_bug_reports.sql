CREATE TABLE IF NOT EXISTS bug_reports (
    id            UUID PRIMARY KEY,
    site_id       TEXT NOT NULL,
    title         TEXT NOT NULL,
    description   TEXT NOT NULL,
    category      TEXT NOT NULL CHECK (category IN ('design', 'functionality', 'performance', 'content', 'mobile', 'security', 'other')),
    page_url      TEXT,
    contact_type  TEXT CHECK (contact_type IS NULL OR contact_type IN ('phone', 'email', 'telegram', 'instagram')),
    contact_value TEXT,
    first_name    TEXT,
    last_name     TEXT,
    status        TEXT NOT NULL DEFAULT 'new',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_bug_reports_site_created ON bug_reports (site_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_bug_reports_status ON bug_reports (status);
