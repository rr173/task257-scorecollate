package fragment

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func hashOf(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func atoi(s string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(s))
	return n
}

// normalizeVoice 将声部音符串标准化：大写音名、去多余空白、保留音高记号。
func normalizeVoice(s string) string {
	s = strings.ToUpper(strings.TrimSpace(s))
	return strings.Join(strings.Fields(s), " ")
}

var voiceBracketRe = regexp.MustCompile(`\[([^\[\]]*)\]`)

// splitVoices 将 "[C4 E4][G4 A4]" 解析为 ["C4 E4","G4 A4"]，并校验括号数等于声部数。
// 括号数与声部数不一致时返回错误：声明声部数与实际记谱声部对不上，整段应判为不可辨识。
func splitVoices(content string, voices int) ([]string, error) {
	matches := voiceBracketRe.FindAllStringSubmatch(content, -1)
	if len(matches) != voices {
		return nil, fmt.Errorf("声部数不匹配: 声明 %d 声部, 实际 %d 个声部括号", voices, len(matches))
	}
	got := make([]string, 0, len(matches))
	for _, m := range matches {
		got = append(got, normalizeVoice(m[1]))
	}
	return got, nil
}
