package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"git-rewrite-and-go/pkg/demo"
	"git-rewrite-and-go/pkg/rewriter"
	"git-rewrite-and-go/pkg/test"
	"git-rewrite-and-go/pkg/utils"
)

// RewriteConfig はrewriteコマンドの設定を保持する
type RewriteConfig struct {
	GitHubToken        string
	GitHubUser         string
	GitHubEmail        string
	TargetDir          string
	Owner              string
	Organization       string
	Collaborators      string
	CollaboratorConfig string
	PushAll            bool
	Debug              bool
	Private            bool
}

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
	fmt.Println("  rewrite <github_token> --user <user> --email <email> [options] - Git履歴の書き換えとリモートリポジトリ管理")
	fmt.Println("  demo <github_token> --user <user> --email <email>              - リモートリポジトリ作成機能のデモ")
	fmt.Println("  test                                                           - テストの実行")
	fmt.Println("  help, --help, -h                                               - このヘルプを表示")
	fmt.Println("")
	fmt.Println("rewriteコマンドのオプション:")
	fmt.Println("  --user, -u <username>           GitHubユーザー名（必須）")
	fmt.Println("  --email, -e <email>             GitHubメールアドレス（必須）")
	fmt.Println("  --target-dir, -d <directory>    対象ディレクトリ（デフォルト: .）")
	fmt.Println("  --owner, -o <owner>             個人リポジトリ所有者（最高優先度）")
	fmt.Println("  --organization <org>            組織名")
	fmt.Println("  --collaborators <list>          コラボレーター設定（例: user1:push,user2:admin）")
	fmt.Println("  --collaborator-config, -c <file> コラボレーター設定ファイル")
	fmt.Println("  --push-all                      全ブランチ・タグをプッシュ")
	fmt.Println("  --debug                         デバッグモード")
	fmt.Println("  --public                        パブリックリポジトリとして作成（デフォルト: プライベート）")
	fmt.Println("")
	fmt.Println("例:")
	fmt.Println("  git-rewrite --help")
	fmt.Println("  git-rewrite test")
	fmt.Println("  git-rewrite rewrite ghp_xxx --user myuser --email my@email.com")
	fmt.Println("  git-rewrite rewrite ghp_xxx --user myuser --email my@email.com --target-dir ~/projects")
	fmt.Println("  git-rewrite rewrite ghp_xxx --user myuser --email my@email.com --organization myorg")
	fmt.Println("  git-rewrite rewrite ghp_xxx --user myuser --email my@email.com --owner specificuser")
	fmt.Println("  git-rewrite rewrite ghp_xxx --user myuser --email my@email.com --collaborators \"dev1:push,admin1:admin\"")
	fmt.Println("  git-rewrite rewrite ghp_xxx --user myuser --email my@email.com --collaborator-config collaborators.json --push-all")
	fmt.Println("  git-rewrite rewrite ghp_xxx --user myuser --email my@email.com --public --debug")
	fmt.Println("  git-rewrite demo ghp_xxx --user myuser --email my@email.com")
	fmt.Println("")
	fmt.Println("後方互換性:")
	fmt.Println("  環境変数も引き続きサポートされますが、コマンド引数が優先されます。")
	fmt.Println("  GITHUB_USER, GITHUB_EMAIL, GITHUB_ORGANIZATION, GITHUB_REPOSITORY_OWNER,")
	fmt.Println("  GITHUB_COLLABORATORS, GIT_REWRITE_DEBUG")
}

// parseRewriteArgs はrewriteコマンドの引数を解析する
func parseRewriteArgs(args []string) (*RewriteConfig, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("GitHubトークンが必要です")
	}

	fs := flag.NewFlagSet("rewrite", flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Println("使用方法: git-rewrite rewrite <github_token> --user <user> --email <email> [options]")
		fmt.Println("")
		fmt.Println("必須引数:")
		fmt.Println("  --user, -u <username>           GitHubユーザー名")
		fmt.Println("  --email, -e <email>             GitHubメールアドレス")
		fmt.Println("")
		fmt.Println("オプション引数:")
		fmt.Println("  --target-dir, -d <directory>    対象ディレクトリ（デフォルト: .）")
		fmt.Println("  --owner, -o <owner>             個人リポジトリ所有者（最高優先度）")
		fmt.Println("  --organization <org>            組織名")
		fmt.Println("  --collaborators <list>          コラボレーター設定（例: user1:push,user2:admin）")
		fmt.Println("  --collaborator-config, -c <file> コラボレーター設定ファイル")
		fmt.Println("  --push-all                      全ブランチ・タグをプッシュ")
		fmt.Println("  --debug                         デバッグモード")
		fmt.Println("  --public                        パブリックリポジトリとして作成（デフォルト: プライベート）")
	}

	config := &RewriteConfig{
		GitHubToken: args[0],
		TargetDir:   ".",
		Private:     true, // デフォルトはプライベート
	}

	// フラグ定義
	fs.StringVar(&config.GitHubUser, "user", "", "GitHubユーザー名（必須）")
	fs.StringVar(&config.GitHubUser, "u", "", "GitHubユーザー名（必須）")
	fs.StringVar(&config.GitHubEmail, "email", "", "GitHubメールアドレス（必須）")
	fs.StringVar(&config.GitHubEmail, "e", "", "GitHubメールアドレス（必須）")
	fs.StringVar(&config.TargetDir, "target-dir", ".", "対象ディレクトリ")
	fs.StringVar(&config.TargetDir, "d", ".", "対象ディレクトリ")
	fs.StringVar(&config.Owner, "owner", "", "個人リポジトリ所有者")
	fs.StringVar(&config.Owner, "o", "", "個人リポジトリ所有者")
	fs.StringVar(&config.Organization, "organization", "", "組織名")
	fs.StringVar(&config.Collaborators, "collaborators", "", "コラボレーター設定")
	fs.StringVar(&config.CollaboratorConfig, "collaborator-config", "", "コラボレーター設定ファイル")
	fs.StringVar(&config.CollaboratorConfig, "c", "", "コラボレーター設定ファイル")
	fs.BoolVar(&config.PushAll, "push-all", false, "全ブランチ・タグをプッシュ")
	fs.BoolVar(&config.Debug, "debug", false, "デバッグモード")

	// --publicフラグが指定された場合はPrivateをfalseにする
	var public bool
	fs.BoolVar(&public, "public", false, "パブリックリポジトリとして作成")

	// 引数を解析
	if err := fs.Parse(args[1:]); err != nil {
		return nil, err
	}

	// --publicが指定された場合はプライベートをfalseにする
	if public {
		config.Private = false
	}

	// 環境変数からのフォールバック（後方互換性）
	config.GitHubUser = getConfigValue(config.GitHubUser, "GITHUB_USER", "")
	config.GitHubEmail = getConfigValue(config.GitHubEmail, "GITHUB_EMAIL", "")
	config.Owner = getConfigValue(config.Owner, "GITHUB_REPOSITORY_OWNER", "")
	config.Organization = getConfigValue(config.Organization, "GITHUB_ORGANIZATION", "")
	config.Collaborators = getConfigValue(config.Collaborators, "GITHUB_COLLABORATORS", "")

	// デバッグモードの環境変数チェック
	if !config.Debug && os.Getenv("GIT_REWRITE_DEBUG") != "" {
		config.Debug = true
	}

	// 必須フラグの検証
	if config.GitHubUser == "" {
		return nil, fmt.Errorf("--user フラグまたはGITHUB_USER環境変数が必要です")
	}
	if config.GitHubEmail == "" {
		return nil, fmt.Errorf("--email フラグまたはGITHUB_EMAIL環境変数が必要です")
	}

	return config, nil
}

