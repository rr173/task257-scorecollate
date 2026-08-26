package align

// BeatBreakInt 判断两拍数是否构成节拍断裂（拍数不一致即视为断裂）。
func BeatBreakInt(a, b int) bool {
	if a == 0 || b == 0 {
		return false
	}
	return a != b
}

// WellFormed 判断声部内容是否为可接受记谱（非空且非占位）。
func WellFormed(content string) bool {
	return content != "" && content != "-"
}
