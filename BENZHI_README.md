# 本质评测环境说明

## 项目

- 项目编号：`zgw-gowork-0065`
- 项目名称：半导体掩模版变更审核
- 项目说明：半导体掩模版变更审核本地服务

## 固定环境

- Go toolchain：`go1.26.5`
- go.mod language version：`go 1.21`
- GOTOOLCHAIN：`local`
- 支持平台：`linux/amd64`、`linux/arm64`
- Docker 基础镜像：`golang:1.26.5-bookworm`
- Docker manifest：`golang@sha256:53eeac89074db483fdf0ab3be1df32bf6e47562263d2d0d6baa7f26acb4957dd`

## 构建

```bash
./build_benzhi_docker.sh zgw-gowork-0065:benzhi-amd64 linux/amd64
./build_benzhi_docker.sh zgw-gowork-0065:benzhi-arm64 linux/arm64
```

## 运行

```bash
docker run --rm -it --network none zgw-gowork-0065:benzhi-amd64 bash
```

## 容器内验证

```bash
go version
go env GOTOOLCHAIN GOPROXY GOMODCACHE GOCACHE
go test ./...
go vet ./...
go build ./...
```

---

# 项目 README 同步内容

# 半导体掩模版变更审核服务

本服务用于管理半导体制造中的掩模版变更审核流程，支持产品层级、掩模版版本、变更申请、验证批次和审核结论的管理。核心状态链为：起草 → 送审 → 试投 → 验证 → 启用。

## 主要功能

- 管理产品层级与掩模版版本
- 提交变更申请并跟踪状态
- 创建验证批次并记录验证结果
- 执行审核并启用/重起草版本
- 强制业务约束：同一产品层级仅有一个启用版本、验证批次绑定送审版本、关键验证项未通过禁止启用、重新起草使旧审核结论失效
- 提供版本差异摘要和待审核清单导出

## 运行

```bash
go run ./cmd/server
```

## API 摘要

- `POST /api/product-levels`：创建产品层级
- `GET /api/product-levels`：列出产品层级
- `POST /api/mask-versions`：创建掩模版版本（草稿）
- `GET /api/mask-versions`：按产品层级查询版本
- `POST /api/change-requests`：创建变更申请
- `GET /api/change-requests`：列出申请
- `POST /api/change-requests/{id}/submit`：送审
- `POST /api/verification-batches`：创建验证批次
- `POST /api/verification-batches/{id}/result`：提交验证结果
- `POST /api/change-requests/{id}/audit`：审核结论
- `POST /api/mask-versions/{id}/enable`：启用版本
- `POST /api/mask-versions/{id}/redraft`：重新起草
- `GET /api/summary/diff?level=X`：按产品层级版本差异摘要
- `GET /api/export/pending-audit`：导出待审核清单 CSV

## 存储

使用本地文件存储（JSON），支持原子写入和启动恢复。可通过环境变量 `MASKREVIEW_STORE_PATH` 配置存储路径，默认为 `./data/store.json`。

## 测试

```bash
go test ./...
go test -race ./...
go vet ./...
go build ./...
```
