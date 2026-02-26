.PHONY: all build run-dev run-backend run-frontend clean install test help docker-build docker-run docker-compose-up docker-compose-down docker-compose-logs

# 默认目标
all: install build

# 安装依赖
install:
	@echo "Installing backend dependencies..."
	cd backend && go mod tidy
	@echo "Installing frontend dependencies..."
	cd frontend && npm install

# 构建项目
build:
	@echo "Building backend..."
	cd backend && go build -o bin/server ./cmd/server
	@echo "Building frontend..."
	cd frontend && npm run build

# 开发模式运行（需要两个终端）
run-backend:
	cd backend && go run ./cmd/server

run-frontend:
	cd frontend && npm run dev

# 清理构建文件
clean:
	rm -rf backend/bin
	rm -rf frontend/dist

# 运行测试
test:
	cd backend && go test ./...

# Docker 构建
docker-build:
	docker build -t domain-scanner:latest .

docker-run: docker-build
	docker run --rm -p 8080:8080 \
		-e NAMECHEAP_API_USER \
		-e NAMECHEAP_API_KEY \
		-e NAMECHEAP_CLIENT_IP \
		domain-scanner:latest

docker-compose-up:
	docker compose up -d --build

docker-compose-down:
	docker compose down

docker-compose-logs:
	docker compose logs -f

# 帮助信息
help:
	@echo "Available targets:"
	@echo "  make install      - Install all dependencies"
	@echo "  make build        - Build backend and frontend"
	@echo "  make run-backend  - Run backend in development mode"
	@echo "  make run-frontend - Run frontend in development mode"
	@echo "  make test         - Run tests"
	@echo "  make clean        - Clean build files"
	@echo "  make docker-build - Build Docker image"
	@echo "  make docker-run   - Run Docker container"
	@echo "  make docker-compose-up   - Start app by docker compose"
	@echo "  make docker-compose-down - Stop app by docker compose"
	@echo "  make docker-compose-logs - Tail docker compose logs"
	@echo "  make help         - Show this help message"
