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

// 声明两声部但只抄了一个声部括号：必须判为不可辨识，而非补成空声部小节。
const rawVoiceMismatch = "M1 b4 v2 : [C4 E4]\nM2 b3 v2 : [D4 F4][B4 C5]\n"

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

func TestParseRejectsVoiceCountMismatch(t *testing.T) {
	svc := newTestService(t)
	p, err := svc.CreateProject("声部数校验", "")
	if err != nil {
		t.Fatal(err)
	}
	srcA, err := svc.CreateSource(p.ID, "A", "祖本", "", "")
	if err != nil {
		t.Fatal(err)
	}
	// 声明 v2 但 M1 只有一个声部括号 -> 拒绝解析，标记不可辨识
	fBad, err := svc.CreateFragment(p.ID, srcA.ID, "缺声部", rawVoiceMismatch)
	if err != nil {
		t.Fatal(err)
	}
	fBadRes, _, err := svc.ParseFragment(fBad.ID)
	if !errors.Is(err, model.ErrUnreadable) {
		t.Fatalf("声部数对不上应判为不可辨识，实际 err=%v", err)
	}
	if fBadRes == nil || fBadRes.State != model.FragUnreadable {
		t.Fatalf("片段应标记为 unreadable，实际 state=%v", fBadRes)
	}
	ms, err := svc.store.ListMeasures(fBad.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(ms) != 0 {
		t.Fatalf("不可辨识片段不应落库小节，实际 %d 条", len(ms))
	}

	// 括号数与声部数一致的片段仍可正常解析为可用小节
	srcB, err := svc.CreateSource(p.ID, "B", "传抄件", srcA.ID, "")
	if err != nil {
		t.Fatal(err)
	}
	fGood, err := svc.CreateFragment(p.ID, srcB.ID, "完整", rawGood)
	if err != nil {
		t.Fatal(err)
	}
	fGoodRes, measures, err := svc.ParseFragment(fGood.ID)
	if err != nil {
		t.Fatalf("括号数正确的片段应解析成功，实际 err=%v", err)
	}
	if fGoodRes.State != model.FragAligned {
		t.Fatalf("片段应标记为 aligned，实际 %s", fGoodRes.State)
	}
	if len(measures) != 2 {
		t.Fatalf("期望 2 个小节，实际 %d", len(measures))
	}
}
