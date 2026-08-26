package store

import "database/sql"

// schemaStatements 返回建表与建索引的全部 DDL（幂等）。
func schemaStatements() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS projects (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			state TEXT NOT NULL,
			version INTEGER NOT NULL DEFAULT 1,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS sources (
			id TEXT PRIMARY KEY,
			project_id TEXT NOT NULL REFERENCES projects(id),
			siglum TEXT NOT NULL,
			title TEXT NOT NULL DEFAULT '',
			parent_id TEXT,
			description TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS fragments (
			id TEXT PRIMARY KEY,
			project_id TEXT NOT NULL REFERENCES projects(id),
			source_id TEXT NOT NULL REFERENCES sources(id),
			label TEXT NOT NULL,
			state TEXT NOT NULL,
			fingerprint TEXT NOT NULL DEFAULT '',
			raw_notation TEXT NOT NULL DEFAULT '',
			version INTEGER NOT NULL DEFAULT 1,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS measures (
			id TEXT PRIMARY KEY,
			fragment_id TEXT NOT NULL REFERENCES fragments(id),
			number INTEGER NOT NULL,
			beats INTEGER NOT NULL DEFAULT 0,
			voices INTEGER NOT NULL DEFAULT 0,
			voices_json TEXT NOT NULL DEFAULT '',
			hash TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			UNIQUE(fragment_id, number)
		)`,
		`CREATE TABLE IF NOT EXISTS variants (
			id TEXT PRIMARY KEY,
			project_id TEXT NOT NULL REFERENCES projects(id),
			measure_number INTEGER NOT NULL,
			voice INTEGER NOT NULL,
			fragment_a_id TEXT NOT NULL,
			fragment_b_id TEXT NOT NULL,
			content_a TEXT NOT NULL DEFAULT '',
			content_b TEXT NOT NULL DEFAULT '',
			detected_kind TEXT NOT NULL,
			state TEXT NOT NULL,
			support_count INTEGER NOT NULL DEFAULT 0,
			version INTEGER NOT NULL DEFAULT 1,
			created_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS adjudications (
			id TEXT PRIMARY KEY,
			variant_id TEXT NOT NULL REFERENCES variants(id),
			decision TEXT NOT NULL,
			reason TEXT NOT NULL DEFAULT '',
			editor TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS editions (
			id TEXT PRIMARY KEY,
			project_id TEXT NOT NULL REFERENCES projects(id),
			title TEXT NOT NULL,
			state TEXT NOT NULL,
			version INTEGER NOT NULL DEFAULT 1,
			created_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS edition_variants (
			edition_id TEXT NOT NULL REFERENCES editions(id),
			variant_id TEXT NOT NULL REFERENCES variants(id),
			included INTEGER NOT NULL DEFAULT 1,
			PRIMARY KEY (edition_id, variant_id)
		)`,
		`CREATE TABLE IF NOT EXISTS audit_log (
			id TEXT PRIMARY KEY,
			project_id TEXT NOT NULL,
			action TEXT NOT NULL,
			detail TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_fragments_project ON fragments(project_id)`,
		`CREATE INDEX IF NOT EXISTS idx_sources_project ON sources(project_id)`,
		`CREATE INDEX IF NOT EXISTS idx_measures_fragment ON measures(fragment_id)`,
		`CREATE INDEX IF NOT EXISTS idx_variants_project ON variants(project_id)`,
		`CREATE INDEX IF NOT EXISTS idx_editions_project ON editions(project_id)`,
	}
}

// Migrate 执行全部建表/建索引语句。
func Migrate(db *sql.DB) error {
	for _, stmt := range schemaStatements() {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}
