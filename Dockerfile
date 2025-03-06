# ビルドステージ
FROM golang:1.22-alpine AS builder

# 必要なパッケージをインストール
RUN apk add --no-cache git

# 作業ディレクトリの設定
WORKDIR /app

# 依存関係のコピーとダウンロード
COPY go.mod go.sum ./
RUN go mod download

# ソースコードのコピー
COPY . .

# アプリケーションのビルド
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gox-api ./cmd/api

# 実行ステージ
FROM alpine:latest

# 必要なパッケージのインストール
RUN apk --no-cache add ca-certificates tzdata

# タイムゾーンの設定
ENV TZ=Asia/Tokyo

# 作業ディレクトリの設定
WORKDIR /root/

# ビルドステージからのバイナリをコピー
COPY --from=builder /app/gox-api .
COPY --from=builder /app/.env .

# アプリケーションの実行
CMD ["./gox-api"]

# ヘルスチェック設定
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://localhost:8080/health || exit 1

# ポート設定
EXPOSE 8080 