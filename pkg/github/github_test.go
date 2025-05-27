package github

import (
	"os"
	"testing"
)

func TestParseCollaboratorsFromEnv(t *testing.T) {
	tests := []struct {
		name        string
		envValue    string
		expected    int
		usernames   []string
		permissions []string
	}{
		{
			name:     "空の環境変数",
			envValue: "",
			expected: 0,
		},
		{
			name:        "単一のコラボレーター",
			envValue:    "user1:push",
			expected:    1,
			usernames:   []string{"user1"},
			permissions: []string{"push"},
		},
		{
			name:        "複数のコラボレーター",
			envValue:    "user1:push,user2:admin,user3:pull",
			expected:    3,
			usernames:   []string{"user1", "user2", "user3"},
			permissions: []string{"push", "admin", "pull"},
		},
		{
			name:        "無効な権限を含む",
			envValue:    "user1:push,user2:invalid,user3:admin",
			expected:    2, // 無効な権限は除外される
			usernames:   []string{"user1", "user3"},
			permissions: []string{"push", "admin"},
		},
		{
			name:        "スペースを含む",
			envValue:    " user1 : push , user2 : admin ",
			expected:    2,
			usernames:   []string{"user1", "user2"},
			permissions: []string{"push", "admin"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 環境変数を設定
			os.Setenv("GITHUB_COLLABORATORS", tt.envValue)
			defer os.Unsetenv("GITHUB_COLLABORATORS")

			collaborators := ParseCollaboratorsFromEnv()

			if len(collaborators) != tt.expected {
				t.Errorf("期待されるコラボレーター数: %d, 実際: %d", tt.expected, len(collaborators))
			}

			for i, collaborator := range collaborators {
				if i < len(tt.usernames) && collaborator.Username != tt.usernames[i] {
					t.Errorf("期待されるユーザー名: %s, 実際: %s", tt.usernames[i], collaborator.Username)
				}
				if i < len(tt.permissions) && collaborator.Permission != tt.permissions[i] {
					t.Errorf("期待される権限: %s, 実際: %s", tt.permissions[i], collaborator.Permission)
				}
			}
		})
	}
}

func TestIsValidPermission(t *testing.T) {
	validPermissions := []string{"pull", "push", "admin", "maintain", "triage"}
	invalidPermissions := []string{"invalid", "read", "write", "owner", ""}

	for _, permission := range validPermissions {
		if !isValidPermission(permission) {
			t.Errorf("権限 %s は有効であるべきです", permission)
		}
	}

	for _, permission := range invalidPermissions {
		if isValidPermission(permission) {
			t.Errorf("権限 %s は無効であるべきです", permission)
		}
	}
}

func TestValidateCollaborators(t *testing.T) {
	input := []Collaborator{
		{Username: "user1", Permission: "push"},
		{Username: "user2", Permission: "invalid"},
		{Username: "user3", Permission: "admin"},
		{Username: "user4", Permission: ""},
	}

	result := validateCollaborators(input)

	expected := 2 // user1とuser3のみ有効
	if len(result) != expected {
		t.Errorf("期待される有効なコラボレーター数: %d, 実際: %d", expected, len(result))
	}

	if result[0].Username != "user1" || result[0].Permission != "push" {
		t.Errorf("最初の有効なコラボレーターが正しくありません")
	}

	if result[1].Username != "user3" || result[1].Permission != "admin" {
		t.Errorf("2番目の有効なコラボレーターが正しくありません")
	}
}

func TestRemoveDuplicateCollaborators(t *testing.T) {
	input := []Collaborator{
		{Username: "user1", Permission: "push"},
		{Username: "user2", Permission: "admin"},
		{Username: "user1", Permission: "admin"}, // 重複（後の方が優先される）
		{Username: "user3", Permission: "pull"},
	}

	result := removeDuplicateCollaborators(input)

	expected := 3 // user1, user2, user3
	if len(result) != expected {
		t.Errorf("期待される重複除去後のコラボレーター数: %d, 実際: %d", expected, len(result))
	}

	// user1の権限が後から設定されたadminになっているかチェック
	user1Found := false
	for _, collaborator := range result {
		if collaborator.Username == "user1" {
			user1Found = true
			if collaborator.Permission != "admin" {
				t.Errorf("user1の権限は admin であるべきです（後から設定された値）, 実際: %s", collaborator.Permission)
			}
		}
	}

	if !user1Found {
		t.Errorf("user1が結果に含まれていません")
	}
}

func TestLoadCollaboratorConfig(t *testing.T) {
	// 存在しないファイルのテスト
	config, err := LoadCollaboratorConfig("nonexistent.json")
	if err != nil {
		t.Errorf("存在しないファイルでエラーが発生しました: %v", err)
	}
	if config == nil {
		t.Errorf("存在しないファイルでも空の設定が返されるべきです")
	}

	// 実際のファイルが存在する場合のテスト（collaborators.jsonが存在する場合）
	if _, err := os.Stat("../../collaborators.json"); err == nil {
		config, err := LoadCollaboratorConfig("../../collaborators.json")
		if err != nil {
			t.Errorf("設定ファイル読み込みエラー: %v", err)
		}
		if config == nil {
			t.Errorf("設定ファイルから設定が読み込まれませんでした")
		}
	}
}

// TestIsOrganizationLogic は組織判定ロジックのテスト（モック）
func TestIsOrganizationLogic(t *testing.T) {
	// 実際のAPIを呼ばないモックテスト
	// 実際のテストでは、HTTPクライアントをモックする必要があります

	// ここでは基本的なロジックのテストのみ
	client := NewClient("")

	// トークンが空の場合のテスト
	_, err := client.IsOrganization("test-org")
	if err == nil {
		t.Errorf("トークンが空の場合はエラーが発生するべきです")
	}

	expectedError := "GitHub トークンが設定されていません"
	if err.Error() != expectedError {
		t.Errorf("期待されるエラーメッセージ: %s, 実際: %s", expectedError, err.Error())
	}
}

// TestCreateRepoErrorHandling はリポジトリ作成時のエラーハンドリングのテスト
func TestCreateRepoErrorHandling(t *testing.T) {
	// 実際のAPIを呼ばないモックテスト
	client := NewClient("")

	// トークンが空の場合のテスト
	err := client.CreateRepo("test-owner", "test-repo", true)
	if err == nil {
		t.Errorf("トークンが空の場合はエラーが発生するべきです")
	}

	expectedError := "GitHub トークンが設定されていません"
	if err.Error() != expectedError {
		t.Errorf("期待されるエラーメッセージ: %s, 実際: %s", expectedError, err.Error())
	}
}

// TestCreateRepoWithCollaboratorsErrorHandling はコラボレーター付きリポジトリ作成時のエラーハンドリングのテスト
func TestCreateRepoWithCollaboratorsErrorHandling(t *testing.T) {
	// 実際のAPIを呼ばないモックテスト
	client := NewClient("")

	// トークンが空の場合のテスト
	err := client.CreateRepoWithCollaborators("test-owner", "test-repo", true, "")
	if err == nil {
		t.Errorf("トークンが空の場合はエラーが発生するべきです")
	}

	expectedError := "GitHub トークンが設定されていません"
	if err.Error() != expectedError {
		t.Errorf("期待されるエラーメッセージ: %s, 実際: %s", expectedError, err.Error())
	}
}
