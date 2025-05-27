package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"git-rewrite-and-go/pkg/demo"
	"git-rewrite-and-go/pkg/rewriter"
	"git-rewrite-and-go/pkg/test"
	"git-rewrite-and-go/pkg/utils"
)

// osExit はテスト時にos.Exitをモック可能にするための変数
var osExit = os.Exit

func main() {
	if len(os.Args) < 2 {
		showHelp()
		osExit(1)
	}

	command := os.Args[1]

	// --help オプションのチェック
	if command == "--help" || command == "-h" || command == "help" {
		showHelp()
		osExit(0)
	}

	switch command {
	case "rewrite":
		runRewrite(os.Args[2:])
	case "demo":
		runDemo(os.Args[2:])
	case "test":
		runTests()
	default:
		fmt.Printf("不明なコマンド: %s\n", command)
		fmt.Println("")
		showHelp()
		osExit(1)
	}
}

func showHelp() {
	fmt.Println("使用方法:")
	fmt.Println("  git-rewrite <command> [options]")
	fmt.Println("  git-rewrite --help")
	fmt.Println("")
	fmt.Println("利用可能なコマンド:")
	fmt.Println("  rewrite <github_token> [target_directory] [collaborator_config] [--push-all] - Git履歴の書き換えとリモートリポジトリ管理")
	fmt.Println("  demo <github_token>                                                          - リモートリポジトリ作成機能のデモ")
	fmt.Println("  test                                                                         - テストの実行")
	fmt.Println("  help, --help, -h                                                             - このヘルプを表示")
	fmt.Println("")
	fmt.Println("環境変数:")
	fmt.Println("  GITHUB_USER           - GitHubユーザー名")
	fmt.Println("  GITHUB_EMAIL          - GitHubメールアドレス")
	fmt.Println("  GITHUB_REPOSITORY_OWNER - 個人リポジトリの所有者（優先度: 最高）")
	fmt.Println("  GITHUB_ORGANIZATION   - GitHubの組織名（組織または個人ユーザー名を指定可能）")
	fmt.Println("  GITHUB_COLLABORATORS  - コラボレーター設定（例: user1:push,user2:admin）")
	fmt.Println("  GIT_REWRITE_DEBUG     - デバッグモード（設定時は詳細なログを表示）")
	fmt.Println("")
	fmt.Println("コラボレーター設定:")
	fmt.Println("  環境変数またはJSONファイルでコラボレーターを設定可能")
	fmt.Println("  権限: pull, push, admin, maintain, triage")
	fmt.Println("")
	fmt.Println("例:")
	fmt.Println("  git-rewrite --help")
	fmt.Println("  git-rewrite test")
	fmt.Println("  git-rewrite rewrite ghp_xxxxxxxxxxxxxxxxxxxx")
	fmt.Println("  git-rewrite rewrite ghp_xxxxxxxxxxxxxxxxxxxx ~/projects")
	fmt.Println("  git-rewrite rewrite ghp_xxxxxxxxxxxxxxxxxxxx ~/projects collaborators.json")
	fmt.Println("  git-rewrite rewrite ghp_xxxxxxxxxxxxxxxxxxxx ~/projects collaborators.json --push-all")
	fmt.Println("  git-rewrite rewrite ghp_xxxxxxxxxxxxxxxxxxxx --push-all")
	fmt.Println("  git-rewrite demo ghp_xxxxxxxxxxxxxxxxxxxx")
	fmt.Println("")
	fmt.Println("コラボレーター設定例:")
	fmt.Println("  # 環境変数で設定")
	fmt.Println("  export GITHUB_COLLABORATORS=\"dev1:push,admin1:admin,viewer1:pull\"")
	fmt.Println("  git-rewrite rewrite ghp_xxxxxxxxxxxxxxxxxxxx")
	fmt.Println("")
	fmt.Println("  # 個人リポジトリの所有者を指定")
	fmt.Println("  export GITHUB_REPOSITORY_OWNER=\"specific-user\"")
	fmt.Println("  git-rewrite rewrite ghp_xxxxxxxxxxxxxxxxxxxx")
	fmt.Println("")
	fmt.Println("  # 組織リポジトリとして設定")
	fmt.Println("  export GITHUB_ORGANIZATION=\"my-organization\"")
	fmt.Println("  git-rewrite rewrite ghp_xxxxxxxxxxxxxxxxxxxx")
	fmt.Println("")
	fmt.Println("  # 他の個人ユーザーのリポジトリとして設定（注意：権限が必要）")
	fmt.Println("  export GITHUB_ORGANIZATION=\"other-user\"")
	fmt.Println("  git-rewrite rewrite ghp_xxxxxxxxxxxxxxxxxxxx")
	fmt.Println("")
	fmt.Println("  # 設定ファイルで設定")
	fmt.Println("  echo '{\"default_collaborators\":[{\"username\":\"dev1\",\"permission\":\"push\"}]}' > collaborators.json")
	fmt.Println("  git-rewrite rewrite ghp_xxxxxxxxxxxxxxxxxxxx ~/projects collaborators.json")
}

