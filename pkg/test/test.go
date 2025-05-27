package test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"git-rewrite-and-go/pkg/utils"
)

// RunTests は基本的なテストを実行する
func RunTests() error {
	fmt.Println("=== Git Rewrite Tools テスト実行 ===")

	// 1. ユーティリティ関数のテスト
	if err := testUtilityFunctions(); err != nil {
		return fmt.Errorf("ユーティリティ関数テスト失敗: %v", err)
	}

	// 2. Git操作のテスト
	if err := testGitOperations(); err != nil {
		return fmt.Errorf("Git操作テスト失敗: %v", err)
	}

	fmt.Println("\n=== すべてのテストが成功しました ===")
	return nil
}

// testUtilityFunctions はユーティリティ関数をテストする
func testUtilityFunctions() error {
	fmt.Println("\n--- ユーティリティ関数テスト ---")

	// SafeDecode テスト
	fmt.Println("1. SafeDecode テスト:")
	testData := []byte("こんにちは")
	result := utils.SafeDecode(testData)
	if result != "こんにちは" {
		return fmt.Errorf("SafeDecode テスト失敗: 期待値 'こんにちは', 実際 '%s'", result)
	}
	fmt.Println("   ✓ SafeDecode テスト成功")

	// ExtractRepoInfoFromURL テスト
	fmt.Println("2. ExtractRepoInfoFromURL テスト:")
	testCases := []struct {
		url      string
		owner    string
		repo     string
		expected bool
	}{
		{"https://github.com/testuser/testrepo.git", "testuser", "testrepo", true},
		{"git@github.com:testuser/testrepo.git", "testuser", "testrepo", true},
		{"https://github.com/testuser/testrepo", "testuser", "testrepo", true},
		{"invalid-url", "", "", false},
	}

	for _, tc := range testCases {
		owner, repo := utils.ExtractRepoInfoFromURL(tc.url)
		if tc.expected {
			if owner != tc.owner || repo != tc.repo {
				return fmt.Errorf("ExtractRepoInfoFromURL テスト失敗: URL '%s', 期待値 (%s, %s), 実際 (%s, %s)",
					tc.url, tc.owner, tc.repo, owner, repo)
			}
		} else {
			if owner != "" || repo != "" {
				return fmt.Errorf("ExtractRepoInfoFromURL テスト失敗: 無効なURL '%s' で空文字列が期待されるが (%s, %s) が返された",
					tc.url, owner, repo)
			}
		}
	}
	fmt.Println("   ✓ ExtractRepoInfoFromURL テスト成功")

	// FileExists テスト
	fmt.Println("3. FileExists テスト:")
	tempFile, err := ioutil.TempFile("", "test_file_")
	if err != nil {
		return fmt.Errorf("一時ファイル作成エラー: %v", err)
	}
	tempFile.Close()
	defer os.Remove(tempFile.Name())

	if !utils.FileExists(tempFile.Name()) {
		return fmt.Errorf("FileExists テスト失敗: 存在するファイルが検出されない")
	}

	if utils.FileExists("/nonexistent/file/path") {
		return fmt.Errorf("FileExists テスト失敗: 存在しないファイルが検出された")
	}
	fmt.Println("   ✓ FileExists テスト成功")

	return nil
}

// testGitOperations はGit操作をテストする
func testGitOperations() error {
	fmt.Println("\n--- Git操作テスト ---")

	// 一時ディレクトリを作成
	tempDir, err := ioutil.TempDir("", "git_test_")
	if err != nil {
		return fmt.Errorf("一時ディレクトリ作成エラー: %v", err)
	}
	defer os.RemoveAll(tempDir)

	fmt.Printf("テストディレクトリ: %s\n", tempDir)

	// Git リポジトリを初期化
	fmt.Println("1. Git リポジトリ初期化テスト:")
	stdout, stderr, err := utils.RunCommand(tempDir, "git", "init")
	if err != nil {
		return fmt.Errorf("git init 失敗: %v, stderr: %s", err, stderr)
	}
	if !strings.Contains(stdout, "Initialized") && !strings.Contains(stderr, "Initialized") {
		return fmt.Errorf("git init の出力が期待されるものと異なる: %s", stdout)
	}
	fmt.Println("   ✓ Git リポジトリ初期化成功")

	// ファイル作成とコミット
	fmt.Println("2. ファイル作成とコミットテスト:")
	testFilePath := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFilePath, []byte("test content"), 0644); err != nil {
		return fmt.Errorf("テストファイル作成エラー: %v", err)
	}

	// git add
	_, stderr, err = utils.RunCommand(tempDir, "git", "add", "test.txt")
	if err != nil {
		return fmt.Errorf("git add 失敗: %v, stderr: %s", err, stderr)
	}

	// git commit
	_, stderr, err = utils.RunCommand(tempDir, "git", "-c", "user.name=Test User", "-c", "user.email=test@example.com", "commit", "-m", "test commit")
	if err != nil {
		return fmt.Errorf("git commit 失敗: %v, stderr: %s", err, stderr)
	}
	fmt.Println("   ✓ ファイル作成とコミット成功")

	// リモート設定テスト
	fmt.Println("3. リモート設定テスト:")
	_, stderr, err = utils.RunCommand(tempDir, "git", "remote", "add", "origin", "https://github.com/testuser/testrepo.git")
	if err != nil {
		return fmt.Errorf("git remote add 失敗: %v, stderr: %s", err, stderr)
	}

	// リモートURL取得
	stdout, stderr, err = utils.RunCommand(tempDir, "git", "remote", "get-url", "origin")
	if err != nil {
		return fmt.Errorf("git remote get-url 失敗: %v, stderr: %s", err, stderr)
	}

	expectedURL := "https://github.com/testuser/testrepo.git"
	if strings.TrimSpace(stdout) != expectedURL {
		return fmt.Errorf("リモートURL が期待値と異なる: 期待値 '%s', 実際 '%s'", expectedURL, strings.TrimSpace(stdout))
	}
	fmt.Println("   ✓ リモート設定成功")

	// FindGitDirs テスト
	fmt.Println("4. FindGitDirs テスト:")
	gitDirs, err := utils.FindGitDirs(tempDir)
	if err != nil {
		return fmt.Errorf("FindGitDirs エラー: %v", err)
	}

	if len(gitDirs) != 1 {
		return fmt.Errorf("FindGitDirs テスト失敗: 期待値 1, 実際 %d", len(gitDirs))
	}

	if gitDirs[0] != tempDir {
		return fmt.Errorf("FindGitDirs テスト失敗: 期待値 '%s', 実際 '%s'", tempDir, gitDirs[0])
	}
	fmt.Println("   ✓ FindGitDirs テスト成功")

	return nil
}
