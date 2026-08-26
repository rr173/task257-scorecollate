package service

import (
	"fmt"

	"github.com/google/uuid"
	"task257-scorecollate/internal/adjudicate"
	"task257-scorecollate/internal/align"
	"task257-scorecollate/internal/edition"
	"task257-scorecollate/internal/fragment"
	"task257-scorecollate/internal/model"
	"task257-scorecollate/internal/source"
	"task257-scorecollate/internal/store"
)

// Service 是跨模块的业务编排层。
type Service struct {
	store *store.Store
}

// New 构造服务。
func New(st *store.Store) *Service { return &Service{store: st} }

// ---------------------------------------------------------------------------
// 项目
// ---------------------------------------------------------------------------

func (s *Service) CreateProject(title, desc string) (*model.Project, error) {
	if title == "" {
		return nil, model.ErrInvalidInput
	}
	now := model.Now()
	p := &model.Project{
		ID: uuid.NewString(), Title: title, Description: desc,
		State: model.ProjectArranging, Version: 1, CreatedAt: now, UpdatedAt: now,
	}
	if err := s.store.CreateProject(p); err != nil {
		return nil, err
	}
	_ = s.store.AppendAudit(&model.AuditEntry{ID: uuid.NewString(), ProjectID: p.ID, Action: "create_project", Detail: title, CreatedAt: now})
	return p, nil
}

func (s *Service) GetProject(id string) (*model.Project, error) {
	return s.store.GetProject(id)
}

func (s *Service) ListProjects() ([]*model.Project, error) {
	return s.store.ListProjects()
}

// TransitionProject 按状态机迁移项目状态（封存后不可再动）。
func (s *Service) TransitionProject(id string, to model.ProjectState) (*model.Project, error) {
	p, err := s.store.GetProject(id)
	if err != nil {
		return nil, err
	}
	if p.State.IsTerminal() {
		return nil, model.ErrSealed
	}
	if !p.State.CanTransitionTo(to) {
		return nil, model.ErrInvalidState
	}
	newVer, err := s.store.UpdateProjectState(id, to, p.Version)
	if err != nil {
		return nil, err
	}
	p.State = to
	p.Version = newVer
	_ = s.store.AppendAudit(&model.AuditEntry{ID: uuid.NewString(), ProjectID: id, Action: "transition_project", Detail: string(to), CreatedAt: model.Now()})
	return p, nil
}

// ---------------------------------------------------------------------------
// 来源谱系
// ---------------------------------------------------------------------------

// CreateSource 新建传抄来源，校验父本存在、属于同一项目且不构成谱系环。
func (s *Service) CreateSource(projectID, siglum, title, parentID, desc string) (*model.Source, error) {
	if projectID == "" || siglum == "" {
		return nil, model.ErrInvalidInput
	}
	p, err := s.store.GetProject(projectID)
	if err != nil {
		return nil, err
	}
	if p.State.IsTerminal() {
		return nil, model.ErrSealed
	}
	var par *string
	if parentID != "" {
		src, err := s.store.GetSource(parentID)
		if err != nil {
			return nil, err
		}
		if src.ProjectID != projectID {
			return nil, model.ErrInvalidInput
		}
		par = &parentID
	}
	srcID := uuid.NewString()
	sources, err := s.store.ListSources(projectID)
	if err != nil {
		return nil, err
	}
	if err := source.DetectCycle(append(sources, &model.Source{ID: srcID, ParentID: par}), srcID, parentID); err != nil {
		return nil, err
	}
	src := &model.Source{ID: srcID, ProjectID: projectID, Siglum: siglum, Title: title, ParentID: par, Description: desc, CreatedAt: model.Now()}
	if err := s.store.CreateSource(src); err != nil {
		return nil, err
	}
	return src, nil
}

func (s *Service) ListSources(projectID string) ([]*model.Source, error) {
	return s.store.ListSources(projectID)
}

// ReparentSource 调整来源的传抄父本；若新父本会使谱系成环则拒绝。
func (s *Service) ReparentSource(id string, newParentID string) (*model.Source, error) {
	src, err := s.store.GetSource(id)
	if err != nil {
		return nil, err
	}
	p, err := s.store.GetProject(src.ProjectID)
	if err != nil {
		return nil, err
	}
	if p.State.IsTerminal() {
		return nil, model.ErrSealed
	}
	var par *string
	if newParentID != "" {
		np, err := s.store.GetSource(newParentID)
		if err != nil {
			return nil, err
		}
		if np.ProjectID != src.ProjectID {
			return nil, model.ErrInvalidInput
		}
		par = &newParentID
	}
	sources, err := s.store.ListSources(src.ProjectID)
	if err != nil {
		return nil, err
	}
	extended := make([]*model.Source, 0, len(sources))
	for _, s := range sources {
		if s.ID == src.ID {
			extended = append(extended, &model.Source{ID: src.ID, ProjectID: src.ProjectID, ParentID: par})
		} else {
			extended = append(extended, s)
		}
	}
	if err := source.DetectCycle(extended, src.ID, newParentID); err != nil {
		return nil, err
	}
	if err := s.store.UpdateSourceParent(id, par); err != nil {
		return nil, err
	}
	src.ParentID = par
	return src, nil
}

