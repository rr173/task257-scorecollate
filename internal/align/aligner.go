package align

import (
	"encoding/json"

	"task257-scorecollate/internal/model"
)

// Candidate 是参考片段与另一片段在某小节/声部上的差异候选。
type Candidate struct {
	MeasureNumber int
	Voice         int
	FragmentAID   string
	FragmentBID   string
	ContentA      string
	ContentB      string
	BeatsA        int
	BeatsB        int
}

// CompareFragmentPair 返回参考片段与另一片段的全部音符差异候选。
// 当某小节仅一方存在时，视为整小节缺抄并按每个声部各产出一个候选
// （缺失方内容为空、拍数为 0）；当双方都有该小节时保留原有逐声部比对
// （含节拍断裂判定）。
func CompareFragmentPair(refID string, ref []model.Measure, otherID string, other []model.Measure) []Candidate {
	refByNum := indexByNumber(ref)
	otherByNum := indexByNumber(other)
	nums := make(map[int]struct{}, len(refByNum)+len(otherByNum))
	for n := range refByNum {
		nums[n] = struct{}{}
	}
	for n := range otherByNum {
		nums[n] = struct{}{}
	}
	var cands []Candidate
	for num := range nums {
		rm, rok := refByNum[num]
		om, ook := otherByNum[num]
		switch {
		case rok && ook:
			// 双方都有该小节：逐声部比对（原有节拍断裂判定走 BeatsA/BeatsB）。
			maxV := rm.Voices
			if om.Voices > maxV {
				maxV = om.Voices
			}
			ra, _ := voiceSlice(rm.VoicesJSON)
			ob, _ := voiceSlice(om.VoicesJSON)
			for v := 0; v < maxV; v++ {
				ca := voiceAt(ra, v)
				cb := voiceAt(ob, v)
				if ca != cb {
					cands = append(cands, Candidate{
						MeasureNumber: num, Voice: v,
						FragmentAID: refID, FragmentBID: otherID,
						ContentA: ca, ContentB: cb,
						BeatsA: rm.Beats, BeatsB: om.Beats,
					})
				}
			}
		case rok:
			// 仅参考有此小节：other 整小节缺抄，每声部一个候选。
			ra, _ := voiceSlice(rm.VoicesJSON)
			for v := 0; v < rm.Voices; v++ {
				cands = append(cands, Candidate{
					MeasureNumber: num, Voice: v,
					FragmentAID: refID, FragmentBID: otherID,
					ContentA: voiceAt(ra, v), ContentB: "",
					BeatsA: rm.Beats, BeatsB: 0,
				})
			}
		case ook:
			// 仅 other 有此小节：参考整小节缺抄，每声部一个候选。
			ob, _ := voiceSlice(om.VoicesJSON)
			for v := 0; v < om.Voices; v++ {
				cands = append(cands, Candidate{
					MeasureNumber: num, Voice: v,
					FragmentAID: refID, FragmentBID: otherID,
					ContentA: "", ContentB: voiceAt(ob, v),
					BeatsA: 0, BeatsB: om.Beats,
				})
			}
		}
	}
	return cands
}

func indexByNumber(ms []model.Measure) map[int]model.Measure {
	m := make(map[int]model.Measure, len(ms))
	for _, x := range ms {
		m[x.Number] = x
	}
	return m
}

func voiceSlice(jsonStr string) ([]string, error) {
	var out []string
	if err := json.Unmarshal([]byte(jsonStr), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// VoiceSlice 解析小节声部 JSON（公开，供跨包调用）。
func VoiceSlice(jsonStr string) ([]string, error) { return voiceSlice(jsonStr) }

func voiceAt(voices []string, i int) string {
	if i < 0 || i >= len(voices) {
		return ""
	}
	return voices[i]
}
