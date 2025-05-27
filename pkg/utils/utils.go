package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"
)

// SafeDecode はバイナリデータを安全にデコードする
func SafeDecode(data []byte) string {
	if utf8.Valid(data) {
		return string(data)
	}
	// UTF-8でない場合は、無効な文字を置換してデコード
	return strings.ToValidUTF8(string(data), "�")
}

// ExtractRepoInfoFromURL はリモートURLからユーザー名とリポジトリ名を抽出する
func ExtractRepoInfoFromURL(remoteURL string) (string, string) {
	// URLをトリムして余分な空白文字を除去
	remoteURL = strings.TrimSpace(remoteURL)

	// デバッグモードの場合のみ詳細情報を表示
	debug := os.Getenv("GIT_REWRITE_DEBUG") != ""
	if debug {
		fmt.Printf("デバッグ: ExtractRepoInfoFromURL 入力URL: '%s' (長さ: %d)\n", remoteURL, len(remoteURL))
	}

	// HTTPS形式: https://github.com/user/repo.git
	httpsRegex := regexp.MustCompile(`https://github\.com/([^/]+)/([^/]+?)(?:\.git)?/?$`)
	if matches := httpsRegex.FindStringSubmatch(remoteURL); matches != nil {
		if debug {
			fmt.Printf("デバッグ: HTTPS形式でマッチ - Owner: '%s', Repo: '%s'\n", matches[1], matches[2])
		}
		return matches[1], matches[2]
	}

	// SSH形式: git@github.com:user/repo.git
	sshRegex := regexp.MustCompile(`git@github\.com:([^/]+)/([^/\s]+?)(?:\.git)?/?$`)
	if matches := sshRegex.FindStringSubmatch(remoteURL); matches != nil {
		if debug {
			fmt.Printf("デバッグ: SSH形式でマッチ - Owner: '%s', Repo: '%s'\n", matches[1], matches[2])
		}
		return matches[1], matches[2]
	}

	if debug {
		fmt.Printf("デバッグ: どの形式にもマッチしませんでした\n")
	}
	return "", ""
}

// FindGitDirs は指定されたディレクトリ以下のGitリポジトリを検索する
func FindGitDirs(rootDir string) ([]string, error) {
	var gitDirs []string

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && info.Name() == ".git" {
			// .gitディレクトリの親ディレクトリを追加
			gitDirs = append(gitDirs, filepath.Dir(path))
			return filepath.SkipDir // サブディレクトリをスキップ
		}

		return nil
	})

	return gitDirs, err
}

// RunCommand はコマンドを実行し、結果を返す
func RunCommand(dir, command string, args ...string) (string, string, error) {
	cmd := exec.Command(command, args...)
	cmd.Dir = dir

	stdout, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			stderr := SafeDecode(exitError.Stderr)
			return "", stderr, err
		}
		return "", "", err
	}

	return SafeDecode(stdout), "", nil
}

// CheckEnvironmentVariables は必要な環境変数をチェックする
func CheckEnvironmentVariables() (string, string, error) {
	githubUser := os.Getenv("GITHUB_USER")
	githubEmail := os.Getenv("GITHUB_EMAIL")

	if githubUser == "" {
		return "", "", fmt.Errorf("GITHUB_USER環境変数が設定されていません")
	}

	if githubEmail == "" {
		return "", "", fmt.Errorf("GITHUB_EMAIL環境変数が設定されていません")
	}

	return githubUser, githubEmail, nil
}

// FileExists はファイルが存在するかチェックする
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// GetTargetOwner は設定を考慮してターゲットオーナーを決定する
func GetTargetOwner(defaultUser, owner, organization string) string {
	// 引数で指定された個人リポジトリ所有者が最優先
	if owner != "" {
		return owner
	}
	// 次に引数で指定された組織
	if organization != "" {
		return organization
	}
	// 環境変数からのフォールバック（後方互換性）
	if repoOwner := os.Getenv("GITHUB_REPOSITORY_OWNER"); repoOwner != "" {
		return repoOwner
	}
	if githubOrg := os.Getenv("GITHUB_ORGANIZATION"); githubOrg != "" {
		return githubOrg
	}
	return defaultUser
}

// IsPersonalRepository は個人リポジトリかどうかを判定する
func IsPersonalRepository(owner string) bool {
	// 引数で個人リポジトリ所有者が指定されている場合
	if owner != "" {
		return true
	}
	// 環境変数からのフォールバック（後方互換性）
	return os.Getenv("GITHUB_REPOSITORY_OWNER") != ""
}
