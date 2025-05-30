package git

import (
	"fmt"
	"os"
	"strings"

	"git-rewrite/pkg/utils"
)

// UpdateRemoteURL はリモートURLを更新する
func UpdateRemoteURL(gitDir, githubUser, owner, organization string) error {
	targetOwner := utils.GetTargetOwner(githubUser, owner, organization)
	fmt.Printf("[2/2] Git remoteのorganization部分を%sに変更します...\n", targetOwner)

	// remote originが存在するかチェック
	stdout, _, err := utils.RunCommand(gitDir, "git", "remote", "get-url", "origin")
	if err != nil {
		fmt.Println("警告: remote originが設定されていません。スキップします。")
		return nil
	}

	remoteURL := strings.TrimSpace(stdout)
	fmt.Printf("現在のremote URL: %s\n", remoteURL)

	// URLを解析して新しいURLを生成
	newURL, err := generateNewRemoteURL(remoteURL, githubUser, owner, organization)
	if err != nil {
		fmt.Printf("警告: %v\n", err)
		fmt.Println("remote URLの変更をスキップします。")
		return nil
	}

	// リモートURLを更新
	_, _, err = utils.RunCommand(gitDir, "git", "remote", "set-url", "origin", newURL)
	if err != nil {
		return fmt.Errorf("remote URL更新エラー: %v", err)
	}

	fmt.Printf("remote URLを%sに変更しました。\n", newURL)
	return nil
}

// generateNewRemoteURL は新しいリモートURLを生成する
func generateNewRemoteURL(remoteURL, githubUser, owner, organization string) (string, error) {
	ownerFromURL, repo := utils.ExtractRepoInfoFromURL(remoteURL)
	if os.Getenv("GIT_REWRITE_DEBUG") != "" {
		fmt.Printf("デバッグ: URL解析結果 - URL: %s, Owner: '%s', Repo: '%s'\n", remoteURL, ownerFromURL, repo)
	}
	if ownerFromURL == "" || repo == "" {
		return "", fmt.Errorf("remote URLが想定外の形式です: %s (解析結果: owner='%s', repo='%s')", remoteURL, ownerFromURL, repo)
	}

	targetOwner := utils.GetTargetOwner(githubUser, owner, organization)
	if targetOwner != githubUser {
		if utils.IsPersonalRepository(owner) {
			fmt.Printf("個人リポジトリ所有者が設定されています: %s\n", targetOwner)
		} else {
			fmt.Printf("組織が設定されています: %s\n", targetOwner)
		}
	}

	// HTTPS形式かSSH形式かを判定
	if strings.HasPrefix(remoteURL, "https://") {
		return fmt.Sprintf("https://github.com/%s/%s", targetOwner, repo), nil
	} else if strings.HasPrefix(remoteURL, "git@") {
		return fmt.Sprintf("git@github.com:%s/%s.git", targetOwner, repo), nil
	}

	return "", fmt.Errorf("サポートされていないURL形式: %s", remoteURL)
}
