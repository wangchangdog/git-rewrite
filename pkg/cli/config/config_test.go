package config

import (
	"os"
	"testing"
)

// TestParseRewriteArgs はParseRewriteArgs関数をテストする
func TestParseRewriteArgs(t *testing.T) {
	// 環境変数をクリーンアップ
	originalUser := os.Getenv("GITHUB_USER")
	originalEmail := os.Getenv("GITHUB_EMAIL")
	originalOwner := os.Getenv("GITHUB_REPOSITORY_OWNER")
	originalOrg := os.Getenv("GITHUB_ORGANIZATION")
	originalCollaborators := os.Getenv("GITHUB_COLLABORATORS")
	originalDebug := os.Getenv("GIT_REWRITE_DEBUG")

	defer func() {
		restoreEnv("GITHUB_USER", originalUser)
		restoreEnv("GITHUB_EMAIL", originalEmail)
		restoreEnv("GITHUB_REPOSITORY_OWNER", originalOwner)
		restoreEnv("GITHUB_ORGANIZATION", originalOrg)
		restoreEnv("GITHUB_COLLABORATORS", originalCollaborators)
		restoreEnv("GIT_REWRITE_DEBUG", originalDebug)
	}()

	tests := []struct {
		name        string
		args        []string
		shouldError bool
		description string
	}{
		{
			name:        "引数なし",
			args:        []string{},
			shouldError: true,
			description: "GitHubトークンが必要",
		},
		{
			name:        "トークンのみ",
			args:        []string{"ghp_test123"},
			shouldError: true,
			description: "--userと--emailが必要",
		},
		{
			name:        "基本的な引数",
			args:        []string{"ghp_test123", "--user", "testuser", "--email", "test@example.com"},
			shouldError: false,
			description: "正常なケース",
		},
		{
			name:        "短縮形オプション",
			args:        []string{"ghp_test123", "-u", "testuser", "-e", "test@example.com"},
			shouldError: false,
			description: "短縮形オプションの使用",
		},
		{
			name:        "全オプション",
			args:        []string{"ghp_test123", "--user", "testuser", "--email", "test@example.com", "--target-dir", "/tmp", "--owner", "owner", "--organization", "org", "--collaborators", "user1:push", "--push-all", "--debug", "--public"},
			shouldError: false,
			description: "全オプションの指定",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 環境変数をクリア
			clearTestEnvs()

			config, err := ParseRewriteArgs(tt.args)

			if tt.shouldError && err == nil {
				t.Errorf("エラーが期待されましたが、エラーが発生しませんでした: %s", tt.description)
			}
			if !tt.shouldError && err != nil {
				t.Errorf("エラーが期待されませんでしたが、エラーが発生しました: %v (%s)", err, tt.description)
			}

			if !tt.shouldError && config != nil {
				// 基本的な設定値の確認
				if config.GitHubToken == "" {
					t.Error("GitHubTokenが設定されていません")
				}
				if config.GitHubUser == "" {
					t.Error("GitHubUserが設定されていません")
				}
				if config.GitHubEmail == "" {
					t.Error("GitHubEmailが設定されていません")
				}
			}
		})
	}
}

