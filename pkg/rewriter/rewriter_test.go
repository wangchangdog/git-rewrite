package rewriter

import (
	"testing"
)

func TestGenerateNewRemoteURL(t *testing.T) {
	// テスト用のRewriterを作成
	rewriter := &Rewriter{
		GitHubUser: "testuser",
	}

	tests := []struct {
		name        string
		inputURL    string
		expectedURL string
		shouldError bool
	}{
		{
			name:        "HTTPS形式 .gitなし",
			inputURL:    "https://github.com/olduser/repo",
			expectedURL: "https://github.com/testuser/repo.git",
			shouldError: false,
		},
		{
			name:        "SSH形式 .gitあり",
			inputURL:    "git@github.com:olduser/repo.git",
			expectedURL: "git@github.com:testuser/repo.git",
			shouldError: false,
		},
		{
			name:        "SSH形式 実際の例",
			inputURL:    "git@github.com:corochanhub/yuyama_interview_app",
			expectedURL: "git@github.com:testuser/yuyama_interview_app.git",
			shouldError: false,
		},
		{
			name:        "無効なURL",
			inputURL:    "invalid-url",
			expectedURL: "",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := rewriter.generateNewRemoteURL(tt.inputURL)

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

func TestGenerateNewRemoteURLWithOrganization(t *testing.T) {
	// テスト用のRewriterを作成
	rewriter := &Rewriter{
		GitHubUser: "testuser",
	}

	// GITHUB_ORGANIZATION環境変数を設定
	t.Setenv("GITHUB_ORGANIZATION", "myorg")

	tests := []struct {
		name        string
		inputURL    string
		expectedURL string
	}{
		{
			name:        "HTTPS形式 組織設定",
			inputURL:    "https://github.com/olduser/repo.git",
			expectedURL: "https://github.com/myorg/repo.git",
		},
		{
			name:        "SSH形式 組織設定",
			inputURL:    "git@github.com:olduser/repo",
			expectedURL: "git@github.com:myorg/repo.git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := rewriter.generateNewRemoteURL(tt.inputURL)

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