// ---------------------------------------------------------------------------
// 片段
// ---------------------------------------------------------------------------

// CreateFragment 新建乐谱片段（待解析态）。
func (s *Service) CreateFragment(projectID, sourceID, label, raw string) (*model.Fragment, error) {
	if projectID == "" || sourceID == "" || label == "" {
		return nil, model.ErrInvalidInput
	}
	p, err := s.store.GetProject(projectID)
	if err != nil {
		return nil, err
	}
	if p.State.IsTerminal() {
		return nil, model.ErrSealed
	}
	src, err := s.store.GetSource(sourceID)
	if err != nil {
		return nil, err
	}
	if src.ProjectID != projectID {
		return nil, model.ErrInvalidInput
	}
	now := model.Now()
	f := &model.Fragment{
		ID: uuid.NewString(), ProjectID: projectID, SourceID: sourceID, Label: label,
		State: model.FragPendingParse, RawNotation: raw, Version: 1, CreatedAt: now, UpdatedAt: now,
	}
	if err := s.store.CreateFragment(f); err != nil {
		return nil, err
	}
	return f, nil
}

func (s *Service) GetFragment(id string) (*model.Fragment, error) {
	return s.store.GetFragment(id)
}

func (s *Service) ListFragments(projectID string) ([]*model.Fragment, error) {
	return s.store.ListFragments(projectID)
}

func (s *Service) ListMeasures(fragmentID string) ([]model.Measure, error) {
	return s.store.ListMeasures(fragmentID)
}

// SetFragmentState 迁移片段状态（标记不可辨识/排除等）。
func (s *Service) SetFragmentState(id string, to model.FragmentState) (*model.Fragment, error) {
	f, err := s.store.GetFragment(id)
	if err != nil {
		return nil, err
	}
	if !f.State.CanTransitionTo(to) {
		return nil, model.ErrInvalidState
	}
	newVer, err := s.store.UpdateFragmentState(id, to, f.Version)
	if err != nil {
		return nil, err
	}
	f.State = to
	f.Version = newVer
	return f, nil
}

// ParseFragment 解析片段记谱为小节并持久化；解析失败标记不可辨识。
func (s *Service) ParseFragment(id string) (*model.Fragment, []model.Measure, error) {
	f, err := s.store.GetFragment(id)
	if err != nil {
		return nil, nil, err
	}
	if f.State != model.FragPendingParse && f.State != model.FragUnreadable {
		return nil, nil, model.ErrInvalidState
	}
	parsed, fp, err := fragment.Parse(f.RawNotation)
	if err != nil {
		newVer, e2 := s.store.UpdateFragmentState(id, model.FragUnreadable, f.Version)
		if e2 != nil {
			return nil, nil, e2
		}
		f.State = model.FragUnreadable
		f.Version = newVer
		return f, nil, err
	}
	now := model.Now()
	measures := fragment.ToModelMeasures(id, parsed, now)
	if err := s.store.ReplaceMeasures(id, measures); err != nil {
		return nil, nil, err
	}
	if err := s.store.SetFragmentFingerprint(id, fp); err != nil {
		return nil, nil, err
	}
	newVer, err := s.store.UpdateFragmentState(id, model.FragAligned, f.Version)
	if err != nil {
		return nil, nil, err
	}
	f.State = model.FragAligned
	f.Version = newVer
	f.Fingerprint = fp
	return f, measures, nil
}

// ---------------------------------------------------------------------------
// 对齐与异文
// ---------------------------------------------------------------------------

