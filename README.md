# Domain Scanner

一个用于批量探测域名可用性与价格信息的 Web 系统。

- 后端：Go + Gin + SQLite
- 前端：React + Vite
- 运行方式：本地开发 / Docker / Docker Compose

## 功能概览

- 域名批量探测（SSE 实时进度）
- 结果分类：可注册 / 已注册 / 错误
- 结果排序、筛选与导出 CSV
- 历史任务记录（首页可回看）
- 数据持久化（SQLite）

## 目录结构

```text
.
├── backend/                 # Go 后端
├── frontend/                # React 前端
├── docker-compose.yml       # Compose 一键运行
├── Dockerfile               # 多阶段镜像构建
├── .env.example             # 环境变量示例
└── Makefile                 # 常用命令
```

## 环境变量

Namecheap API 相关变量（必填）：

```bash
NAMECHEAP_API_USER=your_api_user
NAMECHEAP_API_KEY=your_api_key
NAMECHEAP_CLIENT_IP=your_whitelisted_ip
```

## 一键运行（推荐）

### 1) Docker Compose

```bash
cp .env.example .env
# 编辑 .env 填入真实值

docker compose up -d --build
```

访问：<http://localhost:8080>

停止：

```bash
docker compose down
```

查看日志：

```bash
docker compose logs -f
```

## 数据持久化说明

Compose 已配置宿主机目录挂载：

- 容器内：`/app/out`
- 宿主机：`./data/out`

持久化数据包括：

- `probe.sqlite`（任务、结果、TLD 清单）
- 每个任务输出目录（`results.jsonl` / `results.csv` / `report.md` / 缓存文件）

即使容器重建，历史数据仍保留在 `./data/out`。

## 本地开发

### 1) 启动后端

```bash
cd backend
go mod tidy
go run ./cmd/server
```

默认监听：`http://127.0.0.1:8080`

### 2) 启动前端（开发模式）

```bash
cd frontend
npm install
npm run dev
```

默认监听：`http://localhost:3000`（代理到后端 API）

## 后端托管前端静态文件（单服务）

```bash
cd frontend && npm install && npm run build
cd ../backend && go run ./cmd/server -web-dir ../frontend/dist
```

访问：<http://localhost:8080>

## Docker（非 Compose）

构建镜像：

```bash
docker build -t domain-scanner:latest .
```

运行容器：

```bash
docker run --rm -p 8080:8080 \
  -e NAMECHEAP_API_USER \
  -e NAMECHEAP_API_KEY \
  -e NAMECHEAP_CLIENT_IP \
  -v "$(pwd)/data/out:/app/out" \
  domain-scanner:latest
```

## Makefile 常用命令

```bash
make install
make build
make test
make run-backend
make run-frontend

make docker-build
make docker-run
make docker-compose-up
make docker-compose-down
make docker-compose-logs
```

## API 概览

- `GET /api/health`
- `GET /api/tlds`
- `POST /api/probe`
- `GET /api/probe`（历史任务列表）
- `GET /api/probe/:id`
- `GET /api/probe/:id/results`
- `GET /api/probe/:id/stream`

## 常见问题

1. 前端页面打不开  
确认你访问的是 `http://localhost:8080`，并且容器日志里有 `Frontend static hosting enabled`。

2. 没有历史记录  
检查 `./data/out/probe.sqlite` 是否存在；删除容器不会删除该目录，但删除目录会丢失历史。

3. Namecheap 调用失败  
检查 `.env` 的 API 变量和白名单 IP 是否正确。

4. Linux + Docker 启动报错：`unable to open database file: out of memory (14)`  
这通常不是内存问题，而是挂载目录不可写（SQLite `CANTOPEN(14)`）。请在宿主机执行：

```bash
mkdir -p ./data/out
sudo chown -R 10001:10001 ./data/out
chmod 775 ./data/out
```

然后重启容器：

```bash
docker compose down
docker compose up -d --build
```
