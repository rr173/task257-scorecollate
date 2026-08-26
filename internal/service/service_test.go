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

// TestAlignmentDedupsSameLineageSupport 验证校勘独立来源支持按祖本谱系去重：
// 多份同源传抄副本持相同读法只计一次，不会被误判为有效变体；而来自不同祖本
// 谱系的副本才提供独立支持，保留有效变体判定。
func TestAlignmentDedupsSameLineageSupport(t *testing.T) {
	svc := newTestService(t)
	p, err := svc.CreateProject("同源副本去重", "")
	if err != nil {
		t.Fatal(err)
	}
	// 三条独立祖本谱系：rootX、rootY、rootZ。rootX 下还有传抄副本 c1/c2。
	srcRootX, _ := svc.CreateSource(p.ID, "X", "祖本X", "", "")
	srcC1, _ := svc.CreateSource(p.ID, "c1", "传抄", srcRootX.ID, "")
	srcC2, _ := svc.CreateSource(p.ID, "c2", "传抄", srcC1.ID, "")
	srcRootY, _ := svc.CreateSource(p.ID, "Y", "祖本Y", "", "")
	srcRootZ, _ := svc.CreateSource(p.ID, "Z", "祖本Z", "", "")

	// M1 声部0：A 读 X、B 读 Z，构成一处异文。
	// rootX 谱系内的 c1/c2 都读 X，但与 A 同源，不应计入独立支持。
	// rootZ 谱系读 X，是为 A 的读法提供独立支持的不同祖本谱系。
	rawAX := "M1 b4 v1 : [X]\n"
	rawBZ := "M1 b4 v1 : [Z]\n"
	rawX := "M1 b4 v1 : [X]\n"

	fA, _ := svc.CreateFragment(p.ID, srcRootX.ID, "A", rawAX)
	fB, _ := svc.CreateFragment(p.ID, srcRootY.ID, "B", rawBZ)
	fC1, _ := svc.CreateFragment(p.ID, srcC1.ID, "c1", rawX) // 与 A 同谱系
	fC2, _ := svc.CreateFragment(p.ID, srcC2.ID, "c2", rawX) // 与 A 同谱系
	fZ, _ := svc.CreateFragment(p.ID, srcRootZ.ID, "z", rawX) // 不同祖本谱系，独立支持

	for _, fid := range []string{fA.ID, fB.ID, fC1.ID, fC2.ID, fZ.ID} {
		if _, _, err := svc.ParseFragment(fid); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := svc.AlignProject(p.ID); err != nil {
		t.Fatal(err)
	}
	variants, err := svc.ListVariants(p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(variants) == 0 {
		t.Fatal("期望产生异文候选")
	}
	v := variants[0]
	// 仅 rootZ 一支不同祖本谱系独立支持读法 X；rootX 谱系内 c1/c2 与 A 同源不计入。
	if v.SupportCount != 1 {
		t.Fatalf("同源副本应去重，期望独立支持数 1，实际 %d", v.SupportCount)
	}
	if v.DetectedKind != model.VarValidVariant {
		t.Fatalf("存在 1 支不同祖本谱系独立支持应判有效变体，实际 %s", v.DetectedKind)
	}
}
