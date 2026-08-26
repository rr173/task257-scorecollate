package source

import (
	"task257-scorecollate/internal/model"
)

// DetectCycle 检查为 newID 设置 parentID 是否会在来源谱系中形成环。
// 任一循环（含自环、已有环）都拒绝新增边，返回 model.ErrSelfCycle。
func DetectCycle(sources []*model.Source, newID, parentID string) error {
	if parentID == "" {
		return nil
	}
	if newID == parentID {
		return model.ErrSelfCycle
	}
	byID := make(map[string]*model.Source, len(sources))
	for _, s := range sources {
		byID[s.ID] = s
	}
	cur := parentID
	seen := map[string]bool{}
	for cur != "" {
		if cur == newID {
			return model.ErrSelfCycle
		}
		if seen[cur] {
			return model.ErrSelfCycle
		}
		seen[cur] = true
		par, ok := byID[cur]
		if !ok {
			break
		}
		if par.ParentID == nil {
			break
		}
		cur = *par.ParentID
	}
	return nil
}

// Roots 返回没有父本的祖本集合（用于独立来源支持判定）。
func Roots(sources []*model.Source) map[string]bool {
	roots := map[string]bool{}
	for _, s := range sources {
		if s.ParentID == nil {
			roots[s.ID] = true
		}
	}
	return roots
}
