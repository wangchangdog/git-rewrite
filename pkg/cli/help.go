package cli

import "fmt"

// ShowHelp はアプリケーションのヘルプを表示する
func ShowHelp() {
	fmt.Println("使用方法:")
	fmt.Println("  git-rewrite <command> [options]")
	fmt.Println("  git-rewrite --help")
	fmt.Println("")
	fmt.Println("利用可能なコマンド:")
	fmt.Println("  rewrite <github_token> --user <user> --email <email> [options] - Git履歴の書き換えとリモートリポジトリ管理")
	fmt.Println("  demo <github_token> --user <user> --email <email>              - リモートリポジトリ作成機能のデモ")
	fmt.Println("  test                                                           - テストの実行")
	fmt.Println("  help, --help, -h                                               - このヘルプを表示")
	fmt.Println("")
	fmt.Println("rewriteコマンドのオプション:")
	fmt.Println("  --user, -u <username>           GitHubユーザー名（必須）")
	fmt.Println("  --email, -e <email>             GitHubメールアドレス（必須）")
	fmt.Println("  --target-dir, -d <directory>    対象ディレクトリ（デフォルト: .）")
	fmt.Println("  --owner, -o <owner>             個人リポジトリ所有者（最高優先度）")
	fmt.Println("  --organization <org>            組織名")
	fmt.Println("  --collaborators <list>          コラボレーター設定（例: user1:push,user2:admin）")
	fmt.Println("  --collaborator-config, -c <file> コラボレーター設定ファイル")
	fmt.Println("  --push-all                      全ブランチ・タグをプッシュ")
	fmt.Println("  --debug                         デバッグモード")
	fmt.Println("  --public                        パブリックリポジトリとして作成（デフォルト: プライベート）")
	fmt.Println("  --enable-actions                GitHub Actions制御を無効化（デフォルトでActions制御は有効）")
	fmt.Println("")
	fmt.Println("例:")
	fmt.Println("  git-rewrite --help")
	fmt.Println("  git-rewrite test")
	fmt.Println("  git-rewrite rewrite ghp_xxx --user myuser --email my@email.com")
	fmt.Println("  git-rewrite rewrite ghp_xxx --user myuser --email my@email.com --target-dir ~/projects")
	fmt.Println("  git-rewrite rewrite ghp_xxx --user myuser --email my@email.com --organization myorg")
	fmt.Println("  git-rewrite rewrite ghp_xxx --user myuser --email my@email.com --owner specificuser")
	fmt.Println("  git-rewrite rewrite ghp_xxx --user myuser --email my@email.com --collaborators \"dev1:push,admin1:admin\"")
	fmt.Println("  git-rewrite rewrite ghp_xxx --user myuser --email my@email.com --collaborator-config collaborators.json --push-all")
	fmt.Println("  git-rewrite rewrite ghp_xxx --user myuser --email my@email.com --public --debug")
	fmt.Println("  git-rewrite rewrite ghp_xxx --user myuser --email my@email.com --enable-actions")
	fmt.Println("  git-rewrite demo ghp_xxx --user myuser --email my@email.com")
	fmt.Println("")
	fmt.Println("GitHub Actions制御について:")
	fmt.Println("  デフォルトでは、プッシュ前にGitHub Actionsを無効化し、プッシュ後に有効化します。")
	fmt.Println("  これにより、プッシュ時にActionsが実行されることを防げます。")
	fmt.Println("  --enable-actionsオプションを使用すると、この制御を無効化できます。")
	fmt.Println("後方互換性:")
	fmt.Println("  環境変数も引き続きサポートされますが、コマンド引数が優先されます。")
	fmt.Println("  GITHUB_USER, GITHUB_EMAIL, GITHUB_ORGANIZATION, GITHUB_REPOSITORY_OWNER,")
	fmt.Println("  GITHUB_COLLABORATORS, GIT_REWRITE_DEBUG")
}
