package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestMainFunction はmain関数の基本的な動作をテストする
func TestMainFunction(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedOutput string
		expectError    bool
	}{
		{
			name:           "引数なしでヘルプを表示",
			args:           []string{},
			expectedOutput: "使用方法:",
			expectError:    true,
		},
		{
			name:           "不明なコマンド",
			args:           []string{"unknown"},
			expectedOutput: "不明なコマンド: unknown",
			expectError:    true,
		},
		{
			name:           "testコマンド",
			args:           []string{"test"},
			expectedOutput: "Git Rewrite Tools テスト実行",
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// os.Argsを一時的に変更
			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()

			// プログラム名 + テスト引数を設定
			os.Args = append([]string{"git-rewrite"}, tt.args...)

			// 標準出力をキャプチャするためのパイプを作成
			oldStdout := os.Stdout
			oldStderr := os.Stderr
			defer func() {
				os.Stdout = oldStdout
				os.Stderr = oldStderr
			}()

			// main関数を別のgoroutineで実行し、panicをキャッチ
			var exitCode int
			func() {
				defer func() {
					if r := recover(); r != nil {
						// os.Exit()によるpanicをキャッチ
						if exitError, ok := r.(exitError); ok {
							exitCode = int(exitError)
						} else {
							t.Fatalf("予期しないpanic: %v", r)
						}
					}
				}()

				// os.Exitをモック
				osExit = func(code int) {
					panic(exitError(code))
				}
				defer func() {
					osExit = os.Exit
				}()

				main()
			}()

			// 期待される終了コードをチェック
			if tt.expectError && exitCode == 0 {
				t.Errorf("エラーが期待されましたが、正常終了しました")
			}
			if !tt.expectError && exitCode != 0 {
				t.Errorf("正常終了が期待されましたが、エラーコード %d で終了しました", exitCode)
			}
		})
	}
}

// TestBinaryExecution はビルドされたバイナリの実行をテストする
func TestBinaryExecution(t *testing.T) {
	// バイナリが存在するかチェック
	binaryPath := "./git-rewrite"
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Skip("バイナリが見つかりません。先に 'go build' を実行してください")
	}

	tests := []struct {
		name           string
		args           []string
		expectedOutput string
		expectError    bool
	}{
		{
			name:           "ヘルプ表示",
			args:           []string{},
			expectedOutput: "使用方法:",
			expectError:    true,
		},
		{
			name:           "testコマンド実行",
			args:           []string{"test"},
			expectedOutput: "Git Rewrite Tools テスト実行",
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tt.args...)
			output, err := cmd.CombinedOutput()
			outputStr := string(output)

			if tt.expectError && err == nil {
				t.Errorf("エラーが期待されましたが、正常終了しました。出力: %s", outputStr)
			}
			if !tt.expectError && err != nil {
				t.Errorf("正常終了が期待されましたが、エラーが発生しました: %v。出力: %s", err, outputStr)
			}

			if !strings.Contains(outputStr, tt.expectedOutput) {
				t.Errorf("期待される出力が見つかりません。期待: %s、実際: %s", tt.expectedOutput, outputStr)
			}
		})
	}
}

// exitError はos.Exit()をテスト可能にするための型
type exitError int

func (e exitError) Error() string {
	return "exit"
}
