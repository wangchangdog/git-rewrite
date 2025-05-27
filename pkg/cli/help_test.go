package cli

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

// TestShowHelp はShowHelp関数をテストする
func TestShowHelp(t *testing.T) {
	// 標準出力をキャプチャ
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// ShowHelpを実行
	ShowHelp()

	// 標準出力を復元
	w.Close()
	os.Stdout = oldStdout

	// 出力を読み取り
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// 期待される文字列が含まれているかチェック
	expectedStrings := []string{
		"使用方法:",
		"git-rewrite <command> [options]",
		"git-rewrite --help",
		"利用可能なコマンド:",
		"rewrite",
		"demo",
		"test",
		"help, --help, -h",
		"rewriteコマンドのオプション:",
		"--user, -u",
		"--email, -e",
		"--target-dir, -d",
		"--owner, -o",
		"--organization",
		"--collaborators",
		"--collaborator-config, -c",
		"--push-all",
		"--debug",
		"--public",
		"例:",
		"後方互換性:",
		"環境変数も引き続きサポートされます",
		"GITHUB_USER",
		"GITHUB_EMAIL",
		"GITHUB_ORGANIZATION",
		"GITHUB_REPOSITORY_OWNER",
		"GITHUB_COLLABORATORS",
		"GIT_REWRITE_DEBUG",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("期待される文字列が見つかりません: '%s'\n出力: %s", expected, output)
		}
	}
}

// TestShowHelpOutput はShowHelp関数の出力内容をより詳細にテストする
func TestShowHelpOutput(t *testing.T) {
	// 標準出力をキャプチャ
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// ShowHelpを実行
	ShowHelp()

	// 標準出力を復元
	w.Close()
	os.Stdout = oldStdout

	// 出力を読み取り
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// 出力が空でないことを確認
	if len(output) == 0 {
		t.Error("ShowHelpの出力が空です")
	}

	// 各セクションが含まれているかチェック
	sections := []string{
		"使用方法:",
		"利用可能なコマンド:",
		"rewriteコマンドのオプション:",
		"例:",
		"後方互換性:",
	}

	for _, section := range sections {
		if !strings.Contains(output, section) {
			t.Errorf("セクション '%s' が見つかりません", section)
		}
	}

	// コマンドの説明が含まれているかチェック
	commands := []string{
		"rewrite",
		"demo",
		"test",
		"help",
	}

	for _, command := range commands {
		if !strings.Contains(output, command) {
			t.Errorf("コマンド '%s' の説明が見つかりません", command)
		}
	}

	// オプションの説明が含まれているかチェック
	options := []string{
		"--user",
		"--email",
		"--target-dir",
		"--owner",
		"--organization",
		"--collaborators",
		"--collaborator-config",
		"--push-all",
		"--debug",
		"--public",
	}

	for _, option := range options {
		if !strings.Contains(output, option) {
			t.Errorf("オプション '%s' の説明が見つかりません", option)
		}
	}
}

// TestShowHelpFormat はShowHelp関数の出力フォーマットをテストする
func TestShowHelpFormat(t *testing.T) {
	// 標準出力をキャプチャ
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// ShowHelpを実行
	ShowHelp()

	// 標準出力を復元
	w.Close()
	os.Stdout = oldStdout

	// 出力を読み取り
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// 行数をチェック（適切な量の情報が含まれているか）
	lines := strings.Split(output, "\n")
	if len(lines) < 20 {
		t.Errorf("ヘルプの出力が短すぎます。行数: %d", len(lines))
	}

	// 空行が適切に含まれているかチェック（読みやすさのため）
	hasEmptyLines := false
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			hasEmptyLines = true
			break
		}
	}

	if !hasEmptyLines {
		t.Error("ヘルプの出力に空行が含まれていません（読みやすさのため空行が必要）")
	}

	// 日本語が含まれているかチェック
	if !containsJapanese(output) {
		t.Error("ヘルプの出力に日本語が含まれていません")
	}
}

// containsJapanese は文字列に日本語が含まれているかチェックする
func containsJapanese(s string) bool {
	for _, r := range s {
		if (r >= 0x3040 && r <= 0x309F) || // ひらがな
			(r >= 0x30A0 && r <= 0x30FF) || // カタカナ
			(r >= 0x4E00 && r <= 0x9FAF) { // 漢字
			return true
		}
	}
	return false
}
