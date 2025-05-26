# Git Rewrite Tools Makefile

# 変数定義
BINARY_NAME=git-rewrite-tools
MAIN_PACKAGE=.
BUILD_DIR=.
TEST_TIMEOUT=30s

# デフォルトターゲット
.PHONY: all
all: clean build test

# ビルド
.PHONY: build
build:
	@echo "🔨 バイナリをビルドしています..."
	go build -o $(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "✅ ビルド完了: $(BINARY_NAME)"

# テスト実行
.PHONY: test
test: build
	@echo "🧪 単体テストを実行しています..."
	go test -v -timeout $(TEST_TIMEOUT) ./pkg/...
	@echo "🧪 メインパッケージのテストを実行しています..."
	go test -v -timeout $(TEST_TIMEOUT) .
	@echo "🧪 統合テストを実行しています..."
	go test -v -timeout $(TEST_TIMEOUT) ./tests/...
	@echo "✅ すべてのテストが完了しました"

# 単体テストのみ
.PHONY: test-unit
test-unit:
	@echo "🧪 単体テストを実行しています..."
	go test -v -timeout $(TEST_TIMEOUT) ./pkg/...

# 統合テストのみ
.PHONY: test-integration
test-integration: build
	@echo "🧪 統合テストを実行しています..."
	go test -v -timeout $(TEST_TIMEOUT) ./tests/...

# メインパッケージのテストのみ
.PHONY: test-main
test-main: build
	@echo "🧪 メインパッケージのテストを実行しています..."
	go test -v -timeout $(TEST_TIMEOUT) .

# カバレッジ付きテスト
.PHONY: test-coverage
test-coverage: build
	@echo "🧪 カバレッジ付きテストを実行しています..."
	go test -v -timeout $(TEST_TIMEOUT) -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "📊 カバレッジレポートが coverage.html に生成されました"

# 内蔵テストコマンドの実行
.PHONY: test-builtin
test-builtin: build
	@echo "🧪 内蔵テストコマンドを実行しています..."
	./$(BINARY_NAME) test

# デモの実行（GitHub tokenが必要）
.PHONY: demo
demo: build
	@echo "🎯 デモを実行しています..."
	@if [ -z "$(GITHUB_TOKEN)" ]; then \
		echo "❌ エラー: GITHUB_TOKEN環境変数が設定されていません"; \
		echo "使用方法: make demo GITHUB_TOKEN=your_token"; \
		exit 1; \
	fi
	./$(BINARY_NAME) demo $(GITHUB_TOKEN)

# ヘルプの表示
.PHONY: help
help: build
	@echo "📖 ヘルプを表示しています..."
	./$(BINARY_NAME)

# クリーンアップ
.PHONY: clean
clean:
	@echo "🧹 クリーンアップしています..."
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	@echo "✅ クリーンアップ完了"

# 依存関係の確認
.PHONY: deps
deps:
	@echo "📦 依存関係を確認しています..."
	go mod tidy
	go mod verify
	@echo "✅ 依存関係の確認完了"

# フォーマット
.PHONY: fmt
fmt:
	@echo "🎨 コードをフォーマットしています..."
	go fmt ./...
	@echo "✅ フォーマット完了"

# Lint
.PHONY: lint
lint:
	@echo "🔍 Lintを実行しています..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "⚠️  golangci-lint がインストールされていません。go vetを実行します..."; \
		go vet ./...; \
	fi
	@echo "✅ Lint完了"

# 開発用のワッチモード（要: entr）
.PHONY: watch
watch:
	@echo "👀 ファイル変更を監視しています..."
	@if command -v entr >/dev/null 2>&1; then \
		find . -name "*.go" | entr -c make test; \
	else \
		echo "❌ エラー: entr がインストールされていません"; \
		echo "インストール: brew install entr (macOS) または apt-get install entr (Ubuntu)"; \
	fi

# リリース用ビルド（複数プラットフォーム）
.PHONY: build-release
build-release: clean
	@echo "🚀 リリース用ビルドを実行しています..."
	GOOS=darwin GOARCH=amd64 go build -o $(BINARY_NAME)-darwin-amd64 $(MAIN_PACKAGE)
	GOOS=darwin GOARCH=arm64 go build -o $(BINARY_NAME)-darwin-arm64 $(MAIN_PACKAGE)
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME)-linux-amd64 $(MAIN_PACKAGE)
	GOOS=windows GOARCH=amd64 go build -o $(BINARY_NAME)-windows-amd64.exe $(MAIN_PACKAGE)
	@echo "✅ リリース用ビルド完了"

# 使用方法の表示
.PHONY: usage
usage:
	@echo "Git Rewrite Tools - 使用可能なMakeターゲット:"
	@echo ""
	@echo "  build              - バイナリをビルド"
	@echo "  test               - 全テストを実行"
	@echo "  test-unit          - 単体テストのみ実行"
	@echo "  test-integration   - 統合テストのみ実行"
	@echo "  test-main          - メインパッケージのテストのみ実行"
	@echo "  test-coverage      - カバレッジ付きテスト実行"
	@echo "  test-builtin       - 内蔵テストコマンド実行"
	@echo "  demo               - デモ実行 (GITHUB_TOKEN必要)"
	@echo "  help               - ヘルプ表示"
	@echo "  clean              - クリーンアップ"
	@echo "  deps               - 依存関係確認"
	@echo "  fmt                - コードフォーマット"
	@echo "  lint               - Lint実行"
	@echo "  watch              - ファイル変更監視 (entr必要)"
	@echo "  build-release      - リリース用ビルド"
	@echo "  usage              - この使用方法を表示"
	@echo ""
	@echo "例:"
	@echo "  make build"
	@echo "  make test"
	@echo "  make demo GITHUB_TOKEN=your_github_token" 