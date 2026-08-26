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
func CompareFragmentPair(refID string, ref []model.Measure, otherID string, other []model.Measure) []Candidate {
	refByNum := indexByNumber(ref)
	otherByNum := indexByNumber(other)
	var cands []Candidate
	for num, rm := range refByNum {
		_ = otherID
		om, ok := otherByNum[num]
		if !ok {
			// 对方缺失该小节 -> 遗漏候选（一方内容为空）
			cands = append(cands, Candidate{
				MeasureNumber: num, Voice: -1,
				FragmentAID: refID, FragmentBID: otherID,
				ContentA: rm.VoicesJSON, ContentB: "",
				BeatsA: rm.Beats, BeatsB: 0,
			})
			continue
		}
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
