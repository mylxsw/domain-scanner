FROM node:20-alpine AS frontend-builder
WORKDIR /src/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

FROM golang:1.23-alpine AS backend-builder
WORKDIR /src/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/server ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata && adduser -D -u 10001 app
WORKDIR /app
COPY --from=backend-builder /out/server /app/server
COPY --from=frontend-builder /src/frontend/dist /app/web
RUN mkdir -p /app/out && chown -R app:app /app
USER app
EXPOSE 8080
ENV GIN_MODE=release
ENTRYPOINT ["/app/server"]
CMD ["-host", "0.0.0.0", "-port", "8080", "-outdir", "/app/out", "-web-dir", "/app/web"]