// AlignProject 以小节数最多的已对齐片段为参考，对其余片段逐对比对，
// 生成异文候选（含节拍断裂检测与独立来源支持度），并推进项目到待裁决。
func (s *Service) AlignProject(projectID string) (int, error) {
	p, err := s.store.GetProject(projectID)
	if err != nil {
		return 0, err
	}
	if p.State.IsTerminal() {
		return 0, model.ErrSealed
	}
	fragments, err := s.store.ListFragments(projectID)
	if err != nil {
		return 0, err
	}
	var aligned []*model.Fragment
	for _, f := range fragments {
		if f.State == model.FragAligned {
			aligned = append(aligned, f)
		}
	}
	if len(aligned) < 2 {
		return 0, fmt.Errorf("至少需要 2 个已对齐片段: %w", model.ErrInvalidInput)
	}
	ref := aligned[0]
	refMeasures, err := s.store.ListMeasures(ref.ID)
	if err != nil {
		return 0, err
	}
	maxN := len(refMeasures)
	for _, f := range aligned[1:] {
		ms, e := s.store.ListMeasures(f.ID)
		if e != nil {
			return 0, e
		}
		if len(ms) > maxN {
			maxN = len(ms)
			ref = f
			refMeasures = ms
		}
	}
	allMeasures, err := s.store.ListMeasuresByProject(projectID)
	if err != nil {
		return 0, err
	}
	sources, err := s.store.ListSources(projectID)
	if err != nil {
		return 0, err
	}
	rootBySource := make(map[string]string, len(sources))
	for _, src := range sources {
		rootBySource[src.ID] = source.RootID(sources, src.ID)
	}
	byFragment := map[string]map[int]model.Measure{}
	rootByFragment := make(map[string]string, len(fragments))
	for _, f := range fragments {
		rootByFragment[f.ID] = rootBySource[f.SourceID]
	}
	for _, m := range allMeasures {
		if byFragment[m.FragmentID] == nil {
			byFragment[m.FragmentID] = map[int]model.Measure{}
		}
		byFragment[m.FragmentID][m.Number] = m
	}
	if err := s.store.DeleteVariantsByProject(projectID); err != nil {
		return 0, err
	}
	created := 0
	for _, other := range aligned {
		if other.ID == ref.ID {
			continue
		}
		otherMeasures, e := s.store.ListMeasures(other.ID)
		if e != nil {
			return 0, e
		}
		cands := align.CompareFragmentPair(ref.ID, refMeasures, other.ID, otherMeasures)
		for _, c := range cands {
			beatBreak := align.BeatBreakInt(c.BeatsA, c.BeatsB)
			wellFormed := align.WellFormed(c.ContentA) && align.WellFormed(c.ContentB)
			support := adjudicate.SupportCount(byFragment, rootByFragment, c.FragmentAID, c.FragmentBID, c.MeasureNumber, c.Voice, c.ContentA, c.ContentB)
			kind := adjudicate.AssessKind(beatBreak, support, wellFormed)
			v := &model.Variant{
				ID: uuid.NewString(), ProjectID: projectID,
				MeasureNumber: c.MeasureNumber, Voice: c.Voice,
				FragmentAID: c.FragmentAID, FragmentBID: c.FragmentBID,
				ContentA: c.ContentA, ContentB: c.ContentB,
				DetectedKind: kind, State: model.VarCandidate,
				SupportCount: support, Version: 1, CreatedAt: model.Now(),
			}
			if err := s.store.CreateVariant(v); err != nil {
				return 0, err
			}
			created++
		}
	}
	_ = s.store.AppendAudit(&model.AuditEntry{ID: uuid.NewString(), ProjectID: projectID, Action: "align", Detail: fmt.Sprintf("variants=%d", created), CreatedAt: model.Now()})
	if p.State == model.ProjectArranging || p.State == model.ProjectAwaitingAlignment {
		if _, e := s.store.UpdateProjectState(projectID, model.ProjectAwaitingAdjud, p.Version); e == nil {
			p.State = model.ProjectAwaitingAdjud
		}
	}
	return created, nil
}

func (s *Service) ListVariants(projectID string) ([]*model.Variant, error) {
	return s.store.ListVariants(projectID)
}

func (s *Service) GetVariant(id string) (*model.Variant, error) {
	return s.store.GetVariant(id)
}

// AdjudicateVariant 提交裁决：候选 -> 讹写/有效变体/证据不足 -> 确认。
func (s *Service) AdjudicateVariant(id string, decision model.VariantState, reason, editor string) (*model.Variant, error) {
	v, err := s.store.GetVariant(id)
	if err != nil {
		return nil, err
	}
	if _, ok := adjudicate.ValidateAdjudication(v.State, decision); !ok {
		return nil, model.ErrInvalidState
	}
	adj := &model.Adjudication{ID: uuid.NewString(), VariantID: id, Decision: decision, Reason: reason, Editor: editor, CreatedAt: model.Now()}
	if err := s.store.CreateAdjudication(adj); err != nil {
		return nil, err
	}
	newVer, err := s.store.UpdateVariantState(id, decision, v.Version)
	if err != nil {
		return nil, err
	}
	v.State = decision
	v.Version = newVer
	_ = s.store.AppendAudit(&model.AuditEntry{ID: uuid.NewString(), ProjectID: v.ProjectID, Action: "adjudicate", Detail: fmt.Sprintf("%s->%s", id, decision), CreatedAt: model.Now()})
	return v, nil
}

