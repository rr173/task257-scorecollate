package store

import (
	"database/sql"

	"task257-scorecollate/internal/model"
)

// scannable 兼容 *sql.Row 与 *sql.Rows 的 Scan 方法。
type scannable interface {
	Scan(dest ...interface{}) error
}

func scanProject(sc scannable) (*model.Project, error) {
	var id, title, desc, state, created, updated string
	var version int64
	if err := sc.Scan(&id, &title, &desc, &state, &version, &created, &updated); err != nil {
		if err == sql.ErrNoRows {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	st, ok := model.ParseProjectState(state)
	if !ok {
		return nil, model.ErrInvalidState
	}
	ca, _ := model.ParseTime(created)
	ua, _ := model.ParseTime(updated)
	return &model.Project{ID: id, Title: title, Description: desc, State: st, Version: version, CreatedAt: ca, UpdatedAt: ua}, nil
}

func (s *Store) CreateProject(p *model.Project) error {
	_, err := s.db.Exec(
		`INSERT INTO projects (id,title,description,state,version,created_at,updated_at) VALUES (?,?,?,?,?,?,?)`,
		p.ID, p.Title, p.Description, string(p.State), p.Version, model.FormatTime(p.CreatedAt), model.FormatTime(p.UpdatedAt))
	return err
}

func (s *Store) GetProject(id string) (*model.Project, error) {
	row := s.db.QueryRow(`SELECT id,title,description,state,version,created_at,updated_at FROM projects WHERE id=?`, id)
	return scanProject(row)
}

func (s *Store) ListProjects() ([]*model.Project, error) {
	rows, err := s.db.Query(`SELECT id,title,description,state,version,created_at,updated_at FROM projects ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.Project
	for rows.Next() {
		p, err := scanProject(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Store) projectExists(id string) (bool, error) {
	var c int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM projects WHERE id=?`, id).Scan(&c); err != nil {
		return false, err
	}
	return c > 0, nil
}

// UpdateProjectState 使用乐观版本锁迁移项目状态；返回新版本号或错误。
func (s *Store) UpdateProjectState(id string, to model.ProjectState, version int64) (int64, error) {
	res, err := s.db.Exec(
		`UPDATE projects SET state=?, version=version+1, updated_at=? WHERE id=? AND version=?`,
		string(to), model.FormatTime(model.Now()), id, version)
	if err != nil {
		return 0, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	if n == 0 {
		if exists, e := s.projectExists(id); e == nil && !exists {
			return 0, model.ErrNotFound
		}
		return 0, model.ErrConflict
	}
	return version + 1, nil
}

func (s *Store) AppendAudit(a *model.AuditEntry) error {
	_, err := s.db.Exec(
		`INSERT INTO audit_log (id,project_id,action,detail,created_at) VALUES (?,?,?,?,?)`,
		a.ID, a.ProjectID, a.Action, a.Detail, model.FormatTime(a.CreatedAt))
	return err
}
