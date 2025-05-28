package rewriter

import (
	"testing"
)

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

// TestSetDisableActionsOption はSetDisableActionsOptionメソッドをテストする
func TestSetDisableActionsOption(t *testing.T) {
	rewriter := NewRewriter("test-token", "testuser", "test@example.com")

	// デフォルトはtrue（Actions制御有効）
	if !rewriter.DisableActions {
		t.Error("デフォルトのDisableActionsはtrueであるべきです")
	}

	// falseに設定（Actions制御無効）
	rewriter.SetDisableActionsOption(false)
	if rewriter.DisableActions {
		t.Error("SetDisableActionsOption(false)後、DisableActionsはfalseであるべきです")
	}

	// trueに設定（Actions制御有効）
	rewriter.SetDisableActionsOption(true)
	if !rewriter.DisableActions {
		t.Error("SetDisableActionsOption(true)後、DisableActionsはtrueであるべきです")
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
			if !rewriter.DisableActions {
				t.Error("デフォルトのDisableActionsはtrueであるべきです")
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
