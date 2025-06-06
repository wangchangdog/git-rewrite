package demo

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"git-rewrite/pkg/github"
	"git-rewrite/pkg/rewriter"
	"git-rewrite/pkg/utils"
)

// RunDemo はリモートリポジトリ作成機能のデモを実行する
func RunDemo(githubToken string) error {
	if githubToken == "" {
		return fmt.Errorf("GitHub トークンが指定されていません")
	}

	// 環境変数をチェック
	githubUser, _, err := utils.CheckEnvironmentVariables()
	if err != nil {
		return err
	}

	fmt.Printf("GitHub ユーザー: %s\n", githubUser)
	fmt.Printf("GitHub トークン: 設定済み\n")
	fmt.Println()

	// GitHub クライアントを作成
	client := github.NewClient(githubToken)

	// テスト用の一時ディレクトリを作成
	tempDir, err := ioutil.TempDir("", "demo_repo_")
	if err != nil {
		return fmt.Errorf("一時ディレクトリ作成エラー: %v", err)
	}
	defer os.RemoveAll(tempDir)

	fmt.Printf("テストディレクトリ: %s\n", tempDir)

	// Gitリポジトリを初期化
	if _, _, err := utils.RunCommand(tempDir, "git", "init"); err != nil {
		return fmt.Errorf("git init エラー: %v", err)
	}

	// 初期ファイルを作成
	readmePath := filepath.Join(tempDir, "README.md")
	content := "# Demo Repository\n\nThis is a demo repository for testing remote creation.\n"
	if err := os.WriteFile(readmePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("README.md作成エラー: %v", err)
	}

	// 初期コミット
	if _, _, err := utils.RunCommand(tempDir, "git", "add", "README.md"); err != nil {
		return fmt.Errorf("git add エラー: %v", err)
	}

	if _, _, err := utils.RunCommand(tempDir, "git", "-c", "user.name=Demo User", "-c", "user.email=demo@example.com", "commit", "-m", "Initial commit"); err != nil {
		return fmt.Errorf("git commit エラー: %v", err)
	}

	// デモ用のリモートURLを設定（存在しないリポジトリ）
	demoRepoName := fmt.Sprintf("demo-repo-%s", filepath.Base(tempDir))
	remoteURL := fmt.Sprintf("https://github.com/%s/%s.git", githubUser, demoRepoName)

	fmt.Printf("リモートURL設定: %s\n", remoteURL)
	if _, _, err := utils.RunCommand(tempDir, "git", "remote", "add", "origin", remoteURL); err != nil {
		return fmt.Errorf("git remote add エラー: %v", err)
	}

	fmt.Println("\n=== リモートリポジトリ作成機能のテスト ===")

	// URL解析のテスト
	fmt.Println("1. URL解析テスト:")
	owner, repoName := utils.ExtractRepoInfoFromURL(remoteURL)
	fmt.Printf("   所有者: %s\n", owner)
	fmt.Printf("   リポジトリ名: %s\n", repoName)

	// リポジトリ存在確認のテスト
	fmt.Println("\n2. リポジトリ存在確認テスト:")
	exists, err := client.CheckRepoExists(owner, repoName)
	if err != nil {
		return fmt.Errorf("リポジトリ存在確認エラー: %v", err)
	}
	fmt.Printf("   存在確認結果: %s\n", map[bool]string{true: "存在する", false: "存在しない"}[exists])

	// リポジトリ作成のテスト（存在しない場合のみ）
	if !exists {
		fmt.Println("\n3. リポジトリ作成テスト:")
		if err := client.CreateRepoWithCollaborators(owner, repoName, true, ""); err != nil {
			fmt.Printf("   ✗ リポジトリの作成に失敗しました: %v\n", err)
		} else {
			fmt.Println("   ✅ リポジトリが正常に作成されました。")

			// 作成後の存在確認
			fmt.Println("\n4. 作成後の存在確認:")
			existsAfter, err := client.CheckRepoExists(owner, repoName)
			if err != nil {
				return fmt.Errorf("作成後の存在確認エラー: %v", err)
			}
			fmt.Printf("   存在確認結果: %s\n", map[bool]string{true: "存在する", false: "存在しない"}[existsAfter])

			// プッシュのテスト（トークン認証使用）
			fmt.Println("\n5. プッシュテスト:")
			if _, _, err := utils.RunGitPushWithToken(tempDir, githubToken, "-u", "origin", "master"); err != nil {
				fmt.Printf("   ✗ プッシュに失敗しました: %v\n", err)
			} else {
				fmt.Println("   ✅ プッシュが正常に完了しました。")
			}
		}
	} else {
		fmt.Println("\n3. リポジトリ作成テスト: スキップ（既に存在）")
	}

	fmt.Printf("\n=== デモ完了 ===\n")
	fmt.Printf("作成されたリポジトリ（存在する場合）: https://github.com/%s/%s\n", owner, repoName)

	return nil
}

