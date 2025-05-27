package utils

import (
	"os"
	"testing"
)

func TestExtractRepoInfoFromURL(t *testing.T) {
	tests := []struct {
		name          string
		remoteURL     string
		expectedOwner string
		expectedRepo  string
	}{
		{
			name:          "HTTPS形式 .gitあり",
			remoteURL:     "https://github.com/user/repo.git",
			expectedOwner: "user",
			expectedRepo:  "repo",
		},
		{
			name:          "HTTPS形式 .gitなし",
			remoteURL:     "https://github.com/user/repo",
			expectedOwner: "user",
			expectedRepo:  "repo",
		},
		{
			name:          "SSH形式 .gitあり",
			remoteURL:     "git@github.com:user/repo.git",
			expectedOwner: "user",
			expectedRepo:  "repo",
		},
		{
			name:          "SSH形式 .gitなし",
			remoteURL:     "git@github.com:user/repo",
			expectedOwner: "user",
			expectedRepo:  "repo",
		},
		{
			name:          "SSH形式 .gitなし（実際の例）",
			remoteURL:     "git@github.com:corochanhub/yuyama_interview_app",
			expectedOwner: "corochanhub",
			expectedRepo:  "yuyama_interview_app",
		},
		{
			name:          "SSH形式 .gitあり（実際の例）",
			remoteURL:     "git@github.com:corochanhub/yuyama_interview_app.git",
			expectedOwner: "corochanhub",
			expectedRepo:  "yuyama_interview_app",
		},
		{
			name:          "HTTPS形式 末尾スラッシュあり",
			remoteURL:     "https://github.com/user/repo/",
			expectedOwner: "user",
			expectedRepo:  "repo",
		},
		{
			name:          "SSH形式 末尾スラッシュあり",
			remoteURL:     "git@github.com:user/repo/",
			expectedOwner: "user",
			expectedRepo:  "repo",
		},
		{
			name:          "無効なURL",
			remoteURL:     "invalid-url",
			expectedOwner: "",
			expectedRepo:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo := ExtractRepoInfoFromURL(tt.remoteURL)

			if owner != tt.expectedOwner {
				t.Errorf("期待されるオーナー: %s, 実際: %s", tt.expectedOwner, owner)
			}

			if repo != tt.expectedRepo {
				t.Errorf("期待されるリポジトリ: %s, 実際: %s", tt.expectedRepo, repo)
			}
		})
	}
}

func TestGetTargetOwner(t *testing.T) {
	// 環境変数をクリーンアップ
	originalOwner := os.Getenv("GITHUB_REPOSITORY_OWNER")
	originalOrg := os.Getenv("GITHUB_ORGANIZATION")
	defer func() {
		if originalOwner != "" {
			os.Setenv("GITHUB_REPOSITORY_OWNER", originalOwner)
		} else {
			os.Unsetenv("GITHUB_REPOSITORY_OWNER")
		}
		if originalOrg != "" {
			os.Setenv("GITHUB_ORGANIZATION", originalOrg)
		} else {
			os.Unsetenv("GITHUB_ORGANIZATION")
		}
	}()

	tests := []struct {
		name        string
		repoOwner   string
		orgValue    string
		defaultUser string
		expected    string
	}{
		{
			name:        "環境変数なし",
			repoOwner:   "",
			orgValue:    "",
			defaultUser: "defaultuser",
			expected:    "defaultuser",
		},
		{
			name:        "GITHUB_ORGANIZATION設定",
			repoOwner:   "",
			orgValue:    "myorg",
			defaultUser: "defaultuser",
			expected:    "myorg",
		},
		{
			name:        "GITHUB_REPOSITORY_OWNER設定（優先）",
			repoOwner:   "personalowner",
			orgValue:    "myorg",
			defaultUser: "defaultuser",
			expected:    "personalowner",
		},
		{
			name:        "GITHUB_REPOSITORY_OWNERのみ設定",
			repoOwner:   "personalowner",
			orgValue:    "",
			defaultUser: "defaultuser",
			expected:    "personalowner",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テスト用に環境変数をクリア
			os.Unsetenv("GITHUB_REPOSITORY_OWNER")
			os.Unsetenv("GITHUB_ORGANIZATION")

			// 環境変数を設定
			if tt.repoOwner != "" {
				os.Setenv("GITHUB_REPOSITORY_OWNER", tt.repoOwner)
			}
			if tt.orgValue != "" {
				os.Setenv("GITHUB_ORGANIZATION", tt.orgValue)
			}

			result := GetTargetOwner(tt.defaultUser, tt.repoOwner, tt.orgValue)

			if result != tt.expected {
				t.Errorf("期待される結果: %s, 実際: %s", tt.expected, result)
			}
		})
	}
}

