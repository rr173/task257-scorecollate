基于 Go 实现的古典乐谱异文传抄证据复核 Web 项目，一款后端服务，完成乐谱片段解析、小节声部对齐、节拍连续性校验与异文来源支持评估，并发布不可变校勘版本。

# BENZHI 评测说明

本项目为纯后端 Go 服务，对外暴露 `/api` 前缀的 HTTP 接口，使用 SQLite 持久化，
支持进程关闭后重新打开同一数据库恢复全部业务数据。

## 标准命令

```bash
CGO_ENABLED=0 GOTOOLCHAIN=local go build ./...
CGO_ENABLED=0 GOTOOLCHAIN=local go vet   ./...
CGO_ENABLED=0 GOTOOLCHAIN=local go test  ./...
go run ./cmd/scorecollate --smoke-test
go run ./cmd/scorecollate --addr :8080 --db scorecollate.db
```

- `--addr`：HTTP 监听地址，默认 `:8080`
- `--db`：SQLite 数据库文件路径，默认 `task257-scorecollate.db`
- `--smoke-test`：不常驻；跑完端到端场景后关闭并重新打开数据库，退出码 0 表示通过

## 冒烟自测契约（--smoke-test）

创建临时数据库 → 写入校勘项目与来源谱系 → 导入两个乐谱片段并解析为小节 → 对齐后识别终止小节时值差异 → 裁决为讹写并确认 → 发布冻结校勘版本 → 关闭并重新打开数据库，校验异文数、已确认异文与冻结版本全部恢复后退出 0。

## Docker 构建与双架构验证

`Dockerfile` 与 `benzhi.Dockerfile` 内容完全一致。使用项目提供的
`build_benzhi_docker.sh` 构建评测镜像；Dockerfile 不声明端口，服务监听地址由
运行参数 `--addr` 指定。

```bash
./build_benzhi_docker.sh task257-scorecollate:amd64 linux/amd64
docker run --rm task257-scorecollate:amd64 --smoke-test

./build_benzhi_docker.sh task257-scorecollate:arm64 linux/arm64
docker run --rm task257-scorecollate:arm64 --smoke-test

docker run --rm -P task257-scorecollate:amd64 --addr :8080 --db ./app.db
```

## 核心 API（`/api` 前缀）

- 自检：`GET /api/health`、`POST /api/projects/{id}/selfcheck`
- 项目：`POST /api/projects`、`GET /api/projects`、`GET /api/projects/{id}`、`POST /api/projects/{id}/state`
- 来源谱系：`POST /api/projects/{id}/sources`、`GET /api/projects/{id}/sources`、`POST /api/sources/{id}/parent`
- 片段：`POST /api/projects/{id}/fragments`、`GET /api/projects/{id}/fragments`、`GET /api/fragments/{id}`、`POST /api/fragments/{id}/state`、`POST /api/fragments/{id}/parse`、`GET /api/fragments/{id}/measures`
- 对齐：`POST /api/projects/{id}/align`
- 异文：`GET /api/projects/{id}/variants`、`GET /api/variants/{id}`、`POST /api/variants/{id}/adjudicate`
- 校勘版本：`POST /api/projects/{id}/editions`、`GET /api/projects/{id}/editions`、`GET /api/editions/{id}`、`POST /api/editions/{id}/publish`、`POST /api/editions/{id}/supersede`

## 业务不变量

- 持久化：SQLite（modernc.org/sqlite，CGO 无关），重启同一数据库可恢复未发布版本与已冻结快照。
- 片段指纹幂等：相同指纹的重复导入被去重，不重复落库。
- 冻结版本不可变：对已冻结校勘版本重复发布会被拒绝。
- 来源谱系：父本链成环（含自环）在创建与重挂时均被拒绝。