// RunEmptyRepoDemo は空のリポジトリでの初期コミット作成テストを実行する
func RunEmptyRepoDemo(githubToken, githubUser, githubEmail string) error {
	// 空のリポジトリでの初期コミット作成テスト
	fmt.Println("\n=== 空のリポジトリでの初期コミット作成テスト ===")

	tempDir, err := ioutil.TempDir("", "empty_demo_repo_")
	if err != nil {
		return fmt.Errorf("一時ディレクトリ作成エラー: %v", err)
	}
	defer os.RemoveAll(tempDir)

	fmt.Printf("空のテストディレクトリ: %s\n", tempDir)

	// Gitリポジトリを初期化（コミットなし）
	if _, _, err := utils.RunCommand(tempDir, "git", "init"); err != nil {
		return fmt.Errorf("git init エラー: %v", err)
	}

	emptyRepoName := fmt.Sprintf("empty-demo-repo-%s", filepath.Base(tempDir))
	emptyRemoteURL := fmt.Sprintf("https://github.com/%s/%s.git", githubUser, emptyRepoName)

	fmt.Printf("空のリポジトリのリモートURL設定: %s\n", emptyRemoteURL)
	if _, _, err := utils.RunCommand(tempDir, "git", "remote", "add", "origin", emptyRemoteURL); err != nil {
		return fmt.Errorf("git remote add エラー: %v", err)
	}

	// 初期コミット作成テスト
	fmt.Println("\n6. 初期コミット作成テスト:")
	rewriter := rewriter.NewRewriter(githubToken, githubUser, githubEmail)
	if err := rewriter.CreateInitialCommit(tempDir); err != nil {
		fmt.Printf("   ✗ 初期コミットの作成に失敗しました: %v\n", err)
		return err
	}

	fmt.Println("   ✅ 初期コミットが正常に作成されました。")

	// コミット確認
	stdout, _, err := utils.RunCommand(tempDir, "git", "log", "--oneline", "-1")
	if err != nil {
		return fmt.Errorf("git log エラー: %v", err)
	}
	fmt.Printf("   作成されたコミット: %s\n", stdout)

	// README.mdの確認
	readmePath := filepath.Join(tempDir, "README.md")
	if utils.FileExists(readmePath) {
		fmt.Println("   ✅ README.mdファイルが作成されました。")
		content, err := os.ReadFile(readmePath)
		if err != nil {
			return fmt.Errorf("README.md読み取りエラー: %v", err)
		}
		fmt.Printf("   README.md内容: %s...\n", string(content)[:50])
	}

	// 追加ブランチとタグの作成テスト（--push-all機能のデモ用）
	fmt.Println("\n7. 追加ブランチ・タグ作成テスト（--push-all機能デモ用）:")

	// 新しいブランチを作成
	if _, _, err := utils.RunCommand(tempDir, "git", "checkout", "-b", "feature-branch"); err != nil {
		fmt.Printf("   ⚠️  ブランチ作成に失敗しました: %v\n", err)
	} else {
		fmt.Println("   ✅ feature-branchを作成しました。")

		// ブランチで追加のコミット
		featurePath := filepath.Join(tempDir, "feature.txt")
		if err := os.WriteFile(featurePath, []byte("Feature content\n"), 0644); err == nil {
			if _, _, err := utils.RunCommand(tempDir, "git", "add", "feature.txt"); err == nil {
				if _, _, err := utils.RunCommand(tempDir, "git", "-c", "user.name=Demo User", "-c", "user.email=demo@example.com", "commit", "-m", "Add feature"); err == nil {
					fmt.Println("   ✅ feature-branchにコミットを追加しました。")
				}
			}
		}
	}

	// mainブランチに戻る
	if _, _, err := utils.RunCommand(tempDir, "git", "checkout", "master"); err != nil {
		// masterブランチが存在しない場合はmainを試行
		if _, _, err := utils.RunCommand(tempDir, "git", "checkout", "main"); err != nil {
			fmt.Printf("   ⚠️  メインブランチへの切り替えに失敗しました: %v\n", err)
		}
	}

	// タグを作成
	if _, _, err := utils.RunCommand(tempDir, "git", "tag", "v1.0.0"); err != nil {
		fmt.Printf("   ⚠️  タグ作成に失敗しました: %v\n", err)
	} else {
		fmt.Println("   ✅ v1.0.0タグを作成しました。")
	}

	// --push-all機能のテスト
	fmt.Println("\n8. --push-all機能テスト:")
	rewriter.SetPushAllOption(true)
	if err := rewriter.PushAllBranchesAndTags(tempDir); err != nil {
		fmt.Printf("   ⚠️  --push-all機能のテストをスキップしました（リモートが設定されていないため）: %v\n", err)
	} else {
		fmt.Println("   ✅ --push-all機能のテストが完了しました。")
	}

	fmt.Println("\n=== 全テスト完了 ===")
	return nil
}
