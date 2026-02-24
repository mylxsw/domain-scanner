# Namecheap Domain Probe - Web 版

一个现代化的 Web 应用，用于探测 Namecheap 域名可用性和价格信息。

## 项目结构

```
.
├── backend/          # Go + Gin 后端 API
├── frontend/         # React + TypeScript + Vite 前端
├── Makefile          # 常用命令
└── README.md
```

## 快速开始

### 环境要求

- Go 1.23+
- Node.js 18+
- Namecheap API 账号

### 配置

设置 Namecheap API 环境变量：

```bash
export NAMECHEAP_API_USER="your_api_user"
export NAMECHEAP_API_KEY="your_api_key"
export NAMECHEAP_CLIENT_IP="your_whitelisted_ip"
```

### 启动后端

```bash
cd backend
go mod tidy
go run ./cmd/server

# 或使用自定义端口
go run ./cmd/server -host 0.0.0.0 -port 8080
```

后端将在 `http://localhost:8080` 启动。

### 启动前端

```bash
cd frontend
npm install
npm run dev
```

前端将在 `http://localhost:3000` 启动，并自动代理 API 请求到后端。

## 功能特性

### 后端 API

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/api/health` | 健康检查 |
| GET | `/api/tlds` | 获取 TLD 列表 |
| POST | `/api/probe` | 启动域名探测任务 |
| GET | `/api/probe/:id` | 获取任务状态 |
| GET | `/api/probe/:id/results` | 获取探测结果 |
| GET | `/api/probe/:id/stream` | SSE 实时流 |

### 前端功能

- **域名搜索**: 输入关键词，选择 TLD 模式（全部/热门/低价）
- **实时进度**: SSE 实时显示探测进度
- **结果展示**:
  - 可排序、筛选的结果表格
  - 价格分布图表
  - 可注册/已注册/出错分类统计
- **数据导出**: 支持 CSV 导出

## 技术栈

### 后端
- **Go 1.23**: 高性能后端语言
- **Gin**: Web 框架
- **UUID**: 任务标识
- **CORS**: 跨域支持

### 前端
- **React 18**: UI 框架
- **TypeScript**: 类型安全
- **Vite**: 构建工具
- **Tailwind CSS**: 样式
- **shadcn/ui**: 组件库
- **React Query**: 状态管理
- **Recharts**: 图表
- **React Router**: 路由

## API 使用示例

### 启动探测任务

```bash
curl -X POST http://localhost:8080/api/probe \
  -H "Content-Type: application/json" \
  -d '{"word": "example", "tld_mode": "mainstream-only"}'
```

响应：
```json
{
  "task_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "running",
  "word": "example"
}
```

### 获取实时流

```bash
curl http://localhost:8080/api/probe/550e8400-e29b-41d4-a716-446655440000/stream
```

### 获取结果

```bash
curl http://localhost:localhost:8080/api/probe/550e8400-e29b-41d4-a716-446655440000/results
```

## 开发说明

### 后端开发

```bash
cd backend

# 运行测试
go test ./...

# 构建
go build -o bin/server ./cmd/server
```

### 前端开发

```bash
cd frontend

# 安装依赖
npm install

# 开发模式
npm run dev

# 构建
npm run build

# 预览生产构建
npm run preview
```

### 使用 Makefile

```bash
# 安装所有依赖
make install

# 构建项目
make build

# 开发模式（需要两个终端）
make run-backend
make run-fronten
```

## 原始命令行工具

原有的命令行工具仍然可用：

```bash
# 使用原有代码
go run . probe mybrand --outdir ./out

# 启动 SSE 服务
go run . serve --host 0.0.0.0 --port 8000
```

## License

MIT
