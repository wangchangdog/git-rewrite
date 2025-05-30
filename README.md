# Git Rewrite Tools

Git履歴の書き換えとリモートリポジトリ管理を自動化するGoツールです。複数のGitリポジトリのauthor/emailを一括で変更し、GitHubリポジトリの作成・プッシュまでを自動化します。

## 🚀 機能

- **Git履歴の一括書き換え**: 複数リポジトリのauthor/emailを一度に変更
- **リモートリポジトリ自動作成**: GitHub APIを使用してリポジトリを自動作成
- **GitHub Actions制御**: プッシュ前にActionsを無効化、プッシュ後に有効化（デフォルト）
- **コラボレーター自動追加**: 環境変数またはJSONファイルでコラボレーターを自動設定
- **複数リポジトリ対応**: 指定ディレクトリ以下のすべてのGitリポジトリを自動検出・処理
- **GitHub API統合**: Personal Access Tokenを使用した安全な認証
- **包括的なテスト**: 単体テスト、統合テスト、エンドツーエンドテストを完備

## 📋 前提条件

- **Go**: 1.24.3以上
- **Git**: 2.0以上
- **GitHub Personal Access Token**: `repo`および`actions`スコープ付き
- **環境変数**: `GITHUB_USER`と`GITHUB_EMAIL`の設定

## 🔧 インストール

### 1. リポジトリのクローン

```bash
git clone <repository-url>
cd git-rewrite
```

### 2. 環境変数の設定

```bash
export GITHUB_USER="your-github-username"
export GITHUB_EMAIL="your-github-email@example.com"

# オプション: コラボレーター設定（環境変数）
export GITHUB_COLLABORATORS="user1:push,user2:admin,user3:pull"
```

### 3. ビルド

```bash
# 標準的なビルド
go build -o git-rewrite .

# Makefileを使用（推奨）
make build

# 依存関係の確認も含む
make deps build
```

## 📖 使用方法

### 新しいコマンド形式（推奨）

```bash
# ヘルプを表示
./git-rewrite --help

# 基本的な使用方法
./git-rewrite rewrite <github_token> --user <username> --email <email>

# 特定のディレクトリを指定
./git-rewrite rewrite <github_token> --user <username> --email <email> --target-dir ~/projects

# 組織リポジトリとして作成
./git-rewrite rewrite <github_token> --user <username> --email <email> --organization myorg

# 個人リポジトリ所有者を指定
./git-rewrite rewrite <github_token> --user <username> --email <email> --owner specificuser

# パブリックリポジトリとして作成
./git-rewrite rewrite <github_token> --user <username> --email <email> --public

# 全ブランチ・タグをプッシュ
./git-rewrite rewrite <github_token> --user <username> --email <email> --push-all

# GitHub Actions制御を無効化
./git-rewrite rewrite <github_token> --user <username> --email <email> --enable-actions

# デバッグモード
./git-rewrite rewrite <github_token> --user <username> --email <email> --debug
```

### コラボレーター設定

```bash
# コマンドラインでコラボレーター指定
./git-rewrite rewrite <github_token> --user <username> --email <email> --collaborators "dev1:push,admin1:admin"

# 設定ファイルを使用
./git-rewrite rewrite <github_token> --user <username> --email <email> --collaborator-config collaborators.json
```

### デモ機能

```bash
# デモ機能の実行
./git-rewrite demo <github_token> --user <username> --email <email>

# 内蔵テストの実行
./git-rewrite test
```

## 🎛️ GitHub Actions制御

### 概要

デフォルトでは、プッシュ時にGitHub Actionsが実行されることを防ぐため、以下の制御を自動的に行います：

1. **プッシュ前**: GitHub Actionsを無効化
2. **プッシュ実行**: 通常のGitプッシュ処理
3. **プッシュ後**: GitHub Actionsを有効化（GitHubのデフォルト状態）

### 設定オプション

