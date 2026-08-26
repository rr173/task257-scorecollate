package adjudicate

import (
	"task257-scorecollate/internal/align"
	"task257-scorecollate/internal/model"
)

// AssessKind 根据节拍断裂、独立来源支持与合法性初判异文类型：
//   - 拍数断裂或一方缺失/非法 -> 讹写(error)
//   - 双方合法且存在独立来源支持 -> 有效变体(variant)
//   - 双方合法但无独立支持 -> 证据不足(insufficient)
func AssessKind(beatBreak bool, support int, wellFormed bool) model.VariantState {
	if beatBreak || !wellFormed {
		return model.VarError
	}
	if support >= 1 {
		return model.VarValidVariant
	}
	return model.VarInsufficient
}

// SupportCount 统计除 A/B 所在祖本谱系外，于同一小节/声部上持相同读法
// （等于 contentA 或 contentB）的独立祖本谱系数。
//
// 仅当某读法得到来自不同祖本谱系的副本支持时才算独立来源：同一祖本谱系
// 内的多份传抄副本属同源，不重复计入独立支持；源自 A 或 B 自身谱系的副本
// 也不算独立支持。fragmentSource 将片段ID映射到其来源ID，sourceRoot 将
// 来源ID映射到所属祖本ID（同祖本谱系内的来源映射到同一值）。
//
// byFragment 为 片段ID -> (小节号 -> 该小节 Measure)。
func SupportCount(byFragment map[string]map[int]model.Measure, excludeA, excludeB string, number, voice int, contentA, contentB string, fragmentSource map[string]string, sourceRoot map[string]string) int {
	// A、B 各自所属的祖本谱系，任何源自这两条谱系的副本都不算独立支持。
	excludedRoots := map[string]bool{}
	for _, fid := range []string{excludeA, excludeB} {
		if sid, ok := fragmentSource[fid]; ok {
			if root, ok := sourceRoot[sid]; ok {
				excludedRoots[root] = true
			}
		}
	}
	countedRoots := map[string]bool{}
	count := 0
	for fid, byNum := range byFragment {
		if fid == excludeA || fid == excludeB {
			continue
		}
		m, ok := byNum[number]
		if !ok {
			continue
		}
		voices, err := align.VoiceSlice(m.VoicesJSON)
		if err != nil {
			continue
		}
		if voice < 0 || voice >= len(voices) {
			continue
		}
		c := voices[voice]
		if c != contentA && c != contentB {
			continue
		}
		// 按祖本谱系去重：同源副本只计一次独立支持。
		sid, ok := fragmentSource[fid]
		if !ok {
			continue
		}
		root, ok := sourceRoot[sid]
		if !ok {
			continue
		}
		if excludedRoots[root] {
			continue
		}
		if countedRoots[root] {
			continue
		}
		countedRoots[root] = true
		count++
	}
	return count
}
