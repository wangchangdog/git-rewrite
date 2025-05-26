package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestBinaryIntegration はバイナリの統合テストを実行する
func TestBinaryIntegration(t *testing.T) {
	// バイナリのパスを設定
	binaryPath := "../git-rewrite-tools"

	// バイナリが存在するかチェック
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		// バイナリをビルド
		t.Log("バイナリが見つかりません。ビルドを実行します...")
		cmd := exec.Command("go", "build", "-o", "git-rewrite-tools", ".")
		cmd.Dir = ".."
		if err := cmd.Run(); err != nil {
			t.Fatalf("バイナリのビルドに失敗しました: %v", err)
		}
	}

	t.Run("ヘルプ表示テスト", func(t *testing.T) {
		cmd := exec.Command(binaryPath)
		output, err := cmd.CombinedOutput()

		// ヘルプ表示時は終了コード1が期待される
		if err == nil {
			t.Error("ヘルプ表示時にエラーが期待されましたが、正常終了しました")
		}

		outputStr := string(output)
		expectedStrings := []string{
			"使用方法:",
			"git-rewrite-tools <command> [options]",
			"rewrite",
			"demo",
			"test",
		}

		for _, expected := range expectedStrings {
			if !strings.Contains(outputStr, expected) {
				t.Errorf("期待される文字列が見つかりません: %s\n出力: %s", expected, outputStr)
			}
		}
	})

	t.Run("testコマンド実行テスト", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "test")
		output, err := cmd.CombinedOutput()

		if err != nil {
			t.Errorf("testコマンドの実行に失敗しました: %v\n出力: %s", err, string(output))
		}

		outputStr := string(output)
		expectedStrings := []string{
			"Git Rewrite Tools テスト実行",
			"すべてのテストが成功しました",
		}

		for _, expected := range expectedStrings {
			if !strings.Contains(outputStr, expected) {
				t.Errorf("期待される文字列が見つかりません: %s\n出力: %s", expected, outputStr)
			}
		}
	})

	t.Run("不明なコマンドテスト", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "unknown-command")
		output, err := cmd.CombinedOutput()

		// 不明なコマンド時は終了コード1が期待される
		if err == nil {
			t.Error("不明なコマンド時にエラーが期待されましたが、正常終了しました")
		}

		outputStr := string(output)
		if !strings.Contains(outputStr, "不明なコマンド: unknown-command") {
			t.Errorf("期待されるエラーメッセージが見つかりません。出力: %s", outputStr)
		}
	})
}

// TestBinaryWithMockGitRepo はモックGitリポジトリを使用したテストを実行する
func TestBinaryWithMockGitRepo(t *testing.T) {
	// 一時ディレクトリを作成
	tempDir, err := os.MkdirTemp("", "git_test_")
	if err != nil {
		t.Fatalf("一時ディレクトリ作成エラー: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// モックGitリポジトリを作成
	gitDir := filepath.Join(tempDir, "test-repo")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("テストディレクトリ作成エラー: %v", err)
	}

	// Gitリポジトリを初期化
	cmd := exec.Command("git", "init")
	cmd.Dir = gitDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("git init エラー: %v", err)
	}

	// テストファイルを作成
	testFile := filepath.Join(gitDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("テストファイル作成エラー: %v", err)
	}

	// 初期コミット
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = gitDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("git add エラー: %v", err)
	}

	cmd = exec.Command("git", "-c", "user.name=Test User", "-c", "user.email=test@example.com", "commit", "-m", "Initial commit")
	cmd.Dir = gitDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("git commit エラー: %v", err)
	}

	t.Run("rewriteコマンド引数不足テスト", func(t *testing.T) {
		binaryPath := "../git-rewrite-tools"
		cmd := exec.Command(binaryPath, "rewrite")
		output, err := cmd.CombinedOutput()

		// 引数不足時は終了コード1が期待される
		if err == nil {
			t.Error("引数不足時にエラーが期待されましたが、正常終了しました")
		}

		outputStr := string(output)
		if !strings.Contains(outputStr, "使用方法: git-rewrite-tools rewrite") {
			t.Errorf("期待される使用方法メッセージが見つかりません。出力: %s", outputStr)
		}
	})

	t.Run("demoコマンド引数不足テスト", func(t *testing.T) {
		binaryPath := "../git-rewrite-tools"
		cmd := exec.Command(binaryPath, "demo")
		output, err := cmd.CombinedOutput()

		// 引数不足時は終了コード1が期待される
		if err == nil {
			t.Error("引数不足時にエラーが期待されましたが、正常終了しました")
		}

		outputStr := string(output)
		if !strings.Contains(outputStr, "使用方法: git-rewrite-tools demo") {
			t.Errorf("期待される使用方法メッセージが見つかりません。出力: %s", outputStr)
		}
	})
}
