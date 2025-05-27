package git

import (
	"os"
	"testing"
)

// TestPushAllBranchesAndTags はPushAllBranchesAndTags関数をテストする
func TestPushAllBranchesAndTags(t *testing.T) {
	// 環境変数をクリーンアップ
	originalDebug := os.Getenv("GIT_REWRITE_DEBUG")
	defer func() {
		if originalDebug != "" {
			os.Setenv("GIT_REWRITE_DEBUG", originalDebug)
		} else {
			os.Unsetenv("GIT_REWRITE_DEBUG")
		}
	}()

	tests := []struct {
		name        string
		gitDir      string
		token       string
		shouldError bool
		description string
	}{
		{
			name:        "空のトークン",
			gitDir:      "/tmp/test",
			token:       "",
			shouldError: true,
			description: "空のトークンでエラーが発生することを確認",
		},
		{
			name:        "無効なディレクトリ",
			gitDir:      "/nonexistent/directory",
			token:       "ghp_test123",
			shouldError: true,
			description: "存在しないディレクトリでエラーが発生することを確認",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := PushAllBranchesAndTags(tt.gitDir, tt.token)

			if tt.shouldError && err == nil {
				t.Errorf("エラーが期待されましたが、エラーが発生しませんでした: %s", tt.description)
			}
			if !tt.shouldError && err != nil {
				t.Errorf("エラーが期待されませんでしたが、エラーが発生しました: %v (%s)", err, tt.description)
			}
		})
	}
}

// TestPushToRemote はPushToRemote関数をテストする
func TestPushToRemote(t *testing.T) {
	tests := []struct {
		name        string
		gitDir      string
		token       string
		shouldError bool
		description string
	}{
		{
			name:        "空のトークン",
			gitDir:      "/tmp/test",
			token:       "",
			shouldError: true,
			description: "空のトークンでエラーが発生することを確認",
		},
		{
			name:        "無効なディレクトリ",
			gitDir:      "/nonexistent/directory",
			token:       "ghp_test123",
			shouldError: true,
			description: "存在しないディレクトリでエラーが発生することを確認",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := PushToRemote(tt.gitDir, tt.token)

			if tt.shouldError && err == nil {
				t.Errorf("エラーが期待されましたが、エラーが発生しませんでした: %s", tt.description)
			}
			if !tt.shouldError && err != nil {
				t.Errorf("エラーが期待されませんでしたが、エラーが発生しました: %v (%s)", err, tt.description)
			}
		})
	}
}

// TestPushFunctionParameters は関数のパラメータ検証をテストする
func TestPushFunctionParameters(t *testing.T) {
	t.Run("PushAllBranchesAndTags パラメータ検証", func(t *testing.T) {
		// 空のgitDirでテスト
		err := PushAllBranchesAndTags("", "token")
		if err == nil {
			t.Error("空のgitDirでエラーが期待されましたが、エラーが発生しませんでした")
		}

		// 空のtokenでテスト
		err = PushAllBranchesAndTags("/tmp", "")
		if err == nil {
			t.Error("空のtokenでエラーが期待されましたが、エラーが発生しませんでした")
		}
	})

	t.Run("PushToRemote パラメータ検証", func(t *testing.T) {
		// 空のgitDirでテスト
		err := PushToRemote("", "token")
		if err == nil {
			t.Error("空のgitDirでエラーが期待されましたが、エラーが発生しませんでした")
		}

		// 空のtokenでテスト
		err = PushToRemote("/tmp", "")
		if err == nil {
			t.Error("空のtokenでエラーが期待されましたが、エラーが発生しませんでした")
		}
	})
}
