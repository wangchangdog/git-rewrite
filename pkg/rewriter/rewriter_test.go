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
			expectedURL: "https://github.com/testuser/repo",
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
			expectedURL: "https://github.com/myorg/repo",
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

func TestGenerateNewRemoteURLWithRepositoryOwner(t *testing.T) {
	// テスト用のRewriterを作成
	rewriter := &Rewriter{
		GitHubUser: "testuser",
	}

	// GITHUB_REPOSITORY_OWNER環境変数を設定（優先度最高）
	t.Setenv("GITHUB_REPOSITORY_OWNER", "personalowner")
	t.Setenv("GITHUB_ORGANIZATION", "myorg") // これは無視される

	tests := []struct {
		name        string
		inputURL    string
		expectedURL string
	}{
		{
			name:        "HTTPS形式 個人リポジトリ所有者設定",
			inputURL:    "https://github.com/olduser/repo.git",
			expectedURL: "https://github.com/personalowner/repo",
		},
		{
			name:        "SSH形式 個人リポジトリ所有者設定",
			inputURL:    "git@github.com:olduser/repo",
			expectedURL: "git@github.com:personalowner/repo.git",
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

// TestSetPushAllOption はSetPushAllOptionメソッドをテストする
func TestSetPushAllOption(t *testing.T) {
	rewriter := NewRewriter("test-token", "testuser", "test@example.com")

	// デフォルトはfalse
	if rewriter.PushAll {
		t.Error("デフォルトのPushAllはfalseであるべきです")
	}

	// trueに設定
	rewriter.SetPushAllOption(true)
	if !rewriter.PushAll {
		t.Error("SetPushAllOption(true)後、PushAllはtrueであるべきです")
	}

	// falseに設定
	rewriter.SetPushAllOption(false)
	if rewriter.PushAll {
		t.Error("SetPushAllOption(false)後、PushAllはfalseであるべきです")
	}
}

// TestNewRewriterPushAllDefault は新しいRewriterのPushAllデフォルト値をテストする
func TestNewRewriterPushAllDefault(t *testing.T) {
	tests := []struct {
		name       string
		createFunc func() *Rewriter
	}{
		{
			name: "NewRewriter",
			createFunc: func() *Rewriter {
				return NewRewriter("test-token", "testuser", "test@example.com")
			},
		},
		{
			name: "NewRewriterWithConfig",
			createFunc: func() *Rewriter {
				return NewRewriterWithConfig("test-token", "testuser", "test@example.com", "config.json")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rewriter := tt.createFunc()
			if rewriter.PushAll {
				t.Errorf("%s: デフォルトのPushAllはfalseであるべきです", tt.name)
			}
		})
	}
}

// TestSetOwnershipConfig はSetOwnershipConfigメソッドをテストする
func TestSetOwnershipConfig(t *testing.T) {
	rewriter := NewRewriter("test-token", "testuser", "test@example.com")

	// 初期状態は空
	if rewriter.Owner != "" || rewriter.Organization != "" {
		t.Error("初期状態では Owner と Organization は空であるべきです")
	}

	// 設定を変更
	rewriter.SetOwnershipConfig("testowner", "testorg")

	if rewriter.Owner != "testowner" {
		t.Errorf("Owner が正しく設定されていません。期待値: testowner, 実際: %s", rewriter.Owner)
	}
	if rewriter.Organization != "testorg" {
		t.Errorf("Organization が正しく設定されていません。期待値: testorg, 実際: %s", rewriter.Organization)
	}
}

// TestSetPrivateOption はSetPrivateOptionメソッドをテストする
func TestSetPrivateOption(t *testing.T) {
	rewriter := NewRewriter("test-token", "testuser", "test@example.com")

	// デフォルトはtrue（プライベート）
	if !rewriter.Private {
		t.Error("デフォルトのPrivateはtrueであるべきです")
	}

	// falseに設定（パブリック）
	rewriter.SetPrivateOption(false)
	if rewriter.Private {
		t.Error("SetPrivateOption(false)後、Privateはfalseであるべきです")
	}

	// trueに設定（プライベート）
	rewriter.SetPrivateOption(true)
	if !rewriter.Private {
		t.Error("SetPrivateOption(true)後、Privateはtrueであるべきです")
	}
}

// TestSetCollaboratorsFromString はSetCollaboratorsFromStringメソッドをテストする
func TestSetCollaboratorsFromString(t *testing.T) {
	rewriter := NewRewriter("test-token", "testuser", "test@example.com")

	// 初期状態は空
	if rewriter.CollaboratorsString != "" {
		t.Error("初期状態では CollaboratorsString は空であるべきです")
	}

	// 設定を変更
	collaborators := "user1:push,user2:admin"
	rewriter.SetCollaboratorsFromString(collaborators)

	if rewriter.CollaboratorsString != collaborators {
		t.Errorf("CollaboratorsString が正しく設定されていません。期待値: %s, 実際: %s", collaborators, rewriter.CollaboratorsString)
	}
}

// TestNewRewriterDefaults は新しいRewriterのデフォルト値をテストする
func TestNewRewriterDefaults(t *testing.T) {
	tests := []struct {
		name       string
		createFunc func() *Rewriter
	}{
		{
			name: "NewRewriter",
			createFunc: func() *Rewriter {
				return NewRewriter("test-token", "testuser", "test@example.com")
			},
		},
		{
			name: "NewRewriterWithConfig",
			createFunc: func() *Rewriter {
				return NewRewriterWithConfig("test-token", "testuser", "test@example.com", "config.json")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rewriter := tt.createFunc()

			// デフォルト値の確認
			if !rewriter.Private {
				t.Error("デフォルトのPrivateはtrueであるべきです")
			}
			if rewriter.PushAll {
				t.Error("デフォルトのPushAllはfalseであるべきです")
			}
			if rewriter.Owner != "" {
				t.Error("デフォルトのOwnerは空であるべきです")
			}
			if rewriter.Organization != "" {
				t.Error("デフォルトのOrganizationは空であるべきです")
			}
			if rewriter.CollaboratorsString != "" {
				t.Error("デフォルトのCollaboratorsStringは空であるべきです")
			}
		})
	}
}

// TestRewriterTokenStorage はRewriterがトークンを正しく保存することをテストする
func TestRewriterTokenStorage(t *testing.T) {
	token := "ghp_test123456"
	user := "testuser"
	email := "test@example.com"

	tests := []struct {
		name       string
		createFunc func() *Rewriter
	}{
		{
			name: "NewRewriter",
			createFunc: func() *Rewriter {
				return NewRewriter(token, user, email)
			},
		},
		{
			name: "NewRewriterWithConfig",
			createFunc: func() *Rewriter {
				return NewRewriterWithConfig(token, user, email, "config.json")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rewriter := tt.createFunc()

			if rewriter.GitHubToken != token {
				t.Errorf("GitHubTokenが正しく設定されていません。期待値: %s, 実際: %s", token, rewriter.GitHubToken)
			}
			if rewriter.GitHubUser != user {
				t.Errorf("GitHubUserが正しく設定されていません。期待値: %s, 実際: %s", user, rewriter.GitHubUser)
			}
			if rewriter.GitHubEmail != email {
				t.Errorf("GitHubEmailが正しく設定されていません。期待値: %s, 実際: %s", email, rewriter.GitHubEmail)
			}
		})
	}
}
