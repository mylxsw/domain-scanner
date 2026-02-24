# Namecheap Domain Probe (Go 版本)

这是一个用 Go 重写的 Namecheap 域名探测工具，原 Python 版本的完整功能移植。

## 功能特性

- 获取全部 TLD 列表（`namecheap.domains.getTldList`）
- 获取 1 年注册价格（`namecheap.users.getPricing`，REGISTER）
- 批量检查可用性与溢价信息（`namecheap.domains.check`，最多 50/次）
- **流式输出**：每批次返回后逐条输出（JSON Lines 或 SSE）
- **限流处理**：内置节流 + 429/"Too many requests" 自动退避重试，直到处理完
- 最终生成 **按价格从低到高** 的详尽报告（Markdown + CSV）

## 安装

需要 Go 1.23+：

```bash
# 克隆或下载代码后
cd namecheap-domain-probe

# 下载依赖
go mod tidy

# 编译
go build -o namecheap-domain-probe .

# 或者直接运行
go run .
```

## 配置（必须）

Namecheap API 需要：ApiUser、ApiKey、UserName、ClientIp（白名单 IPv4）。

设置环境变量：

```bash
export NAMECHEAP_API_USER="yourApiUser"
export NAMECHEAP_API_KEY="yourApiKey"
export NAMECHEAP_USERNAME="yourUserName"   # 通常与 ApiUser 相同
export NAMECHEAP_CLIENT_IP="1.2.3.4"       # 需在 Namecheap 控制台白名单
# 可选：使用 Sandbox
export NAMECHEAP_API_BASE="https://api.sandbox.namecheap.com/xml.response"
# 生产环境默认： https://api.namecheap.com/xml.response
```

## 命令行使用（JSON Lines 流式输出）

```bash
./namecheap-domain-probe probe mybrand \
  --outdir ./out \
  --rate-per-min 45 \
  --max-batch 50
```

参数说明：
- `--outdir`：输出目录（默认 `./out`）
- `--rate-per-min`：本地节流上限（默认 45/min，低于官方 50/min，留余量）
- `--max-batch`：domains.check 每次最多检查数量（默认 50，官方限制）
- `--tld-mode`：`all` / `api-registerable-only` / `mainstream-only`
- `--cache-ttl-hours`：TLD 列表/价格缓存 TTL（默认 24h）

运行过程中会逐行输出 JSON（适合管道处理）：
- `type=progress`：单个域名的检测结果（可用性、价格、错误等）
- `type=summary`：全部结束后的统计

输出文件（以 `{word}_` 为前缀，方便区分不同单词的结果）：
- `out/{word}_report.md`：最终详尽报告（按价格升序）
- `out/{word}_results.csv`：结构化结果（可用于 Excel/BI）
- `out/{word}_results.jsonl`：全量过程输出存档

## 启动 SSE 服务端

```bash
./namecheap-domain-probe serve --host 0.0.0.0 --port 8000
```

访问：
- 健康检查：`GET /health`
- SSE 流：`GET /probe/{word}?rate_per_min=45&tld_mode=all`

示例（curl）：

```bash
curl -N http://127.0.0.1:8000/probe/mybrand
```

## 与 Python 版本的差异

| 特性 | Python 版本 | Go 版本 |
|------|------------|---------|
| 运行时依赖 | Python 3.10+ | Go 1.23+（编译后无依赖） |
| 并发模型 | asyncio | Goroutines + Channels |
| HTTP 框架 | FastAPI + uvicorn | 标准库 net/http |
| XML 解析 | xml.etree.ElementTree | bevik/etree |
| 打包 | 需要 venv + pip | 单二进制文件 |
| 启动速度 | 较慢（解释型） | 快（编译型） |

## 项目结构

```
.
├── main.go          # CLI 入口和 HTTP 服务器
├── config.go        # 配置管理
├── namecheap.go     # Namecheap API 客户端
├── retry.go         # 限流器和重试机制
├── xmlutil.go       # XML 解析工具
├── probe.go         # 核心探测逻辑
├── go.mod           # Go 模块定义
└── README_GO.md     # 本文件
```

## 免责声明

- Namecheap 官方限制：一般为 **50/min、700/hour、8000/day（按 key）**。项目会尽量避免触发，但也可能因并发/网络波动被限流。
- "价格"以 `users.getPricing` 返回的 1 年 REGISTER 价格为准；溢价域名以 `domains.check` 的 PremiumRegistrationPrice 为准。
