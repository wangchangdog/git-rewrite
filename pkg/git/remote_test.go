package git

import (
	"os"
	"testing"
)

func TestGenerateNewRemoteURL(t *testing.T) {
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

	// テスト用に環境変数をクリア
	os.Unsetenv("GITHUB_REPOSITORY_OWNER")
	os.Unsetenv("GITHUB_ORGANIZATION")

	tests := []struct {
		name         string
		remoteURL    string
		githubUser   string
		owner        string
		organization string
		expectedURL  string
		shouldError  bool
	}{
		{
			name:         "HTTPS形式 基本",
			remoteURL:    "https://github.com/olduser/repo",
			githubUser:   "testuser",
			owner:        "",
			organization: "",
			expectedURL:  "https://github.com/testuser/repo",
			shouldError:  false,
		},
		{
			name:         "SSH形式 基本",
			remoteURL:    "git@github.com:olduser/repo.git",
			githubUser:   "testuser",
			owner:        "",
			organization: "",
			expectedURL:  "git@github.com:testuser/repo.git",
			shouldError:  false,
		},
		{
			name:         "組織設定",
			remoteURL:    "https://github.com/olduser/repo.git",
			githubUser:   "testuser",
			owner:        "",
			organization: "myorg",
			expectedURL:  "https://github.com/myorg/repo",
			shouldError:  false,
		},
		{
			name:         "個人リポジトリ所有者設定（最高優先度）",
			remoteURL:    "git@github.com:olduser/repo",
			githubUser:   "testuser",
			owner:        "personalowner",
			organization: "myorg",
			expectedURL:  "git@github.com:personalowner/repo.git",
			shouldError:  false,
		},
		{
			name:         "無効なURL",
			remoteURL:    "invalid-url",
			githubUser:   "testuser",
			owner:        "",
			organization: "",
			expectedURL:  "",
			shouldError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := generateNewRemoteURL(tt.remoteURL, tt.githubUser, tt.owner, tt.organization)

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

func TestGenerateNewRemoteURLWithEnvironmentVariables(t *testing.T) {
	// 環境変数をクリーンアップ
	originalDebug := os.Getenv("GIT_REWRITE_DEBUG")
	originalOwner := os.Getenv("GITHUB_REPOSITORY_OWNER")
	originalOrg := os.Getenv("GITHUB_ORGANIZATION")
	defer func() {
		if originalDebug != "" {
			os.Setenv("GIT_REWRITE_DEBUG", originalDebug)
		} else {
			os.Unsetenv("GIT_REWRITE_DEBUG")
		}
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

	// テスト用に環境変数を設定
	os.Setenv("GIT_REWRITE_DEBUG", "1")
	os.Unsetenv("GITHUB_REPOSITORY_OWNER")
	os.Unsetenv("GITHUB_ORGANIZATION")

	result, err := generateNewRemoteURL("https://github.com/olduser/repo", "testuser", "", "")
	if err != nil {
		t.Errorf("予期しないエラー: %v", err)
		return
	}

	expected := "https://github.com/testuser/repo"
	if result != expected {
		t.Errorf("期待されるURL: %s, 実際: %s", expected, result)
	}
}