// TestParseRewriteArgsWithEnvironmentVariables は環境変数を使用したテストを行う
func TestParseRewriteArgsWithEnvironmentVariables(t *testing.T) {
	// 環境変数をクリーンアップ
	originalUser := os.Getenv("GITHUB_USER")
	originalEmail := os.Getenv("GITHUB_EMAIL")
	originalOwner := os.Getenv("GITHUB_REPOSITORY_OWNER")
	originalOrg := os.Getenv("GITHUB_ORGANIZATION")
	originalCollaborators := os.Getenv("GITHUB_COLLABORATORS")
	originalDebug := os.Getenv("GIT_REWRITE_DEBUG")

	defer func() {
		restoreEnv("GITHUB_USER", originalUser)
		restoreEnv("GITHUB_EMAIL", originalEmail)
		restoreEnv("GITHUB_REPOSITORY_OWNER", originalOwner)
		restoreEnv("GITHUB_ORGANIZATION", originalOrg)
		restoreEnv("GITHUB_COLLABORATORS", originalCollaborators)
		restoreEnv("GIT_REWRITE_DEBUG", originalDebug)
	}()

	// 環境変数を設定
	os.Setenv("GITHUB_USER", "envuser")
	os.Setenv("GITHUB_EMAIL", "env@example.com")
	os.Setenv("GITHUB_REPOSITORY_OWNER", "envowner")
	os.Setenv("GITHUB_ORGANIZATION", "envorg")
	os.Setenv("GITHUB_COLLABORATORS", "envuser1:push")
	os.Setenv("GIT_REWRITE_DEBUG", "1")

	config, err := ParseRewriteArgs([]string{"ghp_test123"})
	if err != nil {
		t.Errorf("環境変数を使用したテストでエラーが発生しました: %v", err)
		return
	}

	// 環境変数からの値が設定されているかチェック
	if config.GitHubUser != "envuser" {
		t.Errorf("環境変数からのGitHubUserが設定されていません。期待値: envuser, 実際: %s", config.GitHubUser)
	}
	if config.GitHubEmail != "env@example.com" {
		t.Errorf("環境変数からのGitHubEmailが設定されていません。期待値: env@example.com, 実際: %s", config.GitHubEmail)
	}
	if config.Owner != "envowner" {
		t.Errorf("環境変数からのOwnerが設定されていません。期待値: envowner, 実際: %s", config.Owner)
	}
	if config.Organization != "envorg" {
		t.Errorf("環境変数からのOrganizationが設定されていません。期待値: envorg, 実際: %s", config.Organization)
	}
	if config.Collaborators != "envuser1:push" {
		t.Errorf("環境変数からのCollaboratorsが設定されていません。期待値: envuser1:push, 実際: %s", config.Collaborators)
	}
	if !config.Debug {
		t.Error("環境変数からのDebugが設定されていません")
	}
}

// TestParseRewriteArgsArgumentPriority は引数の優先度をテストする
func TestParseRewriteArgsArgumentPriority(t *testing.T) {
	// 環境変数をクリーンアップ
	originalUser := os.Getenv("GITHUB_USER")
	originalEmail := os.Getenv("GITHUB_EMAIL")

	defer func() {
		restoreEnv("GITHUB_USER", originalUser)
		restoreEnv("GITHUB_EMAIL", originalEmail)
	}()

	// 環境変数を設定
	os.Setenv("GITHUB_USER", "envuser")
	os.Setenv("GITHUB_EMAIL", "env@example.com")

	// コマンド引数で上書き
	config, err := ParseRewriteArgs([]string{"ghp_test123", "--user", "arguser", "--email", "arg@example.com"})
	if err != nil {
		t.Errorf("引数優先度テストでエラーが発生しました: %v", err)
		return
	}

	// コマンド引数が優先されることを確認
	if config.GitHubUser != "arguser" {
		t.Errorf("コマンド引数が優先されていません。期待値: arguser, 実際: %s", config.GitHubUser)
	}
	if config.GitHubEmail != "arg@example.com" {
		t.Errorf("コマンド引数が優先されていません。期待値: arg@example.com, 実際: %s", config.GitHubEmail)
	}
}

// TestParseRewriteArgsDefaults はデフォルト値をテストする
func TestParseRewriteArgsDefaults(t *testing.T) {
	// 環境変数をクリーンアップ
	clearTestEnvs()

	config, err := ParseRewriteArgs([]string{"ghp_test123", "--user", "testuser", "--email", "test@example.com"})
	if err != nil {
		t.Errorf("デフォルト値テストでエラーが発生しました: %v", err)
		return
	}

	// デフォルト値の確認
	if config.TargetDir != "." {
		t.Errorf("TargetDirのデフォルト値が正しくありません。期待値: ., 実際: %s", config.TargetDir)
	}
	if !config.Private {
		t.Error("Privateのデフォルト値がtrueではありません")
	}
	if config.PushAll {
		t.Error("PushAllのデフォルト値がfalseではありません")
	}
	if config.Debug {
		t.Error("Debugのデフォルト値がfalseではありません")
	}
	if !config.DisableActions {
		t.Error("DisableActionsのデフォルト値がtrueではありません")
	}
}