```bash
# デフォルト（Actions制御有効）
./git-rewrite rewrite <token> --user <user> --email <email>

# Actions制御を無効化（Actionsが通常通り実行される）
./git-rewrite rewrite <token> --user <user> --email <email> --enable-actions
```

### 動作フロー

```
1. リポジトリのActions状態を取得・保存
   ↓
2. GitHub ActionsをOFFに設定
   ↓
3. Git履歴書き換え・プッシュ実行
   ↓
4. GitHub ActionsをON（デフォルト状態）に復元
```

### 必要な権限

GitHub Personal Access Tokenに以下のスコープが必要です：

- `repo`: リポジトリの作成・管理
- `actions`: GitHub Actionsの設定変更

## 🤝 コラボレーター機能

### 概要

リポジトリ作成時に自動的にコラボレーターを追加する機能です。環境変数またはJSONファイルで設定できます。

### 環境変数での設定

```bash
# 基本的な設定
export GITHUB_COLLABORATORS="user1:push,user2:admin,user3:pull"

# 複数の権限レベル
export GITHUB_COLLABORATORS="developer1:push,maintainer1:maintain,admin1:admin,viewer1:pull,triager1:triage"
```

### JSONファイルでの設定

`collaborators.json`ファイルを作成：

```json
{
  "default_collaborators": [
    {
      "username": "team-member1",
      "permission": "push"
    },
    {
      "username": "team-lead",
      "permission": "admin"
    }
  ],
  "project_collaborators": {
    "special-project": [
      {
        "username": "project-lead",
        "permission": "admin"
      },
      {
        "username": "developer1",
        "permission": "push"
      }
    ],
    "public-project": [
      {
        "username": "contributor1",
        "permission": "pull"
      },
      {
        "username": "maintainer1",
        "permission": "maintain"
      }
    ]
  }
}
```

### 権限レベル

| 権限 | 説明 |
|------|------|
| `pull` | 読み取り専用アクセス |
| `push` | 読み取り・書き込みアクセス |
| `admin` | 管理者権限（すべての操作が可能） |
| `maintain` | メンテナー権限（管理者権限の一部制限） |
| `triage` | トリアージ権限（Issue・PRの管理） |

### 優先順位

1. **コマンドライン引数** (`--collaborators`) - 最高優先度
2. **環境変数** (`GITHUB_COLLABORATORS`) - 高優先度
3. **設定ファイル** (`--collaborator-config`) - 中優先度
4. **プロジェクト固有設定** - 設定ファイル内の`project_collaborators`

## 📋 コマンドラインオプション

### 必須オプション

- `--user, -u <username>`: GitHubユーザー名
- `--email, -e <email>`: GitHubメールアドレス

### オプション引数

- `--target-dir, -d <directory>`: 対象ディレクトリ（デフォルト: `.`）
- `--owner, -o <owner>`: 個人リポジトリ所有者（最高優先度）
- `--organization <org>`: 組織名
- `--collaborators <list>`: コラボレーター設定（例: `user1:push,user2:admin`）
- `--collaborator-config, -c <file>`: コラボレーター設定ファイル
- `--push-all`: 全ブランチ・タグをプッシュ
- `--debug`: デバッグモード
- `--public`: パブリックリポジトリとして作成（デフォルト: プライベート）
- `--enable-actions`: GitHub Actions制御を無効化（デフォルトでActions制御は有効）

### 使用例

```bash
# 基本的な使用
./git-rewrite rewrite ghp_xxx --user myuser --email my@email.com

# 組織リポジトリとして作成
./git-rewrite rewrite ghp_xxx --user myuser --email my@email.com --organization myorg

# コラボレーター付きでパブリックリポジトリを作成
./git-rewrite rewrite ghp_xxx --user myuser --email my@email.com --public --collaborators "dev1:push,admin1:admin"

# Actions制御を無効化してデバッグモードで実行
./git-rewrite rewrite ghp_xxx --user myuser --email my@email.com --enable-actions --debug

# 全ブランチ・タグをプッシュ
./git-rewrite rewrite ghp_xxx --user myuser --email my@email.com --push-all
```

