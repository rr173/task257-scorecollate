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

// rawDropped 丢掉了终止小节 M2，模拟少抄一小节的传抄本。
const rawDropped = "M1 b4 v2 : [C4 E4][G4 A4]\n"

func TestAlignmentDetectsDroppedMeasureAsError(t *testing.T) {
	svc := newTestService(t)
	p, err := svc.CreateProject("缺抄校勘", "整小节缺抄应识别为讹写")
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
	fB, err := svc.CreateFragment(p.ID, srcB.ID, "B", rawDropped)
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
		t.Fatalf("整小节缺抄应产出异文候选，实际 %d", n)
	}
	variants, err := svc.ListVariants(p.ID)
	if err != nil {
		t.Fatal(err)
	}
	// 缺抄的终止小节（M2）每个声部都应被判为讹写
	var dropped *model.Variant
	for _, v := range variants {
		if v.MeasureNumber == 2 {
			dropped = v
			break
		}
	}
	if dropped == nil {
		t.Fatalf("未找到 M2 缺抄异文，variants=%v", variants)
	}
	if dropped.DetectedKind != model.VarError {
		t.Fatalf("整小节缺抄应初判为讹写，实际 %s", dropped.DetectedKind)
	}
	if dropped.ContentA == "" || dropped.ContentB != "" {
		t.Fatalf("缺抄异文应有参考读法与空缺对照，实际 A=%q B=%q", dropped.ContentA, dropped.ContentB)
	}
}

func TestAlignmentBothHaveMeasureKeepsBeatBreak(t *testing.T) {
	// 双方都有该小节但拍数不同时仍走原有节拍断裂判定。
	svc := newTestService(t)
	p, _ := svc.CreateProject("节拍断裂保留", "")
	srcA, _ := svc.CreateSource(p.ID, "A", "A", "", "")
	srcB, _ := svc.CreateSource(p.ID, "B", "B", srcA.ID, "")
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
	variants, err := svc.ListVariants(p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(variants) == 0 {
		t.Fatal("期望节拍断裂异文")
	}
	for _, v := range variants {
		if v.DetectedKind != model.VarError {
			t.Fatalf("节拍断裂应初判为讹写，实际 %s", v.DetectedKind)
		}
	}
}