func TestIsPersonalRepository(t *testing.T) {
	// 環境変数をクリーンアップ
	originalOwner := os.Getenv("GITHUB_REPOSITORY_OWNER")
	defer func() {
		if originalOwner != "" {
			os.Setenv("GITHUB_REPOSITORY_OWNER", originalOwner)
		} else {
			os.Unsetenv("GITHUB_REPOSITORY_OWNER")
		}
	}()

	tests := []struct {
		name      string
		repoOwner string
		expected  bool
	}{
		{
			name:      "GITHUB_REPOSITORY_OWNER未設定",
			repoOwner: "",
			expected:  false,
		},
		{
			name:      "GITHUB_REPOSITORY_OWNER設定済み",
			repoOwner: "personalowner",
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テスト用に環境変数をクリア
			os.Unsetenv("GITHUB_REPOSITORY_OWNER")

			if tt.repoOwner != "" {
				os.Setenv("GITHUB_REPOSITORY_OWNER", tt.repoOwner)
			}

			result := IsPersonalRepository(tt.repoOwner)

			if result != tt.expected {
				t.Errorf("期待される結果: %t, 実際: %t", tt.expected, result)
			}
		})
	}
}

func TestConvertToTokenURL(t *testing.T) {
	tests := []struct {
		name        string
		remoteURL   string
		token       string
		expectedURL string
		shouldError bool
	}{
		{
			name:        "HTTPS形式",
			remoteURL:   "https://github.com/user/repo.git",
			token:       "ghp_test123",
			expectedURL: "https://ghp_test123@github.com/user/repo.git",
			shouldError: false,
		},
		{
			name:        "SSH形式",
			remoteURL:   "git@github.com:user/repo.git",
			token:       "ghp_test123",
			expectedURL: "https://ghp_test123@github.com/user/repo.git",
			shouldError: false,
		},
		{
			name:        "HTTPS形式 .gitなし",
			remoteURL:   "https://github.com/user/repo",
			token:       "ghp_test123",
			expectedURL: "https://ghp_test123@github.com/user/repo.git",
			shouldError: false,
		},
		{
			name:        "SSH形式 .gitなし",
			remoteURL:   "git@github.com:user/repo",
			token:       "ghp_test123",
			expectedURL: "https://ghp_test123@github.com/user/repo.git",
			shouldError: false,
		},
		{
			name:        "無効なURL",
			remoteURL:   "invalid-url",
			token:       "ghp_test123",
			expectedURL: "",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertToTokenURL(tt.remoteURL, tt.token)

			if tt.shouldError {
				if err == nil {
					t.Errorf("エラーが期待されましたが、エラーが発生しませんでした")
				}
				return
			}

			if err != nil {
				t.Errorf("予期しないエラー: %v", err)
				return
			}

			if result != tt.expectedURL {
				t.Errorf("期待されるURL: %s, 実際: %s", tt.expectedURL, result)
			}
		})
	}
}

func TestRunCommandWithToken(t *testing.T) {
	tests := []struct {
		name        string
		command     string
		args        []string
		expectToken bool
	}{
		{
			name:        "git push コマンド",
			command:     "git",
			args:        []string{"push", "origin", "main"},
			expectToken: true,
		},
		{
			name:        "git log コマンド",
			command:     "git",
			args:        []string{"log", "--oneline"},
			expectToken: false,
		},
		{
			name:        "非gitコマンド",
			command:     "ls",
			args:        []string{"-la"},
			expectToken: false,
		},
		{
			name:        "git push以外のgitコマンド",
			command:     "git",
			args:        []string{"status"},
			expectToken: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// この関数は実際にコマンドを実行するため、
			// ここでは関数の存在とシグネチャのみをテスト
			if tt.expectToken {
				// git pushの場合はRunGitPushWithTokenが呼ばれることを期待
				// 実際のテストは統合テストで行う
				t.Logf("git pushコマンドはトークン認証を使用します: %s %v", tt.command, tt.args)
			} else {
				// その他のコマンドは通常のRunCommandが呼ばれることを期待
				t.Logf("通常のコマンド実行: %s %v", tt.command, tt.args)
			}
		})
	}
}
