package adjudicate

import (
	"encoding/json"
	"testing"

	"task257-scorecollate/internal/model"
)

// mkMeasures 构造 片段ID -> (小节号 -> Measure)，便于支持度测试。
func mkMeasures(byFragment map[string]map[int][]string) map[string]map[int]model.Measure {
	out := map[string]map[int]model.Measure{}
	for fid, byNum := range byFragment {
		for num, voices := range byNum {
			js, _ := json.Marshal(voices)
			if out[fid] == nil {
				out[fid] = map[int]model.Measure{}
			}
			out[fid][num] = model.Measure{VoicesJSON: string(js), Voices: len(voices)}
		}
	}
	return out
}

// 同一祖本谱系内的多份传抄副本持相同读法，只计一次独立支持，
// 不应把同源副本重复计入而误判为有效变体。
func TestSupportCountDedupsSameLineage(t *testing.T) {
	// 谱系：rootX 是祖本，c1/c2/c3 均传抄自 rootX（同源）；fragmentA 独立祖本。
	byFragment := mkMeasures(map[string]map[int][]string{
		"A":  {1: {"X", "Y"}},
		"B":  {1: {"Z", "W"}},
		"c1": {1: {"X", "Y"}},
		"c2": {1: {"X", "Y"}},
		"c3": {1: {"X", "Y"}},
	})
	fragmentSource := map[string]string{
		"A": "srcA", "B": "srcB",
		"c1": "rootX", "c2": "rootX", "c3": "rootX",
	}
	sourceRoot := map[string]string{
		"srcA": "srcA", "srcB": "srcB", // A、B 各自独立祖本
		"rootX": "rootX",
	}
	got := SupportCount(byFragment, "A", "B", 1, 0, "X", "Z", fragmentSource, sourceRoot)
	if got != 1 {
		t.Fatalf("同源副本应只计 1 次独立支持，实际 %d", got)
	}
}

// 来自不同祖本谱系的副本各持一次读法，按各自祖本计入独立支持，保留有效变体判定。
func TestSupportCountCountsDistinctLineages(t *testing.T) {
	// 三个不同祖本谱系 rootP/rootQ/rootR 各有一份副本支持同一读法。
	byFragment := mkMeasures(map[string]map[int][]string{
		"A": {1: {"X", "Y"}},
		"B": {1: {"Z", "W"}},
		"p": {1: {"X", "Y"}},
		"q": {1: {"X", "Y"}},
		"r": {1: {"X", "Y"}},
	})
	fragmentSource := map[string]string{
		"A": "srcA", "B": "srcB",
		"p": "rootP", "q": "rootQ", "r": "rootR",
	}
	sourceRoot := map[string]string{
		"srcA": "srcA", "srcB": "srcB",
		"rootP": "rootP", "rootQ": "rootQ", "rootR": "rootR",
	}
	got := SupportCount(byFragment, "A", "B", 1, 0, "X", "Z", fragmentSource, sourceRoot)
	if got != 3 {
		t.Fatalf("三个不同祖本谱系应计 3 次独立支持，实际 %d", got)
	}
	if AssessKind(false, got, true) != model.VarValidVariant {
		t.Fatal("存在独立来源支持应判有效变体")
	}
}

// A/B 自身谱系内的副本不算独立支持（不能为自家读法提供独立印证）。
func TestSupportCountExcludesABLineages(t *testing.T) {
	byFragment := mkMeasures(map[string]map[int][]string{
		"A": {1: {"X", "Y"}},
		"B": {1: {"Z", "W"}},
		// 这两份副本与 A 同祖本谱系（srcA 的传抄）
		"a1": {1: {"X", "Y"}},
		"a2": {1: {"X", "Y"}},
	})
	fragmentSource := map[string]string{
		"A": "srcA", "B": "srcB",
		"a1": "childA", "a2": "childA",
	}
	sourceRoot := map[string]string{
		"srcA": "srcA", "srcB": "srcB", "childA": "srcA", // childA 上溯到 srcA，与 A 同谱系
	}
	got := SupportCount(byFragment, "A", "B", 1, 0, "X", "Z", fragmentSource, sourceRoot)
	if got != 0 {
		t.Fatalf("A/B 谱系内副本不应计入独立支持，实际 %d", got)
	}
	if AssessKind(false, got, true) != model.VarInsufficient {
		t.Fatal("无独立来源支持应判证据不足")
	}
}