func runRewrite(args []string) {
	if len(args) < 1 {
		fmt.Println("使用方法: git-rewrite rewrite <github_token> [target_directory] [collaborator_config] [--push-all]")
		fmt.Println("  github_token: GitHubのPersonal Access Token（repositoryアクセス権限付き）")
		fmt.Println("  target_directory: 対象ディレクトリ（省略時は現在のディレクトリ）")
		fmt.Println("  collaborator_config: コラボレーター設定（例: user1:push,user2:admin）")
		fmt.Println("  --push-all: プッシュ成功後にローカルの全ブランチ・タグをプッシュ")
		fmt.Println("")
		fmt.Println("環境変数も必要です:")
		fmt.Println("  GITHUB_USER: GitHubユーザー名")
		fmt.Println("  GITHUB_EMAIL: GitHubメールアドレス")
		osExit(1)
	}

	githubToken := args[0]
	targetDir := "."
	collaboratorConfig := ""
	pushAll := false

	// 引数を解析
	for i := 1; i < len(args); i++ {
		arg := args[i]
		if arg == "--push-all" {
			pushAll = true
		} else if collaboratorConfig == "" && !strings.HasPrefix(arg, "--") {
			if targetDir == "." {
				targetDir = arg
			} else {
				collaboratorConfig = arg
			}
		} else if !strings.HasPrefix(arg, "--") {
			if targetDir == "." {
				targetDir = arg
			} else if collaboratorConfig == "" {
				collaboratorConfig = arg
			}
		}
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
	var gitRewriter *rewriter.Rewriter
	if collaboratorConfig != "" {
		gitRewriter = rewriter.NewRewriterWithConfig(githubToken, githubUser, githubEmail, collaboratorConfig)
		fmt.Printf("コラボレーター設定ファイル: %s\n", collaboratorConfig)
	} else {
		gitRewriter = rewriter.NewRewriter(githubToken, githubUser, githubEmail)
		fmt.Println("コラボレーター設定: 環境変数のみ使用")
	}

	// プッシュオプションを設定
	gitRewriter.SetPushAllOption(pushAll)
	if pushAll {
		fmt.Println("🚀 --push-all オプションが有効です。プッシュ成功後に全ブランチ・タグをプッシュします。")
	}

	// 結果を追跡
	var successCount int
	var failedRepos []string
	var pushFailedRepos []string

	// 各リポジトリを処理
	for i, gitDir := range gitDirs {
		fmt.Printf("\n=== [%d/%d] %s でスクリプトを実行します ===\n", i+1, len(gitDirs), gitDir)

		result := gitRewriter.ProcessRepository(gitDir)

		if result.Success {
			successCount++
			fmt.Printf("✅ %s の全処理が完了しました。\n", gitDir)
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
		fmt.Println("使用方法: git-rewrite demo <github_token>")
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
