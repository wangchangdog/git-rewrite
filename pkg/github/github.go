package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	if owner == currentUser.Login {
		// 個人リポジトリ
		url = "https://api.github.com/user/repos"
	} else {
		// 組織リポジトリ
		url = fmt.Sprintf("https://api.github.com/orgs/%s/repos", owner)
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
		fmt.Printf("✓ GitHubリポジトリ %s/%s を作成しました。\n", owner, repo)
		return nil
	}

	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("リポジトリ作成エラー: %d - %s\n詳細: %s", resp.StatusCode, resp.Status, string(body))
}
