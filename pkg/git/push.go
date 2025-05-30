package git

import (
	"fmt"
	"strings"

	"git-rewrite/pkg/utils"
)

// PushAllBranchesAndTags はローカルの全ブランチとタグをリモートにプッシュする
func PushAllBranchesAndTags(gitDir, token string) error {
	fmt.Println("\n--- 全ブランチ・タグのプッシュ ---")

	// 全ブランチをプッシュ（トークン認証使用）
	fmt.Println("🌿 全ブランチをプッシュしています...")
	stdout, stderr, err := utils.RunGitPushWithToken(gitDir, token, "--all", "origin")
	if err != nil {
		// エラーが発生した場合でも、強制プッシュを試行
		fmt.Println("⚠️  通常のブランチプッシュでエラーが発生しました。強制プッシュを試行します...")
		stdout, stderr, err = utils.RunGitPushWithToken(gitDir, token, "--force", "--all", "origin")
		if err != nil {
			fmt.Printf("❌ 全ブランチの強制プッシュに失敗しました: %v\n", err)
			if stderr != "" {
				fmt.Printf("エラー詳細: %s\n", stderr)
			}
			// ブランチプッシュが失敗してもタグプッシュは試行する
		} else {
			fmt.Println("✅ 全ブランチの強制プッシュが完了しました。")
		}
	} else {
		fmt.Println("✅ 全ブランチのプッシュが完了しました。")
	}

	if stdout != "" {
		fmt.Printf("ブランチプッシュ結果: %s\n", stdout)
	}

	// 全タグをプッシュ（トークン認証使用）
	fmt.Println("🏷️  全タグをプッシュしています...")
	stdout, stderr, err = utils.RunGitPushWithToken(gitDir, token, "--tags", "origin")
	if err != nil {
		// エラーが発生した場合でも、強制プッシュを試行
		fmt.Println("⚠️  通常のタグプッシュでエラーが発生しました。強制プッシュを試行します...")
		stdout, stderr, err = utils.RunGitPushWithToken(gitDir, token, "--force", "--tags", "origin")
		if err != nil {
			fmt.Printf("❌ 全タグの強制プッシュに失敗しました: %v\n", err)
			if stderr != "" {
				fmt.Printf("エラー詳細: %s\n", stderr)
			}
			return fmt.Errorf("タグプッシュエラー: %v", err)
		} else {
			fmt.Println("✅ 全タグの強制プッシュが完了しました。")
		}
	} else {
		fmt.Println("✅ 全タグのプッシュが完了しました。")
	}

	if stdout != "" {
		fmt.Printf("タグプッシュ結果: %s\n", stdout)
	}

	fmt.Println("🚀 全ブランチ・タグのプッシュが完了しました。")
	return nil
}

// PushToRemote はリモートにプッシュする
func PushToRemote(gitDir, token string) error {
	// 現在のブランチを取得
	stdout, _, err := utils.RunCommand(gitDir, "git", "branch", "--show-current")
	if err != nil {
		return fmt.Errorf("ブランチ取得エラー: %v", err)
	}

	currentBranch := strings.TrimSpace(stdout)
	fmt.Printf("現在のブランチ: %s\n", currentBranch)

	// git push origin HEADを実行（トークン認証使用）
	fmt.Println("リモートにプッシュしています...")
	stdout, stderr, err := utils.RunGitPushWithToken(gitDir, token, "origin", "HEAD")
	if err != nil {
		// pushでエラーが出る場合は force pushを試行
		fmt.Println("⚠️  プッシュエラーが発生しました。強制的にプッシュを試行します...")
		stdout, stderr, err = utils.RunGitPushWithToken(gitDir, token, "--force", "origin", "HEAD")
		if err != nil {
			return fmt.Errorf("強制プッシュエラー: %v\nstderr: %s", err, stderr)
		}
		fmt.Println("✅ 強制プッシュが完了しました。")
	}

	if stdout != "" {
		fmt.Printf("プッシュ結果: %s\n", stdout)
	}
	if stderr != "" {
		fmt.Printf("プッシュ情報: %s\n", stderr)
	}

	fmt.Println("✅ リモートへのプッシュが完了しました。")
	return nil
}