// TestParseRewriteArgsActionsControl はActions制御オプションをテストする
func TestParseRewriteArgsActionsControl(t *testing.T) {
	// 環境変数をクリーンアップ
	clearTestEnvs()

	tests := []struct {
		name                   string
		args                   []string
		expectedDisableActions bool
		description            string
	}{
		{
			name:                   "デフォルト（Actions制御有効）",
			args:                   []string{"ghp_test123", "--user", "testuser", "--email", "test@example.com"},
			expectedDisableActions: true,
			description:            "デフォルトでActions制御が有効",
		},
		{
			name:                   "--enable-actionsでActions制御無効",
			args:                   []string{"ghp_test123", "--user", "testuser", "--email", "test@example.com", "--enable-actions"},
			expectedDisableActions: false,
			description:            "--enable-actionsでActions制御を無効化",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ParseRewriteArgs(tt.args)
			if err != nil {
				t.Errorf("Actions制御テストでエラーが発生しました: %v (%s)", err, tt.description)
				return
			}

			if config.DisableActions != tt.expectedDisableActions {
				t.Errorf("DisableActionsが期待値と異なります。期待値: %t, 実際: %t (%s)",
					tt.expectedDisableActions, config.DisableActions, tt.description)
			}
		})
	}
}

// TestGetConfigValue はgetConfigValue関数をテストする
func TestGetConfigValue(t *testing.T) {
	// 環境変数をクリーンアップ
	originalTestEnv := os.Getenv("TEST_ENV_VAR")
	defer restoreEnv("TEST_ENV_VAR", originalTestEnv)

	tests := []struct {
		name         string
		flagValue    string
		envKey       string
		envValue     string
		defaultValue string
		expected     string
	}{
		{
			name:         "フラグ値が優先",
			flagValue:    "flag_value",
			envKey:       "TEST_ENV_VAR",
			envValue:     "env_value",
			defaultValue: "default_value",
			expected:     "flag_value",
		},
		{
			name:         "環境変数が次に優先",
			flagValue:    "",
			envKey:       "TEST_ENV_VAR",
			envValue:     "env_value",
			defaultValue: "default_value",
			expected:     "env_value",
		},
		{
			name:         "デフォルト値が最後",
			flagValue:    "",
			envKey:       "TEST_ENV_VAR",
			envValue:     "",
			defaultValue: "default_value",
			expected:     "default_value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 環境変数を設定
			if tt.envValue != "" {
				os.Setenv(tt.envKey, tt.envValue)
			} else {
				os.Unsetenv(tt.envKey)
			}

			result := getConfigValue(tt.flagValue, tt.envKey, tt.defaultValue)

			if result != tt.expected {
				t.Errorf("期待される結果: %s, 実際: %s", tt.expected, result)
			}
		})
	}
}

// clearTestEnvs はテスト用の環境変数をクリアする
func clearTestEnvs() {
	os.Unsetenv("GITHUB_USER")
	os.Unsetenv("GITHUB_EMAIL")
	os.Unsetenv("GITHUB_REPOSITORY_OWNER")
	os.Unsetenv("GITHUB_ORGANIZATION")
	os.Unsetenv("GITHUB_COLLABORATORS")
	os.Unsetenv("GIT_REWRITE_DEBUG")
}

// restoreEnv は環境変数を復元する
func restoreEnv(key, value string) {
	if value != "" {
		os.Setenv(key, value)
	} else {
		os.Unsetenv(key)
	}
}
