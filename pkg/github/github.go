package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"git-rewrite-and-go/pkg/utils"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// Client はGitHub APIクライアント
type Client struct {
	Token      string
	HTTPClient *http.Client
}

// NewClient は新しいGitHub APIクライアントを作成する
func NewClient(token string) *Client {
	return &Client{
		Token: token,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// User はGitHubユーザー情報
type User struct {
	Login string `json:"login"`
}

// Repository はGitHubリポジトリ情報
type Repository struct {
	Name        string `json:"name"`
	Private     bool   `json:"private"`
	Description string `json:"description"`
	AutoInit    bool   `json:"auto_init"`
}

// Collaborator はコラボレーター情報
type Collaborator struct {
	Username   string `json:"username"`
	Permission string `json:"permission"` // "pull", "push", "admin", "maintain", "triage"
}

// CollaboratorConfig はコラボレーター設定
type CollaboratorConfig struct {
	DefaultCollaborators []Collaborator            `json:"default_collaborators"`
	ProjectCollaborators map[string][]Collaborator `json:"project_collaborators"`
}

// ActionsPermissions はGitHub Actionsの権限設定
type ActionsPermissions struct {
	Enabled bool `json:"enabled"`
}

// CheckRepoExists はリポジトリの存在を確認する
func (c *Client) CheckRepoExists(owner, repo string) (bool, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}

	if c.Token != "" {
		req.Header.Set("Authorization", "token "+c.Token)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		return true, nil
	} else if resp.StatusCode == 404 {
		return false, nil
	}

	return false, fmt.Errorf("GitHub API エラー: %d - %s", resp.StatusCode, resp.Status)
}

// GetCurrentUser は現在のユーザー情報を取得する
func (c *Client) GetCurrentUser() (*User, error) {
	if c.Token == "" {
		return nil, fmt.Errorf("GitHub トークンが設定されていません")
	}

	url := "https://api.github.com/user"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "token "+c.Token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("ユーザー情報取得エラー: %d - %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var user User
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

// IsOrganization は指定されたオーナーが組織かどうかを判定する
func (c *Client) IsOrganization(owner string) (bool, error) {
	if c.Token == "" {
		return false, fmt.Errorf("GitHub トークンが設定されていません")
	}

	// まず組織として確認
	url := fmt.Sprintf("https://api.github.com/orgs/%s", owner)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}

	req.Header.Set("Authorization", "token "+c.Token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		return true, nil // 組織として存在
	} else if resp.StatusCode == 404 {
		return false, nil // 組織として存在しない（個人ユーザーの可能性）
	}

	return false, fmt.Errorf("組織確認エラー: %d - %s", resp.StatusCode, resp.Status)
}

// CreateRepo はリポジトリを作成する
func (c *Client) CreateRepo(owner, repo string, private bool) error {
	if c.Token == "" {
		return fmt.Errorf("GitHub トークンが設定されていません")
	}

	// 現在のユーザー情報を取得
	currentUser, err := c.GetCurrentUser()
	if err != nil {
		return err
	}

	// リポジトリ作成のURL決定
	var url string

	// 個人リポジトリ所有者が設定されている場合は個人リポジトリとして扱う
	if utils.IsPersonalRepository(os.Getenv("GITHUB_REPOSITORY_OWNER")) {
		url = "https://api.github.com/user/repos"
		fmt.Printf("個人リポジトリとして作成します（個人リポジトリ所有者指定）: %s\n", owner)
	} else if owner == currentUser.Login {
		// 現在のユーザーと同じ場合は個人リポジトリ
		url = "https://api.github.com/user/repos"
		fmt.Printf("個人リポジトリとして作成します: %s\n", owner)
	} else {
		// 異なる場合は組織かどうかを確認
		isOrg, err := c.IsOrganization(owner)
		if err != nil {
			fmt.Printf("⚠️  組織確認でエラーが発生しました: %v\n", err)
			fmt.Printf("組織として試行します: %s\n", owner)
			// エラーが発生した場合は組織として試行（フォールバック）
			isOrg = true
		}

		if isOrg {
			// 組織リポジトリ
			url = fmt.Sprintf("https://api.github.com/orgs/%s/repos", owner)
			fmt.Printf("組織リポジトリとして作成します: %s\n", owner)
		} else {
			// 個人ユーザーのリポジトリ（他のユーザー）
			// この場合、現在のユーザーが他のユーザーのリポジトリを作成することはできない
			return fmt.Errorf("他のユーザー '%s' のリポジトリを作成することはできません。組織名を確認してください", owner)
		}
	}

	// リポジトリ作成データ
	repoData := Repository{
		Name:        repo,
		Private:     private,
		Description: fmt.Sprintf("Repository created automatically for %s", repo),
		AutoInit:    false,
	}

	jsonData, err := json.Marshal(repoData)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "token "+c.Token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 201 {
		fmt.Printf("✅ GitHubリポジトリ %s/%s を作成しました。\n", owner, repo)
		return nil
	}

	body, _ := io.ReadAll(resp.Body)

	// 422エラーでリポジトリ名が既に存在する場合の特別処理
	if resp.StatusCode == 422 {
		var errorResponse map[string]interface{}
		if err := json.Unmarshal(body, &errorResponse); err == nil {
			if errors, ok := errorResponse["errors"].([]interface{}); ok {
				for _, errorItem := range errors {
					if errorMap, ok := errorItem.(map[string]interface{}); ok {
						if field, ok := errorMap["field"].(string); ok && field == "name" {
							if code, ok := errorMap["code"].(string); ok && code == "custom" {
								fmt.Printf("⚠️  リポジトリ %s/%s は既に存在します。既存のリポジトリを使用します。\n", owner, repo)
								return nil // エラーではなく正常終了として扱う
							}
						}
					}
				}
			}
		}
	}

	return fmt.Errorf("リポジトリ作成エラー: %d - %s\n詳細: %s", resp.StatusCode, resp.Status, string(body))
}

// AddCollaborator はリポジトリにコラボレーターを追加する
func (c *Client) AddCollaborator(owner, repo, username, permission string) error {
	if c.Token == "" {
		return fmt.Errorf("GitHub トークンが設定されていません")
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/collaborators/%s", owner, repo, username)

	data := map[string]string{
		"permission": permission,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "token "+c.Token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 201 || resp.StatusCode == 204 {
		fmt.Printf("✅ コラボレーター %s を %s/%s に追加しました（権限: %s）\n", username, owner, repo, permission)
		return nil
	}

	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("コラボレーター追加エラー: %d - %s\n詳細: %s", resp.StatusCode, resp.Status, string(body))
}

// ParseCollaboratorsFromEnv は環境変数からコラボレーター情報を解析する
func ParseCollaboratorsFromEnv() []Collaborator {
	var collaborators []Collaborator

	// GITHUB_COLLABORATORS="user1:push,user2:admin,user3:pull"
	envValue := os.Getenv("GITHUB_COLLABORATORS")
	if envValue == "" {
		return collaborators
	}

	pairs := strings.Split(envValue, ",")
	for _, pair := range pairs {
		parts := strings.Split(strings.TrimSpace(pair), ":")
		if len(parts) == 2 {
			username := strings.TrimSpace(parts[0])
			permission := strings.TrimSpace(parts[1])

			// 権限の妥当性チェック
			if isValidPermission(permission) {
				collaborators = append(collaborators, Collaborator{
					Username:   username,
					Permission: permission,
				})
			} else {
				fmt.Printf("⚠️  無効な権限が指定されました: %s (ユーザー: %s)\n", permission, username)
			}
		}
	}

	return collaborators
}

// LoadCollaboratorConfig は設定ファイルからコラボレーター情報を読み込む
func LoadCollaboratorConfig(configPath string) (*CollaboratorConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &CollaboratorConfig{}, nil // 設定ファイルがない場合は空の設定
		}
		return nil, err
	}

	var config CollaboratorConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// 設定ファイル内の権限の妥当性チェック
	config.DefaultCollaborators = validateCollaborators(config.DefaultCollaborators)

	for projectName, collaborators := range config.ProjectCollaborators {
		config.ProjectCollaborators[projectName] = validateCollaborators(collaborators)
	}

	return &config, nil
}

// GetCollaborators は複数のソースからコラボレーター情報を取得する
func (c *Client) GetCollaborators(configPath, repoName string) []Collaborator {
	var collaborators []Collaborator

	// 1. 設定ファイルから読み込み（優先度: 高）
	if configPath != "" {
		if config, err := LoadCollaboratorConfig(configPath); err == nil {
			// デフォルトコラボレーターを追加
			collaborators = append(collaborators, config.DefaultCollaborators...)
			if len(config.DefaultCollaborators) > 0 {
				fmt.Printf("✅ 設定ファイルから %d 人のデフォルトコラボレーターを読み込みました\n", len(config.DefaultCollaborators))
			}

			// プロジェクト固有のコラボレーターを追加
			if projectCollaborators, exists := config.ProjectCollaborators[repoName]; exists {
				collaborators = append(collaborators, projectCollaborators...)
				fmt.Printf("✅ 設定ファイルから %d 人のプロジェクト固有コラボレーターを読み込みました\n", len(projectCollaborators))
			}
		} else {
			fmt.Printf("⚠️  設定ファイル読み込みエラー: %v\n", err)
		}
	}

	// 2. 環境変数から読み込み（優先度: 中）
	envCollaborators := ParseCollaboratorsFromEnv()
	if len(envCollaborators) > 0 {
		collaborators = append(collaborators, envCollaborators...)
		fmt.Printf("✅ 環境変数から %d 人のコラボレーターを読み込みました\n", len(envCollaborators))
	}

	// 重複を除去（後から追加されたものが優先される）
	return removeDuplicateCollaborators(collaborators)
}

// CreateRepoWithCollaborators はリポジトリを作成してコラボレーターを追加する
func (c *Client) CreateRepoWithCollaborators(owner, repo string, private bool, configPath string) error {
	// リポジトリを作成
	if err := c.CreateRepo(owner, repo, private); err != nil {
		return err
	}

	// コラボレーターを取得・追加
	collaborators := c.GetCollaborators(configPath, repo)

	if len(collaborators) == 0 {
		fmt.Println("ℹ️  コラボレーターの設定が見つかりませんでした")
		return nil
	}

	fmt.Printf("📝 %d 人のコラボレーターを追加しています...\n", len(collaborators))

	successCount := 0
	for _, collaborator := range collaborators {
		if err := c.AddCollaborator(owner, repo, collaborator.Username, collaborator.Permission); err != nil {
			fmt.Printf("⚠️  コラボレーター %s の追加に失敗しました: %v\n", collaborator.Username, err)
		} else {
			successCount++
		}
	}

	fmt.Printf("✅ %d/%d 人のコラボレーターを正常に追加しました\n", successCount, len(collaborators))
	return nil
}

// isValidPermission は権限の妥当性をチェックする
func isValidPermission(permission string) bool {
	validPermissions := []string{"pull", "push", "admin", "maintain", "triage"}
	for _, valid := range validPermissions {
		if permission == valid {
			return true
		}
	}
	return false
}

// validateCollaborators はコラボレーターリストの権限を検証する
func validateCollaborators(collaborators []Collaborator) []Collaborator {
	var validCollaborators []Collaborator
	for _, collaborator := range collaborators {
		if isValidPermission(collaborator.Permission) {
			validCollaborators = append(validCollaborators, collaborator)
		} else {
			fmt.Printf("⚠️  無効な権限が指定されました: %s (ユーザー: %s)\n", collaborator.Permission, collaborator.Username)
		}
	}
	return validCollaborators
}

// removeDuplicateCollaborators は重複するコラボレーターを除去する
func removeDuplicateCollaborators(collaborators []Collaborator) []Collaborator {
	seen := make(map[string]Collaborator)

	// 後から追加されたものが優先される（環境変数 > 設定ファイル）
	for _, collaborator := range collaborators {
		seen[collaborator.Username] = collaborator
	}

	var result []Collaborator
	for _, collaborator := range seen {
		result = append(result, collaborator)
	}

	return result
}

// SetActionsEnabled はリポジトリのGitHub Actionsの有効/無効を設定する
func (c *Client) SetActionsEnabled(owner, repo string, enabled bool) error {
	if c.Token == "" {
		return fmt.Errorf("GitHub トークンが設定されていません")
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/actions/permissions", owner, repo)

	permissions := ActionsPermissions{
		Enabled: enabled,
	}

	jsonData, err := json.Marshal(permissions)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "token "+c.Token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 204 {
		status := map[bool]string{true: "有効", false: "無効"}[enabled]
		fmt.Printf("✅ リポジトリ %s/%s のGitHub Actionsを%sにしました\n", owner, repo, status)
		return nil
	}

	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("GitHub Actions設定エラー: %d - %s\n詳細: %s", resp.StatusCode, resp.Status, string(body))
}

// GetActionsEnabled はリポジトリのGitHub Actionsの有効/無効状態を取得する
func (c *Client) GetActionsEnabled(owner, repo string) (bool, error) {
	if c.Token == "" {
		return false, fmt.Errorf("GitHub トークンが設定されていません")
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/actions/permissions", owner, repo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}

	req.Header.Set("Authorization", "token "+c.Token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return false, err
		}

		var permissions ActionsPermissions
		if err := json.Unmarshal(body, &permissions); err != nil {
			return false, err
		}

		return permissions.Enabled, nil
	}

	body, _ := io.ReadAll(resp.Body)
	return false, fmt.Errorf("GitHub Actions状態取得エラー: %d - %s\n詳細: %s", resp.StatusCode, resp.Status, string(body))
}
