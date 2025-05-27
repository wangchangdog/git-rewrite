package rewriter

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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
	fmt.Printf("[1/2] Git履歴のauthor/emailを書き換えます...\n")

	// 現在のディレクトリがGitリポジトリかチェック
	gitPath := filepath.Join(gitDir, ".git")
	if !utils.FileExists(gitPath) {
		return fmt.Errorf("エラー: %s はGitリポジトリではありません", gitDir)
	}

	// 既存のバックアップが存在する場合は削除
	backupPath := filepath.Join(gitDir, ".git", "refs", "original")
	if utils.FileExists(backupPath) {
		fmt.Println("既存のバックアップを削除しています...")
		if err := os.RemoveAll(backupPath); err != nil {
			return fmt.Errorf("バックアップ削除エラー: %v", err)
		}
	}

	// 環境変数を設定
	env := os.Environ()
	env = append(env, "LC_ALL=C.UTF-8")
	env = append(env, "LANG=C.UTF-8")
	env = append(env, "FILTER_BRANCH_SQUELCH_WARNING=1")

	// git filter-branchコマンドを構築
	envFilter := fmt.Sprintf(`
if [ "$GIT_COMMITTER_EMAIL" != "%s" ] || [ "$GIT_COMMITTER_NAME" != "%s" ]; then
    export GIT_COMMITTER_NAME="%s"
    export GIT_COMMITTER_EMAIL="%s"
fi
if [ "$GIT_AUTHOR_EMAIL" != "%s" ] || [ "$GIT_AUTHOR_NAME" != "%s" ]; then
    export GIT_AUTHOR_NAME="%s"
    export GIT_AUTHOR_EMAIL="%s"
fi
`, r.GitHubEmail, r.GitHubUser, r.GitHubUser, r.GitHubEmail,
		r.GitHubEmail, r.GitHubUser, r.GitHubUser, r.GitHubEmail)

	cmd := exec.Command("git", "filter-branch", "-f", "--env-filter", envFilter,
		"--tag-name-filter", "cat", "--", "--branches", "--tags")
	cmd.Dir = gitDir
	cmd.Env = env

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git filter-branchの実行に失敗しました: %v\n出力: %s", err, utils.SafeDecode(output))
	}

	fmt.Printf("✅ Git履歴の書き換えが完了しました。\n")
	return nil
}