// ---------------------------------------------------------------------------
// 校勘版本
// ---------------------------------------------------------------------------

func (s *Service) CreateEdition(projectID, title string) (*model.Edition, error) {
	if projectID == "" || title == "" {
		return nil, model.ErrInvalidInput
	}
	p, err := s.store.GetProject(projectID)
	if err != nil {
		return nil, err
	}
	if p.State.IsTerminal() {
		return nil, model.ErrSealed
	}
	now := model.Now()
	e := &model.Edition{ID: uuid.NewString(), ProjectID: projectID, Title: title, State: model.EditionDraft, Version: 1, CreatedAt: now}
	if err := s.store.CreateEdition(e); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *Service) ListEditions(projectID string) ([]*model.Edition, error) {
	return s.store.ListEditions(projectID)
}

func (s *Service) GetEdition(id string) (*model.Edition, error) {
	return s.store.GetEdition(id)
}

// PublishEdition 把项目内全部已确认异文冻结进版本并转为不可变冻结态。
func (s *Service) PublishEdition(id string) (*model.Edition, error) {
	e, err := s.store.GetEdition(id)
	if err != nil {
		return nil, err
	}
	if e.State.IsImmutable() {
		return nil, model.ErrFrozen
	}
	variants, err := s.store.ListVariants(e.ProjectID)
	if err != nil {
		return nil, err
	}
	confirmedIDs := edition.ConfirmedVariantIDs(variants)
	links := make([]model.EditionVariantLink, 0, len(variants))
	for _, v := range variants {
		links = append(links, model.EditionVariantLink{EditionID: id, VariantID: v.ID, Included: true})
	}
	_ = confirmedIDs
	if err := s.store.LinkEditionVariants(id, links); err != nil {
		return nil, err
	}
	newVer, err := s.store.UpdateEditionState(id, model.EditionFrozen, e.Version)
	if err != nil {
		return nil, err
	}
	e.State = model.EditionFrozen
	e.Version = newVer
	_ = s.store.AppendAudit(&model.AuditEntry{ID: uuid.NewString(), ProjectID: e.ProjectID, Action: "publish_edition", Detail: id, CreatedAt: model.Now()})
	return e, nil
}

// SupersedeEdition 废弃旧版本并新建替代版本（旧版转为 superseded）。
func (s *Service) SupersedeEdition(id string, newTitle string) (*model.Edition, error) {
	e, err := s.store.GetEdition(id)
	if err != nil {
		return nil, err
	}
	if e.State.IsImmutable() {
		return nil, model.ErrFrozen
	}
	title := newTitle
	if title == "" {
		title = e.Title + "（替代）"
	}
	now := model.Now()
	ne := &model.Edition{ID: uuid.NewString(), ProjectID: e.ProjectID, Title: title, State: model.EditionDraft, Version: 1, CreatedAt: now}
	if err := s.store.CreateEdition(ne); err != nil {
		return nil, err
	}
	if _, err := s.store.UpdateEditionState(id, model.EditionSuperseded, e.Version); err != nil {
		return nil, err
	}
	_ = s.store.AppendAudit(&model.AuditEntry{ID: uuid.NewString(), ProjectID: e.ProjectID, Action: "supersede_edition", Detail: fmt.Sprintf("%s->%s", id, ne.ID), CreatedAt: now})
	return ne, nil
}

// ---------------------------------------------------------------------------
// 自检
// ---------------------------------------------------------------------------

// SelfCheck 返回项目各实体的计数快照（供自检与页面展示）。
func (s *Service) SelfCheck(projectID string) (map[string]interface{}, error) {
	p, err := s.store.GetProject(projectID)
	if err != nil {
		return nil, err
	}
	frags, _ := s.store.ListFragments(projectID)
	sources, _ := s.store.ListSources(projectID)
	variants, _ := s.store.ListVariants(projectID)
	editions, _ := s.store.ListEditions(projectID)
	aligned := 0
	for _, f := range frags {
		if f.State == model.FragAligned {
			aligned++
		}
	}
	confirmed := 0
	for _, v := range variants {
		if v.State == model.VarConfirmed {
			confirmed++
		}
	}
	return map[string]interface{}{
		"project":       p.ID,
		"project_state": string(p.State),
		"fragments":     len(frags),
		"aligned":       aligned,
		"sources":       len(sources),
		"variants":      len(variants),
		"confirmed":     confirmed,
		"editions":      len(editions),
		"recoverable":   true,
	}, nil
}