// getConfigValue はフラグ値、環境変数、デフォルト値の優先順位で値を取得する
func getConfigValue(flagValue, envKey, defaultValue string) string {
	if flagValue != "" {
		return flagValue
	}
	if envValue := os.Getenv(envKey); envValue != "" {
		return envValue
	}
	return defaultValue
}

func runRewrite(args []string) {
	config, err := parseRewriteArgs(args)
	if err != nil {
		fmt.Printf("引数解析エラー: %v\n", err)
		fmt.Println("")
		fmt.Println("使用方法: git-rewrite rewrite <github_token> --user <user> --email <email> [options]")
		fmt.Println("詳細なヘルプ: git-rewrite rewrite --help")
		osExit(1)
	}

	// デバッグモードの設定
	if config.Debug {
		os.Setenv("GIT_REWRITE_DEBUG", "1")
		fmt.Println("🐛 デバッグモードが有効です")
	}

	// 設定の表示
	fmt.Printf("📋 設定情報:\n")
	fmt.Printf("  GitHubユーザー: %s\n", config.GitHubUser)
	fmt.Printf("  GitHubメール: %s\n", config.GitHubEmail)
	fmt.Printf("  対象ディレクトリ: %s\n", config.TargetDir)
	if config.Owner != "" {
		fmt.Printf("  個人リポジトリ所有者: %s\n", config.Owner)
	}
	if config.Organization != "" {
		fmt.Printf("  組織: %s\n", config.Organization)
	}
	fmt.Printf("  リポジトリタイプ: %s\n", map[bool]string{true: "プライベート", false: "パブリック"}[config.Private])
	if config.PushAll {
		fmt.Printf("  全ブランチ・タグプッシュ: 有効\n")
	}
	fmt.Println()

	// 対象ディレクトリの絶対パスを取得
	absTargetDir, err := filepath.Abs(config.TargetDir)
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
	if config.CollaboratorConfig != "" {
		gitRewriter = rewriter.NewRewriterWithConfig(config.GitHubToken, config.GitHubUser, config.GitHubEmail, config.CollaboratorConfig)
		fmt.Printf("コラボレーター設定ファイル: %s\n", config.CollaboratorConfig)
	} else {
		gitRewriter = rewriter.NewRewriter(config.GitHubToken, config.GitHubUser, config.GitHubEmail)
		if config.Collaborators != "" {
			fmt.Printf("コラボレーター設定: %s\n", config.Collaborators)
		} else {
			fmt.Println("コラボレーター設定: なし")
		}
	}

	// 設定をRewriterに適用
	gitRewriter.SetPushAllOption(config.PushAll)
	gitRewriter.SetOwnershipConfig(config.Owner, config.Organization)
	gitRewriter.SetPrivateOption(config.Private)
	gitRewriter.SetCollaboratorsFromString(config.Collaborators)

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
	// demoコマンドも新しい引数形式をサポート
	config, err := parseRewriteArgs(args)
	if err != nil {
		fmt.Printf("引数解析エラー: %v\n", err)
		fmt.Println("")
		fmt.Println("使用方法: git-rewrite demo <github_token> --user <user> --email <email>")
		osExit(1)
	}

	// デバッグモードの設定
	if config.Debug {
		os.Setenv("GIT_REWRITE_DEBUG", "1")
	}

	// デモを実行
	if err := demo.RunDemo(config.GitHubToken); err != nil {
		fmt.Printf("デモ実行エラー: %v\n", err)
		osExit(1)
	}

	if err := demo.RunEmptyRepoDemo(config.GitHubToken, config.GitHubUser, config.GitHubEmail); err != nil {
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