## 🧪 テスト

このプロジェクトは包括的なテストスイートを提供しています：

### テストの実行

```bash
# 🎯 すべてのテストを実行（推奨）
make test

# 📦 単体テストのみ
make test-unit

# 🔧 メイン関数のテスト
make test-main

# 🔗 統合テスト
make test-integration

# 📊 カバレッジ付きテスト
make test-coverage

# 🏗️ 内蔵テスト機能
make test-builtin
```

### テストの種類

#### 1. 単体テスト（Unit Tests）
```bash
go test ./pkg/...
```
- ユーティリティ関数のテスト
- Git操作の基本機能テスト
- URL解析・ファイル操作テスト
- GitHub Actions制御機能のテスト

#### 2. メイン関数テスト
```bash
go test .
```
- CLIインターフェースのテスト
- コマンドライン引数の処理テスト
- `os.Exit`のモック化テスト

#### 3. 統合テスト
```bash
go test ./tests/...
```
- 実際のバイナリ実行テスト
- エラーハンドリングテスト
- モックGitリポジトリを使用したテスト

### テスト結果の例

```
🧪 単体テストを実行しています...
✓ SafeDecode テスト成功
✓ ExtractRepoInfoFromURL テスト成功
✓ FileExists テスト成功
✓ SetDisableActionsOption テスト成功
✓ Git リポジトリ初期化成功
✓ ファイル作成とコミット成功
✓ リモート設定成功
✓ FindGitDirs テスト成功

=== すべてのテストが成功しました ===
```

## 🛠️ 開発

### 開発用コマンド

```bash
# コードフォーマット
make fmt

# Lint実行
make lint

# 依存関係の確認・更新
make deps

# ファイル変更の監視（entr必要）
make watch

# 使用可能なコマンド一覧
make usage
```

### デバッグ・開発支援

```bash
# ヘルプの確認
make help

# クリーンアップ
make clean

# 開発環境のセットアップ
make deps fmt lint test
```

## 📦 リリース

### マルチプラットフォームビルド

```bash
make build-release
```

生成されるバイナリ：
- `git-rewrite-darwin-amd64` (macOS Intel)
- `git-rewrite-darwin-arm64` (macOS Apple Silicon)
- `git-rewrite-linux-amd64` (Linux)
- `git-rewrite-windows-amd64.exe` (Windows)

## 📁 プロジェクト構造

```
git-rewrite/
├── 📄 main.go                 # メインエントリーポイント
├── 🧪 main_test.go           # メイン関数のテスト
├── 🔧 Makefile               # ビルド・テスト自動化
├── 📋 go.mod                 # Go モジュール定義
├── 📖 README.md              # このファイル
├── 📦 pkg/                   # 内部パッケージ
│   ├── 🎯 cli/              # CLIインターフェース
│   │   ├── commands/        # コマンド実装
│   │   └── config/          # 設定管理
│   ├── 🎯 demo/             # デモ機能
│   ├── 🐙 github/           # GitHub API クライアント
│   ├── 🔄 git/              # Git操作
│   ├── ✏️  rewriter/         # Git履歴書き換え機能
│   ├── 🧪 test/             # 内蔵テスト機能
│   └── 🔧 utils/            # ユーティリティ関数
└── 🔗 tests/                 # 統合テスト
    └── integration_test.go
```

## 🎯 使用例・ユースケース

### 1. 組織移行時のリポジトリ一括変更

```bash
# 会社のメールアドレスから個人のメールアドレスに一括変更
./git-rewrite rewrite <token> --user personal-account --email personal@example.com --target-dir ~/work-projects
```

### 2. 複数プロジェクトの統一

```bash
# 複数のプロジェクトのauthor情報を統一
./git-rewrite rewrite <token> --user unified-account --email unified@example.com --target-dir ~/all-projects
```

### 3. 新しいGitHubアカウントへの移行

```bash
# 既存のリポジトリを新しいGitHubアカウントに移行
./git-rewrite rewrite <new-account-token> --user new-account --email new@example.com --target-dir ~/repositories
```

