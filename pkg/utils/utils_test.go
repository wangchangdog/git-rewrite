package utils

import (
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
			// 環境変数を設定
			if tt.repoOwner != "" {
				t.Setenv("GITHUB_REPOSITORY_OWNER", tt.repoOwner)
			}
			if tt.orgValue != "" {
				t.Setenv("GITHUB_ORGANIZATION", tt.orgValue)
			}

			result := GetTargetOwner(tt.defaultUser, tt.repoOwner, tt.orgValue)

			if result != tt.expected {
				t.Errorf("期待される結果: %s, 実際: %s", tt.expected, result)
			}
		})
	}
}

func TestIsPersonalRepository(t *testing.T) {
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
			if tt.repoOwner != "" {
				t.Setenv("GITHUB_REPOSITORY_OWNER", tt.repoOwner)
			}

			result := IsPersonalRepository(tt.repoOwner)

			if result != tt.expected {
				t.Errorf("期待される結果: %t, 実際: %t", tt.expected, result)
			}
		})
	}
}
