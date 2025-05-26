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
	// HTTPS形式: https://github.com/user/repo.git
	httpsRegex := regexp.MustCompile(`https://github\.com/([^/]+)/([^/]+?)(?:\.git)?/?$`)
	if matches := httpsRegex.FindStringSubmatch(remoteURL); matches != nil {
		return matches[1], matches[2]
	}

	// SSH形式: git@github.com:user/repo.git
	sshRegex := regexp.MustCompile(`git@github\.com:([^/]+)/([^/]+?)(?:\.git)?/?$`)
	if matches := sshRegex.FindStringSubmatch(remoteURL); matches != nil {
		return matches[1], matches[2]
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