### 4. CI/CDパイプラインでの使用

```bash
# Actions制御を無効化してCI/CDで使用
./git-rewrite rewrite <token> --user ci-user --email ci@example.com --enable-actions --public
```

## 🔒 セキュリティ

### GitHub Personal Access Token

1. **必要なスコープ**: 
   - `repo`: リポジトリの作成・管理
   - `actions`: GitHub Actionsの設定変更
2. **トークンの管理**: 環境変数や設定ファイルで安全に管理
3. **定期的な更新**: トークンの定期的な再生成を推奨

### 実行前の注意

⚠️ **重要**: Git履歴の書き換えは不可逆的な操作です。

```bash
# 実行前に必ずバックアップを作成
cp -r your-repo your-repo-backup

# または
git clone --mirror your-repo your-repo-backup.git
```

## 🧪 テストのベストプラクティス

このプロジェクトで実装されているテストパターン：

1. **テスト可能な設計**
   - `os.Exit`のモック化
   - 依存関係の注入
   - インターフェースの活用

2. **階層化されたテスト**
   - 単体テスト → 統合テスト → エンドツーエンドテスト
   - 各レベルでの適切なテスト範囲

3. **実際のバイナリテスト**
   - `exec.Command`を使用した実行テスト
   - 実際の使用シナリオの再現

4. **モックとスタブ**
   - 外部依存関係の分離
   - 予測可能なテスト環境

5. **テーブル駆動テスト**
   - 複数のテストケースの効率的な実行
   - 保守性の向上

## 🤝 貢献

プルリクエストやイシューの報告を歓迎します！

### 貢献の流れ

1. このリポジトリをフォーク
2. フィーチャーブランチを作成 (`git checkout -b feature/amazing-feature`)
3. 変更をコミット (`git commit -m 'Add amazing feature'`)
4. ブランチにプッシュ (`git push origin feature/amazing-feature`)
5. プルリクエストを作成

### 開発ガイドライン

- テストを追加してください
- コードフォーマットを実行してください (`make fmt`)
- Lintを通してください (`make lint`)
- すべてのテストが通ることを確認してください (`make test`)

## 📄 ライセンス

[ライセンス情報をここに記載]

## 🆘 トラブルシューティング

### よくある問題

#### 1. 環境変数が設定されていない

```bash
# エラー: --user フラグまたはGITHUB_USER環境変数が必要です
./git-rewrite rewrite <token> --user your-username --email your-email@example.com
```

#### 2. GitHub Personal Access Tokenの権限不足

```bash
# エラー: リポジトリ作成エラー: 403 Forbidden
# → トークンに 'repo' および 'actions' スコープが付与されているか確認
```

#### 3. GitHub Actions制御エラー

```bash
# エラー: GitHub Actions状態取得に失敗しました
# → トークンに 'actions' スコープが付与されているか確認
# → リポジトリが存在するか確認
```

#### 4. Gitリポジトリが見つからない

```bash
# エラー: 対象となる.gitディレクトリが見つかりませんでした
# → 指定したディレクトリにGitリポジトリが存在するか確認
ls -la your-directory/.git
```

#### 5. ビルドエラー

```bash
# Go のバージョンを確認
go version

# 依存関係を更新
make deps

# クリーンビルド
make clean build
```

### デバッグ方法

```bash
# 詳細なテスト出力
go test -v ./...

# デバッグモードで実行
./git-rewrite rewrite <token> --user <user> --email <email> --debug

# 内蔵テストでの動作確認
./git-rewrite test

# デモ機能での動作確認
./git-rewrite demo <token> --user <user> --email <email>
```

## 📞 サポート

問題が発生した場合は、以下の情報を含めてイシューを作成してください：

- OS とバージョン
- Go のバージョン
- エラーメッセージの全文
- 実行したコマンド
- 期待される動作と実際の動作
- GitHub Personal Access Tokenのスコープ

---

**Git Rewrite Tools** - Git履歴管理を簡単に 🚀 