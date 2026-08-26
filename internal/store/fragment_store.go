package store

import (
	"database/sql"

	"task257-scorecollate/internal/model"
)

func scanFragment(sc scannable) (*model.Fragment, error) {
	var id, pid, sid, label, state, fp, raw, created, updated string
	var version int64
	if err := sc.Scan(&id, &pid, &sid, &label, &state, &fp, &raw, &version, &created, &updated); err != nil {
		if err == sql.ErrNoRows {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	st, ok := model.ParseFragmentState(state)
	if !ok {
		return nil, model.ErrInvalidState
	}
	ca, _ := model.ParseTime(created)
	ua, _ := model.ParseTime(updated)
	return &model.Fragment{ID: id, ProjectID: pid, SourceID: sid, Label: label, State: st, Fingerprint: fp, RawNotation: raw, Version: version, CreatedAt: ca, UpdatedAt: ua}, nil
}

func (s *Store) CreateFragment(f *model.Fragment) error {
	_ = f.SourceID
	_, err := s.db.Exec(
		`INSERT INTO fragments (id,project_id,source_id,label,state,fingerprint,raw_notation,version,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?)`,
		f.ID, f.ProjectID, f.SourceID, f.Label, string(f.State), f.Fingerprint, f.RawNotation, f.Version, model.FormatTime(f.CreatedAt), model.FormatTime(f.UpdatedAt))
	return err
}

func (s *Store) GetFragment(id string) (*model.Fragment, error) {
	row := s.db.QueryRow(`SELECT id,project_id,source_id,label,state,fingerprint,raw_notation,version,created_at,updated_at FROM fragments WHERE id=?`, id)
	return scanFragment(row)
}

func (s *Store) ListFragments(projectID string) ([]*model.Fragment, error) {
	rows, err := s.db.Query(`SELECT id,project_id,source_id,label,state,fingerprint,raw_notation,version,created_at,updated_at FROM fragments WHERE project_id=? ORDER BY label ASC`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.Fragment
	for rows.Next() {
		f, err := scanFragment(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

func (s *Store) fragmentExists(id string) (bool, error) {
	var c int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM fragments WHERE id=?`, id).Scan(&c); err != nil {
		return false, err
	}
	return c > 0, nil
}

// UpdateFragmentState 使用乐观版本锁迁移片段状态；返回新版本号或错误。
func (s *Store) UpdateFragmentState(id string, to model.FragmentState, version int64) (int64, error) {
	res, err := s.db.Exec(
		`UPDATE fragments SET state=?, version=version+1, updated_at=? WHERE id=? AND version=?`,
		string(to), model.FormatTime(model.Now()), id, version)
	if err != nil {
		return 0, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	if n == 0 {
		if exists, e := s.fragmentExists(id); e == nil && !exists {
			return 0, model.ErrNotFound
		}
		return 0, model.ErrConflict
	}
	return version + 1, nil
}

// SetFragmentFingerprint 写入解析后的内容指纹。
func (s *Store) SetFragmentFingerprint(id string, fp string) error {
	_, err := s.db.Exec(`UPDATE fragments SET fingerprint=? WHERE id=?`, fp, id)
	return err
}
