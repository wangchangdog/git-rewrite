package rewriter

import (
	"fmt"
	"os"
	"strings"

	"git-rewrite-and-go/pkg/git"
	"git-rewrite-and-go/pkg/github"
	"git-rewrite-and-go/pkg/utils"
)

// RewriteResult は書き換え結果を表す
type RewriteResult struct {
	Success          bool
	HistoryRewritten bool
	PushSucceeded    bool
	Error            error
	GitDir           string
}

// Rewriter はGit履歴書き換えを行う
type Rewriter struct {
	GitHubClient           *github.Client
	GitHubToken            string
	GitHubUser             string
	GitHubEmail            string
	CollaboratorConfigPath string
	PushAll                bool
	Owner                  string
	Organization           string
	Private                bool
	CollaboratorsString    string
}

// NewRewriter は新しいRewriterを作成する
func NewRewriter(githubToken, githubUser, githubEmail string) *Rewriter {
	return &Rewriter{
		GitHubClient:           github.NewClient(githubToken),
		GitHubToken:            githubToken,
		GitHubUser:             githubUser,
		GitHubEmail:            githubEmail,
		CollaboratorConfigPath: "", // デフォルトは空（環境変数のみ使用）
		PushAll:                false,
		Private:                true, // デフォルトはプライベート
	}
}

// NewRewriterWithConfig はコラボレーター設定ファイル付きでRewriterを作成する
func NewRewriterWithConfig(githubToken, githubUser, githubEmail, configPath string) *Rewriter {
	return &Rewriter{
		GitHubClient:           github.NewClient(githubToken),
		GitHubToken:            githubToken,
		GitHubUser:             githubUser,
		GitHubEmail:            githubEmail,
		CollaboratorConfigPath: configPath,
		PushAll:                false,
		Private:                true, // デフォルトはプライベート
	}
}

// SetPushAllOption はプッシュオプションを設定する
func (r *Rewriter) SetPushAllOption(pushAll bool) {
	r.PushAll = pushAll
}

// SetOwnershipConfig は所有者設定を行う
func (r *Rewriter) SetOwnershipConfig(owner, organization string) {
	r.Owner = owner
	r.Organization = organization
}

// SetPrivateOption はプライベートリポジトリ設定を行う
func (r *Rewriter) SetPrivateOption(private bool) {
	r.Private = private
}

// SetCollaboratorsFromString は文字列からコラボレーター設定を行う
func (r *Rewriter) SetCollaboratorsFromString(collaborators string) {
	r.CollaboratorsString = collaborators
}

// RewriteGitHistory はGit履歴を書き換える
func (r *Rewriter) RewriteGitHistory(gitDir string) error {
	return git.RewriteHistory(gitDir, r.GitHubUser, r.GitHubEmail)
}

// UpdateRemoteURL はリモートURLを更新する
func (r *Rewriter) UpdateRemoteURL(gitDir string) error {
	return git.UpdateRemoteURL(gitDir, r.GitHubUser, r.Owner, r.Organization)
}

// CreateInitialCommit は初期コミットを作成する
func (r *Rewriter) CreateInitialCommit(gitDir string) error {
	return git.CreateInitialCommit(gitDir, r.GitHubUser, r.GitHubEmail)
}

