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

// ResolveRoots 返回 sourceID -> 祖本ID 的映射：沿 parent 链上溯到无父本的祖本。
// 同一祖本谱系内的全部抄本（含祖本自身）映射到同一个祖本ID，供校勘独立来源支持去重：
// 同一谱系的重复副本不应重复计入独立来源支持。
func ResolveRoots(sources []*model.Source) map[string]string {
	byID := make(map[string]*model.Source, len(sources))
	for _, s := range sources {
		byID[s.ID] = s
	}
	roots := make(map[string]string, len(sources))
	for _, s := range sources {
		cur := s.ID
		seen := map[string]bool{}
		for {
			if seen[cur] {
				// 谱系出现环（理论上创建/重挂时已拒绝）；以当前节点为祖本兜底，避免死循环
				roots[s.ID] = cur
				break
			}
			seen[cur] = true
			node, ok := byID[cur]
			if !ok || node.ParentID == nil {
				roots[s.ID] = cur
				break
			}
			cur = *node.ParentID
		}
	}
	return roots
}