// UpdateRemoteURL はリモートURLを更新する
func (r *Rewriter) UpdateRemoteURL(gitDir string) error {
	targetOwner := utils.GetTargetOwner(r.GitHubUser, r.Owner, r.Organization)
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
	newURL, err := r.generateNewRemoteURL(remoteURL)
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
func (r *Rewriter) generateNewRemoteURL(remoteURL string) (string, error) {
	owner, repo := utils.ExtractRepoInfoFromURL(remoteURL)
	if os.Getenv("GIT_REWRITE_DEBUG") != "" {
		fmt.Printf("デバッグ: URL解析結果 - URL: %s, Owner: '%s', Repo: '%s'\n", remoteURL, owner, repo)
	}
	if owner == "" || repo == "" {
		return "", fmt.Errorf("remote URLが想定外の形式です: %s (解析結果: owner='%s', repo='%s')", remoteURL, owner, repo)
	}

	targetOwner := utils.GetTargetOwner(r.GitHubUser, r.Owner, r.Organization)
	if targetOwner != r.GitHubUser {
		if utils.IsPersonalRepository(r.Owner) {
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

// CreateInitialCommit は初期コミットを作成する
func (r *Rewriter) CreateInitialCommit(gitDir string) error {
	// README.mdファイルが存在するかチェック
	readmePath := filepath.Join(gitDir, "README.md")
	if !utils.FileExists(readmePath) {
		// README.mdを作成
		repoName := filepath.Base(gitDir)
		content := fmt.Sprintf("# %s\n\nこのリポジトリは自動的に作成されました。\n", repoName)
		if err := os.WriteFile(readmePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("README.md作成エラー: %v", err)
		}
		fmt.Println("✅ README.mdファイルを作成しました。")
	}

	// ステージングエリアに追加
	_, _, err := utils.RunCommand(gitDir, "git", "add", ".")
	if err != nil {
		return fmt.Errorf("git add エラー: %v", err)
	}

	// 変更があるかチェック
	_, _, err = utils.RunCommand(gitDir, "git", "diff", "--cached", "--quiet")
	if err == nil {
		// 変更がない場合は空のコミットを作成
		fmt.Println("⚠️  ステージングエリアに変更がありません。空のコミットを作成します。")
		_, _, err = utils.RunCommand(gitDir, "git", "-c", fmt.Sprintf("user.name=%s", r.GitHubUser),
			"-c", fmt.Sprintf("user.email=%s", r.GitHubEmail),
			"commit", "--allow-empty", "-m", "Initial commit")
	} else {
		// 変更がある場合は通常のコミット
		_, _, err = utils.RunCommand(gitDir, "git", "-c", fmt.Sprintf("user.name=%s", r.GitHubUser),
			"-c", fmt.Sprintf("user.email=%s", r.GitHubEmail),
			"commit", "-m", "Initial commit")
	}

	if err != nil {
		return fmt.Errorf("初期コミット作成エラー: %v", err)
	}

	fmt.Println("✅ 初期コミットを作成しました。")
	return nil
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

	// 現在のブランチを取得
	stdout, _, err = utils.RunCommand(gitDir, "git", "branch", "--show-current")
	if err != nil {
		return fmt.Errorf("ブランチ取得エラー: %v", err)
	}

	currentBranch := strings.TrimSpace(stdout)
	fmt.Printf("現在のブランチ: %s\n", currentBranch)

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

	// git push origin HEADを実行
	fmt.Println("リモートにプッシュしています...")
	stdout, stderr, err := utils.RunCommand(gitDir, "git", "push", "origin", "HEAD")
	if err != nil {
		// pushでエラーが出る場合は force pushを試行
		fmt.Println("⚠️  プッシュエラーが発生しました。強制的にプッシュを試行します...")
		stdout, stderr, err = utils.RunCommand(gitDir, "git", "push", "--force", "origin", "HEAD")
		if err != nil {
			return fmt.Errorf("強制プッシュエラー: %v\nstderr: %s", err, stderr)
		}
		fmt.Println("✅ 強制プッシュが完了しました。")
	}

	if stdout != "" {
		fmt.Printf("プッシュ結果: %s\n", stdout)
	}
	if stderr != "" {
		fmt.Printf("プッシュ情報: %s\n", stderr)
	}

	fmt.Println("✅ リモートへのプッシュが完了しました。")

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
	fmt.Println("\n--- 全ブランチ・タグのプッシュ ---")

	// 全ブランチをプッシュ
	fmt.Println("🌿 全ブランチをプッシュしています...")
	stdout, stderr, err := utils.RunCommand(gitDir, "git", "push", "--all", "origin")
	if err != nil {
		// エラーが発生した場合でも、強制プッシュを試行
		fmt.Println("⚠️  通常のブランチプッシュでエラーが発生しました。強制プッシュを試行します...")
		stdout, stderr, err = utils.RunCommand(gitDir, "git", "push", "--force", "--all", "origin")
		if err != nil {
			fmt.Printf("❌ 全ブランチの強制プッシュに失敗しました: %v\n", err)
			if stderr != "" {
				fmt.Printf("エラー詳細: %s\n", stderr)
			}
			// ブランチプッシュが失敗してもタグプッシュは試行する
		} else {
			fmt.Println("✅ 全ブランチの強制プッシュが完了しました。")
		}
	} else {
		fmt.Println("✅ 全ブランチのプッシュが完了しました。")
	}

	if stdout != "" {
		fmt.Printf("ブランチプッシュ結果: %s\n", stdout)
	}

	// 全タグをプッシュ
	fmt.Println("🏷️  全タグをプッシュしています...")
	stdout, stderr, err = utils.RunCommand(gitDir, "git", "push", "--tags", "origin")
	if err != nil {
		// エラーが発生した場合でも、強制プッシュを試行
		fmt.Println("⚠️  通常のタグプッシュでエラーが発生しました。強制プッシュを試行します...")
		stdout, stderr, err = utils.RunCommand(gitDir, "git", "push", "--force", "--tags", "origin")
		if err != nil {
			fmt.Printf("❌ 全タグの強制プッシュに失敗しました: %v\n", err)
			if stderr != "" {
				fmt.Printf("エラー詳細: %s\n", stderr)
			}
			return fmt.Errorf("タグプッシュエラー: %v", err)
		} else {
			fmt.Println("✅ 全タグの強制プッシュが完了しました。")
		}
	} else {
		fmt.Println("✅ 全タグのプッシュが完了しました。")
	}

	if stdout != "" {
		fmt.Printf("タグプッシュ結果: %s\n", stdout)
	}

	fmt.Println("🚀 全ブランチ・タグのプッシュが完了しました。")
	return nil
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
