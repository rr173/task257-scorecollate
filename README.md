# task257-scorecollate 古典乐谱异文传抄证据复核台

面向音乐文献研究者的乐谱校勘后端服务：接收多份传抄手稿的乐谱片段，解析为小节/声部，按节拍与来源谱系对齐识别音符异文，裁决讹写或有效变体，并发布不可变的校勘版本。

## 技术栈

- Go 1.26.3（`GOTOOLCHAIN=local`、`CGO_ENABLED=0`）
- SQLite（`modernc.org/sqlite v1.52.0`，纯 Go 驱动，离线可构建）
- 标准库 `net/http` JSON API + 轻量校勘视图

## 构建与自检

```bash
export GO_BIN=$(command -v go)
CGO_ENABLED=0 GOTOOLCHAIN=local $GO_BIN build ./...
CGO_ENABLED=0 GOTOOLCHAIN=local $GO_BIN vet   ./...
CGO_ENABLED=0 GOTOOLCHAIN=local $GO_BIN test  ./...
$GO_BIN run ./cmd/scorecollate --smoke-test   # 离线端到端自检（退出码 0 即通过）
```

## 启动服务

```bash
$GO_BIN run ./cmd/scorecollate --addr=:8080 --db=task257-scorecollate.db
```

打开 `http://localhost:8080/` 查看并排校勘视图；全部数据操作走 `/api` 前缀的 JSON 接口（详见 BENZHI_README.md 的 API 清单）。

## 记谱格式（片段 raw 字段）

每行一个小节：`M<小节号> b<拍数> v<声部数> : [声部0音符][声部1音符]...`，`#` 开头为注释。

```
M1 b4 v2 : [C4 E4][G4 A4]
M2 b3 v2 : [D4 F4][B4 C5]
```

解析失败（格式非法）的片段自动标记为 `unreadable`，可修正后重新解析。

## 领域模型

| 实体 | 状态机 | 说明 |
| --- | --- | --- |
| 校勘项目 | arranging → awaiting_alignment → awaiting_adjudication → published → sealed | 封存后不可再修改 |
| 来源（谱系） | parent 链 | 父本可在建源时指定，也可通过 `POST /api/sources/{id}/parent` 调整；成环拒绝 |
| 乐谱片段 | pending_parse → aligned / unreadable → excluded | 同片段指纹幂等，重复导入不产生重复小节 |
| 小节 | number 唯一 | 含拍数、声部数与各声部音符串 |
| 异文 | candidate → error / variant / insufficient → confirmed | 裁决使用乐观版本锁，冲突返回 409 |
| 校勘版本 | draft → shared / frozen / superseded | 冻结版本不可变，重复发布返回 403 |

## 校验要点

- 同一异文裁决带版本号，防并发覆盖。
- 冻结版本保留当时来源集合，追加新来源只能开新版本。
- 来源谱系环（含自环）在创建与重挂父本时均被拒绝。
- 对齐以小节数最多的已对齐片段为参考，节拍断裂判讹写、独立来源支持判有效变体、无支持判证据不足。