// VerifyAndPushRemote はリモートリポジトリの確認とプッシュを行う
func (r *Rewriter) VerifyAndPushRemote(gitDir string) error {
	fmt.Println("\n--- リモートリポジトリの確認とプッシュ ---")

	// リモートURLを取得
	stdout, _, err := utils.RunCommand(gitDir, "git", "remote", "get-url", "origin")
	if err != nil {
		return fmt.Errorf("リモートURL取得エラー: %v", err)
	}

	remoteURL := strings.TrimSpace(stdout)
	fmt.Printf("現在のリモートURL: %s\n", remoteURL)

	// リモートURLからユーザー名とリポジトリ名を抽出
	owner, repoName := utils.ExtractRepoInfoFromURL(remoteURL)
	if os.Getenv("GIT_REWRITE_DEBUG") != "" {
		fmt.Printf("デバッグ: VerifyAndPushRemote URL解析結果 - URL: %s, Owner: '%s', Repo: '%s'\n", remoteURL, owner, repoName)
	}
	if owner == "" || repoName == "" {
		return fmt.Errorf("リモートURLからリポジトリ情報を抽出できませんでした: %s (解析結果: owner='%s', repo='%s')", remoteURL, owner, repoName)
	}

	fmt.Printf("リポジトリ情報: %s/%s\n", owner, repoName)

	// 期待されるオーナーを決定
	expectedOwner := utils.GetTargetOwner(r.GitHubUser, r.Owner, r.Organization)

	// リモートURLに期待されるオーナーが含まれているか確認
	if !strings.Contains(remoteURL, expectedOwner) {
		fmt.Printf("⚠️  警告: リモートリポジトリが %s に設定されていません。\n", expectedOwner)
		fmt.Printf("   期待されるオーナー: %s\n", expectedOwner)
		fmt.Printf("   実際のURL: %s\n", remoteURL)
		return fmt.Errorf("リモートリポジトリのオーナーが一致しません")
	}

	fmt.Printf("✅ リモートリポジトリが %s に設定されています。\n", expectedOwner)

	// GitHubリポジトリの存在確認
	fmt.Println("GitHubリポジトリの存在を確認しています...")
	exists, err := r.GitHubClient.CheckRepoExists(owner, repoName)
	if err != nil {
		return fmt.Errorf("リポジトリ存在確認エラー: %v", err)
	}

	if !exists {
		fmt.Printf("⚠️  リモートリポジトリ %s/%s が存在しません。\n", owner, repoName)
		fmt.Println("リポジトリを作成しています...")

		// コラボレーター設定を決定
		collaboratorConfig := r.CollaboratorConfigPath
		if collaboratorConfig == "" && r.CollaboratorsString != "" {
			// 文字列からコラボレーター設定を一時的に環境変数に設定
			os.Setenv("GITHUB_COLLABORATORS", r.CollaboratorsString)
			defer os.Unsetenv("GITHUB_COLLABORATORS")
		}

		if err := r.GitHubClient.CreateRepoWithCollaborators(owner, repoName, r.Private, collaboratorConfig); err != nil {
			return fmt.Errorf("リポジトリ作成エラー: %v", err)
		}
	} else {
		fmt.Printf("✅ リモートリポジトリ %s/%s が存在します。\n", owner, repoName)
	}

	// コミット履歴の確認と初期コミット作成
	fmt.Println("コミット履歴を確認しています...")
	_, _, err = utils.RunCommand(gitDir, "git", "log", "--oneline", "-1")
	if err != nil {
		// コミットが存在しない場合
		fmt.Println("⚠️  コミットが存在しません。初期コミットを作成します。")
		if err := r.CreateInitialCommit(gitDir); err != nil {
			return fmt.Errorf("初期コミット作成エラー: %v", err)
		}
	} else {
		fmt.Println("✅ 既存のコミットが見つかりました。")
	}

	// リモートにプッシュ
	if err := git.PushToRemote(gitDir, r.GitHubToken); err != nil {
		return err
	}

	// --push-all オプションが有効な場合、全ブランチとタグをプッシュ
	if r.PushAll {
		if err := r.PushAllBranchesAndTags(gitDir); err != nil {
			return fmt.Errorf("全ブランチ・タグのプッシュエラー: %v", err)
		}
	}

	return nil
}

// PushAllBranchesAndTags はローカルの全ブランチとタグをリモートにプッシュする
func (r *Rewriter) PushAllBranchesAndTags(gitDir string) error {
	return git.PushAllBranchesAndTags(gitDir, r.GitHubToken)
}

// ProcessRepository は単一のリポジトリを処理する
func (r *Rewriter) ProcessRepository(gitDir string) *RewriteResult {
	result := &RewriteResult{
		GitDir: gitDir,
	}

	// Git履歴の書き換え
	if err := r.RewriteGitHistory(gitDir); err != nil {
		result.Error = err
		return result
	}
	result.HistoryRewritten = true

	// リモートURL更新
	if err := r.UpdateRemoteURL(gitDir); err != nil {
		result.Error = err
		return result
	}

	// リモート確認とプッシュ
	if err := r.VerifyAndPushRemote(gitDir); err != nil {
		result.Error = err
		return result
	}

	result.Success = true
	result.PushSucceeded = true
	return result
}
