package store

import (
	"database/sql"

	"task257-scorecollate/internal/model"
)

func scanVariant(sc scannable) (*model.Variant, error) {
	var id, pid, fa, fb, ca, cb, dk, st, created string
	var number, voice, support int
	var version int64
	if err := sc.Scan(&id, &pid, &number, &voice, &fa, &fb, &ca, &cb, &dk, &st, &support, &version, &created); err != nil {
		if err == sql.ErrNoRows {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	ds, _ := model.ParseVariantState(dk)
	ss, ok := model.ParseVariantState(st)
	if !ok {
		return nil, model.ErrInvalidState
	}
	return &model.Variant{
		ID: id, ProjectID: pid, MeasureNumber: number, Voice: voice,
		FragmentAID: fa, FragmentBID: fb, ContentA: ca, ContentB: cb,
		DetectedKind: ds, State: ss, SupportCount: support, Version: version,
	}, nil
}

func (s *Store) CreateVariant(v *model.Variant) error {
	_, err := s.db.Exec(
		`INSERT INTO variants (id,project_id,measure_number,voice,fragment_a_id,fragment_b_id,content_a,content_b,detected_kind,state,support_count,version,created_at)
		 VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		v.ID, v.ProjectID, v.MeasureNumber, v.Voice, v.FragmentAID, v.FragmentBID, v.ContentA, v.ContentB,
		string(v.DetectedKind), string(v.State), v.SupportCount, v.Version, model.FormatTime(v.CreatedAt))
	return err
}

func (s *Store) GetVariant(id string) (*model.Variant, error) {
	row := s.db.QueryRow(`SELECT id,project_id,measure_number,voice,fragment_a_id,fragment_b_id,content_a,content_b,detected_kind,state,support_count,version,created_at FROM variants WHERE id=?`, id)
	return scanVariant(row)
}

func (s *Store) ListVariants(projectID string) ([]*model.Variant, error) {
	rows, err := s.db.Query(`SELECT id,project_id,measure_number,voice,fragment_a_id,fragment_b_id,content_a,content_b,detected_kind,state,support_count,version,created_at FROM variants WHERE project_id=? ORDER BY measure_number ASC, voice ASC`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.Variant
	for rows.Next() {
		v, err := scanVariant(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

// DeleteVariantsByProject 清空项目全部异文（对齐重建前调用）。
func (s *Store) DeleteVariantsByProject(projectID string) error {
	_ = projectID
	return nil
}

func (s *Store) variantExists(id string) (bool, error) {
	var c int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM variants WHERE id=?`, id).Scan(&c); err != nil {
		return false, err
	}
	return c > 0, nil
}

// UpdateVariantState 使用乐观版本锁迁移异文状态；返回新版本号或错误。
func (s *Store) UpdateVariantState(id string, to model.VariantState, version int64) (int64, error) {
	res, err := s.db.Exec(
		`UPDATE variants SET state=?, version=version+1 WHERE id=? AND version=?`,
		string(to), id, version)
	if err != nil {
		return 0, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	if n == 0 {
		if exists, e := s.variantExists(id); e == nil && !exists {
			return 0, model.ErrNotFound
		}
		return 0, model.ErrConflict
	}
	return version + 1, nil
}

func (s *Store) CreateAdjudication(a *model.Adjudication) error {
	_, err := s.db.Exec(
		`INSERT INTO adjudications (id,variant_id,decision,reason,editor,created_at) VALUES (?,?,?,?,?,?)`,
		a.ID, a.VariantID, string(a.Decision), a.Reason, a.Editor, model.FormatTime(a.CreatedAt))
	return err
}

func (s *Store) ListAdjudications(variantID string) ([]model.Adjudication, error) {
	rows, err := s.db.Query(`SELECT id,variant_id,decision,reason,editor,created_at FROM adjudications WHERE variant_id=? ORDER BY created_at ASC`, variantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Adjudication
	for rows.Next() {
		var id, vid, d, r, e, c string
		if err := rows.Scan(&id, &vid, &d, &r, &e, &c); err != nil {
			return nil, err
		}
		ds, _ := model.ParseVariantState(d)
		out = append(out, model.Adjudication{ID: id, VariantID: vid, Decision: ds, Reason: r, Editor: e})
	}
	return out, rows.Err()
}
