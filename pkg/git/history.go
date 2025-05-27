package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"git-rewrite-and-go/pkg/utils"
)

// RewriteHistory はGit履歴のauthor/emailを書き換える
func RewriteHistory(gitDir, githubUser, githubEmail string) error {
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
`, githubEmail, githubUser, githubUser, githubEmail,
		githubEmail, githubUser, githubUser, githubEmail)

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

// CreateInitialCommit は初期コミットを作成する
func CreateInitialCommit(gitDir, githubUser, githubEmail string) error {
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
		_, _, err = utils.RunCommand(gitDir, "git", "-c", fmt.Sprintf("user.name=%s", githubUser),
			"-c", fmt.Sprintf("user.email=%s", githubEmail),
			"commit", "--allow-empty", "-m", "Initial commit")
	} else {
		// 変更がある場合は通常のコミット
		_, _, err = utils.RunCommand(gitDir, "git", "-c", fmt.Sprintf("user.name=%s", githubUser),
			"-c", fmt.Sprintf("user.email=%s", githubEmail),
			"commit", "-m", "Initial commit")
	}

	if err != nil {
		return fmt.Errorf("初期コミット作成エラー: %v", err)
	}

	fmt.Println("✅ 初期コミットを作成しました。")
	return nil
}
