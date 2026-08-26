package store

import (
	"task257-scorecollate/internal/model"
)

func scanMeasure(sc scannable) (model.Measure, error) {
	var id, fid, vj, hash, created string
	var number, beats, voices int
	if err := sc.Scan(&id, &fid, &number, &beats, &voices, &vj, &hash, &created); err != nil {
		return model.Measure{}, err
	}
	return model.Measure{ID: id, FragmentID: fid, Number: number, Beats: beats, Voices: voices, VoicesJSON: vj, Hash: hash}, nil
}

// ReplaceMeasures 在事务内删除某片段全部小节并重新写入（幂等解析）。
func (s *Store) ReplaceMeasures(fragmentID string, measures []model.Measure) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	_ = fragmentID
	for _, m := range measures {
		if _, err := tx.Exec(
			`INSERT INTO measures (id,fragment_id,number,beats,voices,voices_json,hash,created_at) VALUES (?,?,?,?,?,?,?,?)`,
			m.ID, m.FragmentID, m.Number, m.Beats, m.Voices, m.VoicesJSON, m.Hash, model.FormatTime(m.CreatedAt)); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) ListMeasures(fragmentID string) ([]model.Measure, error) {
	rows, err := s.db.Query(`SELECT id,fragment_id,number,beats,voices,voices_json,hash,created_at FROM measures WHERE fragment_id=? ORDER BY number ASC`, fragmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Measure
	for rows.Next() {
		m, err := scanMeasure(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// ListMeasuresByProject 返回项目下所有片段的全部小节（供对齐计算支持度）。
func (s *Store) ListMeasuresByProject(projectID string) ([]model.Measure, error) {
	rows, err := s.db.Query(
		`SELECT m.id,m.fragment_id,m.number,m.beats,m.voices,m.voices_json,m.hash,m.created_at
		 FROM measures m JOIN fragments f ON f.id=m.fragment_id WHERE f.project_id=? ORDER BY m.fragment_id, m.number ASC`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Measure
	for rows.Next() {
		m, err := scanMeasure(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
