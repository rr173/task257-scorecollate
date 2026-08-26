package service

import (
	"errors"
	"path/filepath"
	"testing"

	"task257-scorecollate/internal/model"
	"task257-scorecollate/internal/store"
)

func newTestService(t *testing.T) *Service {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := store.Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return New(store.New(db))
}

const rawGood = "M1 b4 v2 : [C4 E4][G4 A4]\nM2 b3 v2 : [D4 F4][B4 C5]\n"
const rawBreak = "M1 b4 v2 : [C4 E4][G4 A4]\nM2 b2 v2 : [D4 F4][B4]\n"

func TestAlignmentDetectsBeatBreakAndFreeze(t *testing.T) {
	svc := newTestService(t)
	p, err := svc.CreateProject("测试校勘", "节拍断裂检测")
	if err != nil {
		t.Fatal(err)
	}
	srcA, err := svc.CreateSource(p.ID, "A", "祖本", "", "")
	if err != nil {
		t.Fatal(err)
	}
	srcB, err := svc.CreateSource(p.ID, "B", "传抄本", srcA.ID, "")
	if err != nil {
		t.Fatal(err)
	}
	fA, err := svc.CreateFragment(p.ID, srcA.ID, "A", rawGood)
	if err != nil {
		t.Fatal(err)
	}
	fB, err := svc.CreateFragment(p.ID, srcB.ID, "B", rawBreak)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.ParseFragment(fA.ID); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.ParseFragment(fB.ID); err != nil {
		t.Fatal(err)
	}
	n, err := svc.AlignProject(p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if n < 1 {
		t.Fatalf("期望至少 1 个异文候选，实际 %d", n)
	}
	variants, err := svc.ListVariants(p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if variants[0].DetectedKind != model.VarError {
		t.Fatalf("节拍断裂应初判为讹写，实际 %s", variants[0].DetectedKind)
	}
	if _, err := svc.AdjudicateVariant(variants[0].ID, model.VarError, "复核", "tester"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AdjudicateVariant(variants[0].ID, model.VarConfirmed, "确认", "tester"); err != nil {
		t.Fatal(err)
	}
	ed, err := svc.CreateEdition(p.ID, "校勘版")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.PublishEdition(ed.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.PublishEdition(ed.ID); err == nil {
		t.Fatal("冻结版本应拒绝重复发布")
	}
}

func TestVersionLockRejectsStaleUpdate(t *testing.T) {
	svc := newTestService(t)
	p, _ := svc.CreateProject("并发冲突", "")
	srcA, _ := svc.CreateSource(p.ID, "A", "A", "", "")
	srcB, _ := svc.CreateSource(p.ID, "B", "B", "", "")
	fA, _ := svc.CreateFragment(p.ID, srcA.ID, "A", rawGood)
	fB, _ := svc.CreateFragment(p.ID, srcB.ID, "B", rawBreak)
	if _, _, err := svc.ParseFragment(fA.ID); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.ParseFragment(fB.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AlignProject(p.ID); err != nil {
		t.Fatal(err)
	}
	variants, _ := svc.ListVariants(p.ID)
	v := variants[0]
	// 第一次裁决成功后版本 +1
	if _, err := svc.AdjudicateVariant(v.ID, model.VarError, "r1", "e1"); err != nil {
		t.Fatal(err)
	}
	// 携带过期版本号（version=1）直接更新必须返回版本冲突
	if _, err := svc.store.UpdateVariantState(v.ID, model.VarConfirmed, 1); !errors.Is(err, model.ErrConflict) {
		t.Fatalf("期望版本冲突，实际 %v", err)
	}
	// 状态机非法迁移：error 态不能跳到 variant 态
	if _, err := svc.AdjudicateVariant(v.ID, model.VarValidVariant, "r2", "e2"); err == nil {
		t.Fatal("error 态不能迁移到 variant 态")
	}
}

func TestSourceGenealogyRejectsCycle(t *testing.T) {
	svc := newTestService(t)
	p, _ := svc.CreateProject("谱系环", "")
	srcA, _ := svc.CreateSource(p.ID, "A", "A", "", "")
	srcB, err := svc.CreateSource(p.ID, "B", "B", srcA.ID, "")
	if err != nil {
		t.Fatal(err)
	}
	// 把祖本 A 的父本指向后代 B -> 成环，必须拒绝
	if _, err := svc.ReparentSource(srcA.ID, srcB.ID); !errors.Is(err, model.ErrSelfCycle) {
		t.Fatalf("期望谱系环错误，实际 %v", err)
	}
	// 同一来源自指同样拒绝
	if _, err := svc.ReparentSource(srcA.ID, srcA.ID); !errors.Is(err, model.ErrSelfCycle) {
		t.Fatalf("期望自环错误，实际 %v", err)
	}
}

// TestPublishFreezeExcludesUnadjudicated 验证：发布冻结版本时只收录已确认异文，
// 尚未裁决的候选不得进入冻结快照。
func TestPublishFreezeExcludesUnadjudicated(t *testing.T) {
	svc := newTestService(t)
	p, err := svc.CreateProject("未裁决候选不入快照", "")
	if err != nil {
		t.Fatal(err)
	}
	srcA, err := svc.CreateSource(p.ID, "A", "祖本", "", "")
	if err != nil {
		t.Fatal(err)
	}
	srcB, err := svc.CreateSource(p.ID, "B", "传抄本", srcA.ID, "")
	if err != nil {
		t.Fatal(err)
	}
	// 两处差异：M1 声部0 音符不同（节拍合法 -> 初判证据不足），M2 节拍断裂（-> 初判讹写）。
	rawA := "M1 b4 v2 : [C4 E4][G4 A4]\nM2 b3 v2 : [D4 F4][B4 C5]\n"
	rawB := "M1 b4 v2 : [C#4 E4][G4 A4]\nM2 b2 v2 : [D4 F4][B4]\n"
	fA, err := svc.CreateFragment(p.ID, srcA.ID, "A", rawA)
	if err != nil {
		t.Fatal(err)
	}
	fB, err := svc.CreateFragment(p.ID, srcB.ID, "B", rawB)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.ParseFragment(fA.ID); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.ParseFragment(fB.ID); err != nil {
		t.Fatal(err)
	}
	if n, err := svc.AlignProject(p.ID); err != nil || n < 2 {
		t.Fatalf("期望至少 2 个异文候选，实际 %d err=%v", n, err)
	}
	variants, err := svc.ListVariants(p.ID)
	if err != nil {
		t.Fatal(err)
	}
	// 只确认第一处异文，其余保留为未裁决候选。
	confirmed := variants[0]
	if _, err := svc.AdjudicateVariant(confirmed.ID, confirmed.DetectedKind, "仅裁决第一处", "tester"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AdjudicateVariant(confirmed.ID, model.VarConfirmed, "确认纳入校勘", "tester"); err != nil {
		t.Fatal(err)
	}
	ed, err := svc.CreateEdition(p.ID, "校勘版")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.PublishEdition(ed.ID); err != nil {
		t.Fatal(err)
	}
	links, err := svc.store.ListEditionVariants(ed.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 1 {
		t.Fatalf("冻结快照应只含 1 个已确认异文，实际 %d", len(links))
	}
	if links[0].VariantID != confirmed.ID || !links[0].Included {
		t.Fatalf("冻结快照收录项与已确认异文不符: %+v", links[0])
	}
	// 所有未裁决候选都不应出现在快照中。
	for _, v := range variants[1:] {
		for _, l := range links {
			if l.VariantID == v.ID {
				t.Fatalf("未裁决候选 %s 不应进入冻结快照", v.ID)
			}
		}
	}
}

