package model

import "time"

// Now 返回持久化用的当前时间（UTC），保证跨时区重启后时间语义一致。
func Now() time.Time { return time.Now().UTC() }

// ParseTime 解析 RFC3339 文本时间。
func ParseTime(s string) (time.Time, error) { return time.Parse(time.RFC3339, s) }

// FormatTime 将时间格式化为 RFC3339 文本。
func FormatTime(t time.Time) string { return t.UTC().Format(time.RFC3339) }
