package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"git-rewrite-and-go/pkg/cli/config"
	"git-rewrite-and-go/pkg/rewriter"
	"git-rewrite-and-go/pkg/utils"
)

// RewriteCommand はrewriteコマンドを実行する
type RewriteCommand struct{}

// NewRewriteCommand は新しいRewriteCommandを作成する
func NewRewriteCommand() *RewriteCommand {
	return &RewriteCommand{}
}

// Execute はrewriteコマンドを実行する
func (c *RewriteCommand) Execute(args []string) error {
	config, err := config.ParseRewriteArgs(args)
	if err != nil {
		fmt.Printf("引数解析エラー: %v\n", err)
		fmt.Println("")
		fmt.Println("使用方法: git-rewrite rewrite <github_token> --user <user> --email <email> [options]")
		fmt.Println("詳細なヘルプ: git-rewrite rewrite --help")
		return err
	}

	// デバッグモードの設定
	if config.Debug {
		os.Setenv("GIT_REWRITE_DEBUG", "1")
		fmt.Println("🐛 デバッグモードが有効です")
	}

	// 設定の表示
	c.displayConfig(config)

	// 対象ディレクトリの絶対パスを取得
	absTargetDir, err := filepath.Abs(config.TargetDir)
	if err != nil {
		return fmt.Errorf("ディレクトリパス解決に失敗しました: %v", err)
	}

	// Gitリポジトリを検索
	gitDirs, err := utils.FindGitDirs(absTargetDir)
	if err != nil {
		return fmt.Errorf("Gitリポジトリの検索に失敗しました: %v", err)
	}

	if len(gitDirs) == 0 {
		fmt.Println("対象となる.gitディレクトリが見つかりませんでした。")
		return nil
	}

	fmt.Printf("見つかったGitリポジトリ: %d個\n", len(gitDirs))
	for _, gitDir := range gitDirs {
		fmt.Printf("  - %s\n", gitDir)
	}
	fmt.Println()

	// Rewriterを作成
	gitRewriter := c.createRewriter(config)

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
	return c.displayResults(successCount, len(gitDirs), failedRepos, pushFailedRepos)
}

// displayConfig は設定情報を表示する
func (c *RewriteCommand) displayConfig(config *config.Config) {
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
}

// createRewriter はRewriterを作成・設定する
func (c *RewriteCommand) createRewriter(config *config.Config) *rewriter.Rewriter {
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

	return gitRewriter
}

// displayResults は最終結果を表示する
func (c *RewriteCommand) displayResults(successCount, totalCount int, failedRepos, pushFailedRepos []string) error {
	fmt.Printf("\n=== 実行結果 ===\n")
	fmt.Printf("完全成功: %d/%d リポジトリ\n", successCount, totalCount)

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
		return fmt.Errorf("一部のリポジトリで処理に失敗しました")
	} else {
		fmt.Println("すべてのリポジトリで履歴書き換えとプッシュが正常に完了しました。")
		return nil
	}
}
