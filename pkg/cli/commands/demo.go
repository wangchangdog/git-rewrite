package commands

import (
	"fmt"
	"os"

	"git-rewrite-and-go/pkg/cli/config"
	"git-rewrite-and-go/pkg/demo"
)

// DemoCommand はdemoコマンドを実行する
type DemoCommand struct{}

// NewDemoCommand は新しいDemoCommandを作成する
func NewDemoCommand() *DemoCommand {
	return &DemoCommand{}
}

// Execute はdemoコマンドを実行する
func (c *DemoCommand) Execute(args []string) error {
	// demoコマンドも新しい引数形式をサポート
	config, err := config.ParseRewriteArgs(args)
	if err != nil {
		fmt.Printf("引数解析エラー: %v\n", err)
		fmt.Println("")
		fmt.Println("使用方法: git-rewrite demo <github_token> --user <user> --email <email>")
		return err
	}

	// デバッグモードの設定
	if config.Debug {
		os.Setenv("GIT_REWRITE_DEBUG", "1")
	}

	// デモを実行
	if err := demo.RunDemo(config.GitHubToken); err != nil {
		return fmt.Errorf("デモ実行エラー: %v", err)
	}

	if err := demo.RunEmptyRepoDemo(config.GitHubToken, config.GitHubUser, config.GitHubEmail); err != nil {
		return fmt.Errorf("空のリポジトリデモエラー: %v", err)
	}

	return nil
}
