package demo

import (
	"fmt"
	"os"

	"task257-scorecollate/internal/model"
	"task257-scorecollate/internal/service"
	"task257-scorecollate/internal/store"
)

// RunSmokeTest 执行离线端到端自检：
// 创建项目 -> 建立来源谱系（B 传抄自 A）-> 导入两个片段（B 在终止小节出现时值差异）
// -> 解析为小节 -> 对齐生成异文候选 -> 裁决为讹写并确认 -> 发布冻结校勘版本
// -> 关闭并重新打开 SQLite 校验重启恢复 -> 校验冻结版本不可重复发布，最终退出码 0。
func RunSmokeTest(dbPath string) error {
	// 清理可能残留的数据库与 WAL 文件，保证幂等可重跑
	_ = os.Remove(dbPath)
	_ = os.Remove(dbPath + "-wal")
	_ = os.Remove(dbPath + "-shm")

	db, err := store.Open(dbPath)
	if err != nil {
		return err
	}
	if err := store.Migrate(db); err != nil {
		_ = db.Close()
		return err
	}
	svc := service.New(store.New(db))
	closeDB := func(e error) error {
		_ = db.Close()
		return e
	}

	// 1. 创建校勘项目
	p, err := svc.CreateProject("G大调奏鸣曲第一乐章校勘", "两份手稿终止小节时值差异复核")
	if err != nil {
		return closeDB(err)
	}

	// 2. 来源谱系：A 为祖本，B 传抄自 A
	srcA, err := svc.CreateSource(p.ID, "A", "手稿 A（祖本）", "", "")
	if err != nil {
		return closeDB(err)
	}
	srcB, err := svc.CreateSource(p.ID, "B", "手稿 B（传抄）", srcA.ID, "")
	if err != nil {
		return closeDB(err)
	}

	// 3. 两个片段：B 在终止小节（M4）把 4 拍写成 3 拍，构成节拍断裂
	rawA := "M1 b4 v2 : [C4 E4][G4 A4]\nM2 b3 v2 : [D4 F4][B4 C5]\nM3 b4 v2 : [E4 G4][C5 D5]\nM4 b4 v2 : [F4 A4][D5 E5]\n"
	rawB := "M1 b4 v2 : [C4 E4][G4 A4]\nM2 b3 v2 : [D4 F4][B4 C5]\nM3 b4 v2 : [E4 G4][C5 D5]\nM4 b3 v2 : [F4 A4][D5]\n"
	fragA, err := svc.CreateFragment(p.ID, srcA.ID, "片段 A", rawA)
	if err != nil {
		return closeDB(err)
	}
	fragB, err := svc.CreateFragment(p.ID, srcB.ID, "片段 B", rawB)
	if err != nil {
		return closeDB(err)
	}
	if _, _, err := svc.ParseFragment(fragA.ID); err != nil {
		return closeDB(err)
	}
	if _, _, err := svc.ParseFragment(fragB.ID); err != nil {
		return closeDB(err)
	}

	// 4. 对齐生成异文候选
	n, err := svc.AlignProject(p.ID)
	if err != nil {
		return closeDB(err)
	}
	if n < 1 {
		return closeDB(fmt.Errorf("对齐未产生异文候选"))
	}
	variants, err := svc.ListVariants(p.ID)
	if err != nil {
		return closeDB(err)
	}

	// 5. 裁决：先按初判类型裁决，再确认
	for _, v := range variants {
		decision := model.VarError
		if v.DetectedKind == model.VarValidVariant {
			decision = model.VarValidVariant
		} else if v.DetectedKind == model.VarInsufficient {
			decision = model.VarInsufficient
		}
		if _, err := svc.AdjudicateVariant(v.ID, decision, "时值差异复核", "smoke-editor"); err != nil {
			return closeDB(err)
		}
		if _, err := svc.AdjudicateVariant(v.ID, model.VarConfirmed, "确认纳入校勘", "smoke-editor"); err != nil {
			return closeDB(err)
		}
	}

	// 6. 发布冻结校勘版本
	ed, err := svc.CreateEdition(p.ID, "校勘版第一稿")
	if err != nil {
		return closeDB(err)
	}
	if _, err := svc.PublishEdition(ed.ID); err != nil {
		return closeDB(err)
	}

	// 7. 关闭前快照
	snapBefore, err := svc.SelfCheck(p.ID)
	if err != nil {
		return closeDB(err)
	}

	// 8. 关闭并重新打开数据库，校验重启恢复
	if err := db.Close(); err != nil {
		return err
	}
	db2, err := store.Open(dbPath)
	if err != nil {
		return err
	}
	defer func() { _ = db2.Close() }()
	svc2 := service.New(store.New(db2))
	snapAfter, err := svc2.SelfCheck(p.ID)
	if err != nil {
		return err
	}
	if snapBefore["variants"] != snapAfter["variants"] {
		return fmt.Errorf("重启后异文数不一致: before=%v after=%v", snapBefore["variants"], snapAfter["variants"])
	}
	if confirmed, ok := snapAfter["confirmed"].(int); !ok || confirmed == 0 {
		return fmt.Errorf("重启后没有已确认异文")
	}

	// 9. 冻结版本不可重复发布（不可变性校验）
	if _, err := svc2.PublishEdition(ed.ID); err == nil {
		return fmt.Errorf("冻结版本应拒绝重复发布")
	}
	return nil
}
