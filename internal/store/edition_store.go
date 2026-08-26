package store

import (
	"database/sql"

	"task257-scorecollate/internal/model"
)

func scanEdition(sc scannable) (*model.Edition, error) {
	var id, pid, title, state, created string
	var version int64
	if err := sc.Scan(&id, &pid, &title, &state, &version, &created); err != nil {
		if err == sql.ErrNoRows {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	st, ok := model.ParseEditionState(state)
	if !ok {
		return nil, model.ErrInvalidState
	}
	return &model.Edition{ID: id, ProjectID: pid, Title: title, State: st, Version: version}, nil
}

func (s *Store) CreateEdition(e *model.Edition) error {
	_, err := s.db.Exec(
		`INSERT INTO editions (id,project_id,title,state,version,created_at) VALUES (?,?,?,?,?,?)`,
		e.ID, e.ProjectID, e.Title, string(e.State), e.Version, model.FormatTime(e.CreatedAt))
	return err
}

func (s *Store) GetEdition(id string) (*model.Edition, error) {
	row := s.db.QueryRow(`SELECT id,project_id,title,state,version,created_at FROM editions WHERE id=?`, id)
	return scanEdition(row)
}

func (s *Store) ListEditions(projectID string) ([]*model.Edition, error) {
	rows, err := s.db.Query(`SELECT id,project_id,title,state,version,created_at FROM editions WHERE project_id=? ORDER BY created_at ASC`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.Edition
	for rows.Next() {
		e, err := scanEdition(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (s *Store) editionExists(id string) (bool, error) {
	var c int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM editions WHERE id=?`, id).Scan(&c); err != nil {
		return false, err
	}
	return c > 0, nil
}

// UpdateEditionState 使用乐观版本锁迁移版本状态；返回新版本号或错误。
func (s *Store) UpdateEditionState(id string, to model.EditionState, version int64) (int64, error) {
	res, err := s.db.Exec(
		`UPDATE editions SET state=?, version=version+1 WHERE id=? AND version=?`,
		string(to), id, version)
	if err != nil {
		return 0, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	if n == 0 {
		if exists, e := s.editionExists(id); e == nil && !exists {
			return 0, model.ErrNotFound
		}
		return 0, model.ErrConflict
	}
	return version + 1, nil
}

// LinkEditionVariants 以事务写入版本-异文收录关系。
func (s *Store) LinkEditionVariants(editionID string, links []model.EditionVariantLink) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	for _, l := range links {
		inc := 0
		if l.Included {
			inc = 1
		}
		if _, err := tx.Exec(
			`INSERT INTO edition_variants (edition_id,variant_id,included) VALUES (?,?,?)`,
			editionID, l.VariantID, inc); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) ListEditionVariants(editionID string) ([]model.EditionVariantLink, error) {
	rows, err := s.db.Query(`SELECT edition_id,variant_id,included FROM edition_variants WHERE edition_id=?`, editionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.EditionVariantLink
	for rows.Next() {
		var eid, vid string
		var inc int
		if err := rows.Scan(&eid, &vid, &inc); err != nil {
			return nil, err
		}
		out = append(out, model.EditionVariantLink{EditionID: eid, VariantID: vid, Included: inc == 1})
	}
	return out, rows.Err()
}
