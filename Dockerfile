# syntax=docker/dockerfile:1

# ---- 构建阶段 ----
FROM golang:1.27-alpine AS builder

WORKDIR /src

# 先拷依赖清单，利用分层缓存
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# TARGETOS/TARGETARCH 由 buildx 自动注入；非 buildx 场景回退到构建机默认值
# 例：交叉编译 amd64 -> docker build --build-arg TARGETARCH=amd64 .
ARG TARGETOS=linux
ARG TARGETARCH
RUN CGO_ENABLED=0 \
    GOOS=${TARGETOS:-linux} \
    GOARCH=${TARGETARCH:-$(go env GOARCH)} \
    go build -trimpath -ldflags="-s -w" -o /out/host-agent .

# ---- 产物阶段：只含二进制，方便导出 ----
FROM scratch AS release
COPY --from=builder /out/host-agent /host-agent
