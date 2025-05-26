package main

import (
	"fmt"
	"os"
	"path/filepath"

	"git-rewrite-tools/pkg/demo"
	"git-rewrite-tools/pkg/rewriter"
	"git-rewrite-tools/pkg/test"
	"git-rewrite-tools/pkg/utils"
)

// osExit はテスト時にos.Exitをモック可能にするための変数
var osExit = os.Exit

func main() {
	if len(os.Args) < 2 {
		fmt.Println("使用方法:")
		fmt.Println("  git-rewrite-tools <command> [options]")
		fmt.Println("")
		fmt.Println("利用可能なコマンド:")
		fmt.Println("  rewrite <github_token> [target_directory] - Git履歴の書き換えとリモートリポジトリ管理")
		fmt.Println("  demo <github_token>                      - リモートリポジトリ作成機能のデモ")
		fmt.Println("  test                                     - テストの実行")
		fmt.Println("")
		fmt.Println("環境変数:")
		fmt.Println("  GITHUB_USER  - GitHubユーザー名")
		fmt.Println("  GITHUB_EMAIL - GitHubメールアドレス")
		osExit(1)
	}

	command := os.Args[1]

	switch command {
	case "rewrite":
		runRewrite(os.Args[2:])
	case "demo":
		runDemo(os.Args[2:])
	case "test":
		runTests()
	default:
		fmt.Printf("不明なコマンド: %s\n", command)
		osExit(1)
	}
}

func runRewrite(args []string) {
	if len(args) < 1 {
		fmt.Println("使用方法: git-rewrite-tools rewrite <github_token> [target_directory]")
		fmt.Println("  github_token: GitHubのPersonal Access Token（repositoryアクセス権限付き）")
		fmt.Println("  target_directory: 対象ディレクトリ（省略時は現在のディレクトリ）")
		fmt.Println("")
		fmt.Println("環境変数も必要です:")
		fmt.Println("  GITHUB_USER: GitHubユーザー名")
		fmt.Println("  GITHUB_EMAIL: GitHubメールアドレス")
		osExit(1)
	}

	githubToken := args[0]
	targetDir := "."
	if len(args) > 1 {
		targetDir = args[1]
	}

	// 環境変数をチェック
	githubUser, githubEmail, err := utils.CheckEnvironmentVariables()
	if err != nil {
		fmt.Printf("エラー: %v\n", err)
		osExit(1)
	}

	// 対象ディレクトリの絶対パスを取得
	absTargetDir, err := filepath.Abs(targetDir)
	if err != nil {
		fmt.Printf("エラー: ディレクトリパス解決に失敗しました: %v\n", err)
		osExit(1)
	}

	// Gitリポジトリを検索
	gitDirs, err := utils.FindGitDirs(absTargetDir)
	if err != nil {
		fmt.Printf("エラー: Gitリポジトリの検索に失敗しました: %v\n", err)
		osExit(1)
	}

	if len(gitDirs) == 0 {
		fmt.Println("対象となる.gitディレクトリが見つかりませんでした。")
		osExit(0)
	}

	fmt.Printf("見つかったGitリポジトリ: %d個\n", len(gitDirs))
	for _, gitDir := range gitDirs {
		fmt.Printf("  - %s\n", gitDir)
	}
	fmt.Println()

	// Rewriterを作成
	rewriter := rewriter.NewRewriter(githubToken, githubUser, githubEmail)

	// 結果を追跡
	var successCount int
	var failedRepos []string
	var pushFailedRepos []string

	// 各リポジトリを処理
	for i, gitDir := range gitDirs {
		fmt.Printf("\n=== [%d/%d] %s でスクリプトを実行します ===\n", i+1, len(gitDirs), gitDir)

		result := rewriter.ProcessRepository(gitDir)

		if result.Success {
			successCount++
			fmt.Printf("✓ %s の全処理が完了しました。\n", gitDir)
		} else if result.HistoryRewritten {
			pushFailedRepos = append(pushFailedRepos, gitDir)
			fmt.Printf("⚠️  %s の履歴書き換えは成功しましたが、プッシュに失敗しました。\n", gitDir)
			fmt.Printf("エラー: %v\n", result.Error)
		} else {
			failedRepos = append(failedRepos, gitDir)
			fmt.Printf("✗ %s で処理に失敗しました。\n", gitDir)
			fmt.Printf("エラー: %v\n", result.Error)
		}
	}

	// 最終結果の表示
	fmt.Printf("\n=== 実行結果 ===\n")
	fmt.Printf("完全成功: %d/%d リポジトリ\n", successCount, len(gitDirs))

	if len(pushFailedRepos) > 0 {
		fmt.Printf("履歴書き換え成功・プッシュ失敗: %d リポジトリ\n", len(pushFailedRepos))
		fmt.Println("プッシュに失敗したリポジトリ:")
		for _, repo := range pushFailedRepos {
			fmt.Printf("  - %s\n", repo)
		}
	}

	if len(failedRepos) > 0 {
		fmt.Printf("履歴書き換え失敗: %d リポジトリ\n", len(failedRepos))
		fmt.Println("履歴書き換えに失敗したリポジトリ:")
		for _, repo := range failedRepos {
			fmt.Printf("  - %s\n", repo)
		}
	}

	if len(failedRepos) > 0 || len(pushFailedRepos) > 0 {
		osExit(1)
	} else {
		fmt.Println("すべてのリポジトリで履歴書き換えとプッシュが正常に完了しました。")
	}
}

func runDemo(args []string) {
	if len(args) < 1 {
		fmt.Println("使用方法: git-rewrite-tools demo <github_token>")
		fmt.Println("  github_token: GitHubのPersonal Access Token（repositoryアクセス権限付き）")
		fmt.Println("")
		fmt.Println("環境変数も必要です:")
		fmt.Println("  GITHUB_USER: GitHubユーザー名")
		fmt.Println("  GITHUB_EMAIL: GitHubメールアドレス")
		osExit(1)
	}

	githubToken := args[0]

	// デモを実行
	if err := demo.RunDemo(githubToken); err != nil {
		fmt.Printf("デモ実行エラー: %v\n", err)
		osExit(1)
	}

	// 環境変数を取得して空のリポジトリデモも実行
	githubUser, githubEmail, err := utils.CheckEnvironmentVariables()
	if err != nil {
		fmt.Printf("環境変数エラー: %v\n", err)
		osExit(1)
	}

	if err := demo.RunEmptyRepoDemo(githubUser, githubEmail); err != nil {
		fmt.Printf("空のリポジトリデモエラー: %v\n", err)
		osExit(1)
	}
}

func runTests() {
	fmt.Println("テスト機能を実行中...")

	if err := test.RunTests(); err != nil {
		fmt.Printf("テスト実行エラー: %v\n", err)
		osExit(1)
	}
}
