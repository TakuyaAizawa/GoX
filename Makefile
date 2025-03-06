.PHONY: run build test lint migrate swagger docker-up docker-down

# アプリケーション実行
run:
	go run cmd/api/main.go

# ビルド
build:
	go build -o bin/api cmd/api/main.go

# 依存パッケージインストール
deps:
	go mod download

# テスト実行
test:
	go test -v ./...

# ベンチマークテスト
benchmark:
	go test -bench=. ./...

# テストカバレッジ
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Lintツール実行
lint:
	golangci-lint run ./...

# マイグレーション実行
migrate-up:
	migrate -path migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)" up

# マイグレーションロールバック
migrate-down:
	migrate -path migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)" down 1

# マイグレーションファイル作成
migrate-create:
	@read -p "Migration name: " name; \
	migrate create -ext sql -dir migrations -seq $$name

# データベースセットアップ
db-setup:
	go run cmd/dbsetup/main.go

# データベースマイグレーションロールバック
db-rollback:
	go run cmd/dbsetup/main.go --rollback

# Swaggerドキュメント生成
swagger:
	swag init -g cmd/api/main.go -o docs/swagger

# Dockerコンテナ起動
docker-up:
	docker-compose up -d

# Dockerコンテナ停止
docker-down:
	docker-compose down

# アプリケーションの初期設定
setup: deps docker-up migrate-up swagger
	@echo "Setup completed successfully!" 