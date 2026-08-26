package model

import "time"

// Source 是乐谱的传抄来源（手稿谱系节点）。
type Source struct {
	ID          string
	ProjectID   string
	Siglum      string // 谱号/代号，如 A、B、C
	Title       string
	ParentID    *string // 传抄父本；nil 表示祖本
	Description string
	CreatedAt   time.Time
}
