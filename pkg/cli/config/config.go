package config

import (
	"flag"
	"fmt"
	"os"
)

// Config はアプリケーション全体の設定を保持する
type Config struct {
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

// ParseRewriteArgs はrewriteコマンドの引数を解析する
func ParseRewriteArgs(args []string) (*Config, error) {
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

	config := &Config{
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
