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

// SupportCount 统计在除 A/B 外的片段中，于同一小节/声部上持相同读法（等于 contentA 或 contentB）的来源数。
// byFragment 为 片段ID -> (小节号 -> 该小节 Measure)。
func SupportCount(byFragment map[string]map[int]model.Measure, excludeA, excludeB string, number, voice int, contentA, contentB string) int {
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
		if c == contentA || c == contentB {
			count++
		}
	}
	return count
}
