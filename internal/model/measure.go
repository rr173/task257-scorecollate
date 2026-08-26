package model

import "time"

// Measure 是片段中一个小节，含节拍数与各声部音符内容。
type Measure struct {
	ID         string
	FragmentID string
	Number     int    // 小节编号（在片段内唯一）
	Beats      int    // 该小节拍数
	Voices     int    // 声部数
	VoicesJSON string // JSON 数组：每个声部的音符串
	Hash       string // 内容指纹（拍数 + 声部内容）
	CreatedAt  time.Time
}
