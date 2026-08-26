package store

import (
	"database/sql"

	"task257-scorecollate/internal/model"
)

func scanSource(sc scannable) (*model.Source, error) {
	var id, pid, siglum, title, desc, created string
	var parent sql.NullString
	if err := sc.Scan(&id, &pid, &siglum, &title, &parent, &desc, &created); err != nil {
		if err == sql.ErrNoRows {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	src := &model.Source{ID: id, ProjectID: pid, Siglum: siglum, Title: title, Description: desc}
	if parent.Valid {
		p := parent.String
		src.ParentID = &p
	}
	return src, nil
}

func (s *Store) CreateSource(src *model.Source) error {
	var parent interface{}
	if src.ParentID != nil {
		parent = *src.ParentID
	}
	_, err := s.db.Exec(
		`INSERT INTO sources (id,project_id,siglum,title,parent_id,description,created_at) VALUES (?,?,?,?,?,?,?)`,
		src.ID, src.ProjectID, src.Siglum, src.Title, parent, src.Description, model.FormatTime(src.CreatedAt))
	return err
}

func (s *Store) GetSource(id string) (*model.Source, error) {
	row := s.db.QueryRow(`SELECT id,project_id,siglum,title,parent_id,description,created_at FROM sources WHERE id=?`, id)
	return scanSource(row)
}

func (s *Store) ListSources(projectID string) ([]*model.Source, error) {
	rows, err := s.db.Query(`SELECT id,project_id,siglum,title,parent_id,description,created_at FROM sources WHERE project_id=? ORDER BY siglum ASC`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.Source
	for rows.Next() {
		src, err := scanSource(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, src)
	}
	return out, rows.Err()
}

func (s *Store) SourceExists(id string) (bool, error) {
	var c int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM sources WHERE id=?`, id).Scan(&c); err != nil {
		return false, err
	}
	return c > 0, nil
}

// UpdateSourceParent 重设来源父本（传抄谱系调整）。
func (s *Store) UpdateSourceParent(id string, parent interface{}) error {
	_, err := s.db.Exec(`UPDATE sources SET parent_id=? WHERE id=?`, parent, id)
	return err
}
