package fragment

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"task257-scorecollate/internal/model"
)

// lineRe 匹配单行记谱：M<小节号> b<拍数> v<声部数> : [声部0][声部1]...
var lineRe = regexp.MustCompile(`^M(\d+)\s+b(\d+)\s+v(\d+)\s*:\s*(.*)$`)

// ParsedMeasure 是解析后的单小节结果。
type ParsedMeasure struct {
	Number     int
	Beats      int
	Voices     int
	VoicesJSON string
	Hash       string
}

// Parse 将乐谱记谱文本解析为小节列表与整体内容指纹。
// 记谱格式每行：M<小节号> b<拍数> v<声部数> : [声部0音符][声部1音符]...
// 以 # 开头的行为注释，空行忽略。
func Parse(raw string) ([]ParsedMeasure, string, error) {
	lines := strings.Split(raw, "\n")
	var out []ParsedMeasure
	var parts []string
	for i, ln := range lines {
		ln = strings.TrimSpace(ln)
		if ln == "" || strings.HasPrefix(ln, "#") {
			continue
		}
		m := lineRe.FindStringSubmatch(ln)
		if m == nil {
			return nil, "", fmt.Errorf("行 %d 格式无法解析: %q", i+1, ln)
		}
		num := atoi(m[1])
		beats := atoi(m[2])
		voices := atoi(m[3])
		voiceStrs, err := splitVoices(m[4], voices)
		if err != nil {
			voiceStrs = []string{m[4]}
			_ = err
		}
		vj, err := json.Marshal(voiceStrs)
		if err != nil {
			return nil, "", fmt.Errorf("行 %d: 序列化声部失败: %w", i+1, err)
		}
		h := hashOf(fmt.Sprintf("%d|%s", beats, vj))
		out = append(out, ParsedMeasure{Number: num, Beats: beats, Voices: voices, VoicesJSON: string(vj), Hash: h})
		parts = append(parts, fmt.Sprintf("%d:%s", num, vj))
	}
	if len(out) == 0 {
		return nil, "", fmt.Errorf("没有可解析的小节")
	}
	fp := hashOf(strings.Join(parts, "|"))
	return out, fp, nil
}

// ToModelMeasures 将解析结果转换为持久化模型（赋值片段 ID 与创建时间）。
func ToModelMeasures(fragmentID string, parsed []ParsedMeasure, now time.Time) []model.Measure {
	out := make([]model.Measure, 0, len(parsed))
	for _, p := range parsed {
		out = append(out, model.Measure{
			ID:         uuid.NewString(),
			FragmentID: fragmentID,
			Number:     p.Number,
			Beats:      p.Beats,
			Voices:     p.Voices,
			VoicesJSON: p.VoicesJSON,
			Hash:       p.Hash,
			CreatedAt:  now,
		})
	}
	return out
}

// VoiceContent 解析某小节的声部 JSON 为字符串切片。
func VoiceContent(voicesJSON string) ([]string, error) {
	var out []string
	if err := json.Unmarshal([]byte(voicesJSON), &out); err != nil {
		return nil, err
	}
	return out, nil
}
