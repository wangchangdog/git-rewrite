package git

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"git-rewrite-and-go/pkg/utils"
)

// TestRewriteHistory はRewriteHistory関数をテストする
func TestRewriteHistory(t *testing.T) {
	tests := []struct {
		name        string
		gitDir      string
		githubUser  string
		githubEmail string
		shouldError bool
		description string
	}{
		{
			name:        "空のgitDir",
			gitDir:      "",
			githubUser:  "testuser",
			githubEmail: "test@example.com",
			shouldError: true,
			description: "空のgitDirでエラーが発生することを確認",
		},
		{
			name:        "空のgithubUser",
			gitDir:      "/tmp/test",
			githubUser:  "",
			githubEmail: "test@example.com",
			shouldError: true,
			description: "空のgithubUserでエラーが発生することを確認",
		},
		{
			name:        "空のgithubEmail",
			gitDir:      "/tmp/test",
			githubUser:  "testuser",
			githubEmail: "",
			shouldError: true,
			description: "空のgithubEmailでエラーが発生することを確認",
		},
		{
			name:        "存在しないディレクトリ",
			gitDir:      "/nonexistent/directory",
			githubUser:  "testuser",
			githubEmail: "test@example.com",
			shouldError: true,
			description: "存在しないディレクトリでエラーが発生することを確認",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RewriteHistory(tt.gitDir, tt.githubUser, tt.githubEmail)

			if tt.shouldError && err == nil {
				t.Errorf("エラーが期待されましたが、エラーが発生しませんでした: %s", tt.description)
			}
			if !tt.shouldError && err != nil {
				t.Errorf("エラーが期待されませんでしたが、エラーが発生しました: %v (%s)", err, tt.description)
			}
		})
	}
}

// TestRewriteHistoryWithNonGitDirectory はGitリポジトリでないディレクトリでのテストを行う
func TestRewriteHistoryWithNonGitDirectory(t *testing.T) {
	// 一時ディレクトリを作成
	tempDir, err := ioutil.TempDir("", "non_git_test_")
	if err != nil {
		t.Fatalf("一時ディレクトリ作成エラー: %v", err)
	}
	defer os.RemoveAll(tempDir)

	err = RewriteHistory(tempDir, "testuser", "test@example.com")
	if err == nil {
		t.Error("Gitリポジトリでないディレクトリでエラーが期待されましたが、エラーが発生しませんでした")
	}

	expectedErrorMsg := "はGitリポジトリではありません"
	if err != nil && !contains(err.Error(), expectedErrorMsg) {
		t.Errorf("期待されるエラーメッセージが含まれていません。期待: '%s', 実際: '%s'", expectedErrorMsg, err.Error())
	}
}

// TestCreateInitialCommit はCreateInitialCommit関数をテストする
func TestCreateInitialCommit(t *testing.T) {
	tests := []struct {
		name        string
		gitDir      string
		githubUser  string
		githubEmail string
		shouldError bool
		description string
	}{
		{
			name:        "存在しないディレクトリ",
			gitDir:      "/nonexistent/directory",
			githubUser:  "testuser",
			githubEmail: "test@example.com",
			shouldError: true,
			description: "存在しないディレクトリでエラーが発生することを確認",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CreateInitialCommit(tt.gitDir, tt.githubUser, tt.githubEmail)

			if tt.shouldError && err == nil {
				t.Errorf("エラーが期待されましたが、エラーが発生しませんでした: %s", tt.description)
			}
			if !tt.shouldError && err != nil {
				t.Errorf("エラーが期待されませんでしたが、エラーが発生しました: %v (%s)", err, tt.description)
			}
		})
	}
}

// TestCreateInitialCommitWithMockGitRepo はモックGitリポジトリでのテストを行う
func TestCreateInitialCommitWithMockGitRepo(t *testing.T) {
	// 一時ディレクトリを作成
	tempDir, err := ioutil.TempDir("", "git_test_")
	if err != nil {
		t.Fatalf("一時ディレクトリ作成エラー: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Gitリポジトリを初期化
	_, _, err = utils.RunCommand(tempDir, "git", "init")
	if err != nil {
		t.Fatalf("git init エラー: %v", err)
	}

	// Git設定を追加（テスト環境で必要）
	_, _, err = utils.RunCommand(tempDir, "git", "config", "user.name", "Test User")
	if err != nil {
		t.Fatalf("git config user.name エラー: %v", err)
	}

	_, _, err = utils.RunCommand(tempDir, "git", "config", "user.email", "test@example.com")
	if err != nil {
		t.Fatalf("git config user.email エラー: %v", err)
	}

	// CreateInitialCommitを実行
	err = CreateInitialCommit(tempDir, "testuser", "test@example.com")
	if err != nil {
		t.Errorf("CreateInitialCommitでエラーが発生しました: %v", err)
	}

	// README.mdが作成されているかチェック
	readmePath := filepath.Join(tempDir, "README.md")
	if !utils.FileExists(readmePath) {
		t.Error("README.mdファイルが作成されていません")
	}

	// コミットが作成されているかチェック
	_, _, err = utils.RunCommand(tempDir, "git", "log", "--oneline", "-1")
	if err != nil {
		t.Errorf("コミットが作成されていません: %v", err)
	}
}

// contains は文字列に部分文字列が含まれているかチェックする
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			func() bool {
				for i := 0; i <= len(s)-len(substr); i++ {
					if s[i:i+len(substr)] == substr {
						return true
					}
				}
				return false
			}())))
}
