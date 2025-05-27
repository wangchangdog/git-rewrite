package main

import (
	"fmt"
	"os"

	"git-rewrite-and-go/pkg/cli"
	"git-rewrite-and-go/pkg/cli/commands"
	"git-rewrite-and-go/pkg/test"
)

// osExit はテスト時にos.Exitをモック可能にするための変数
var osExit = os.Exit

func main() {
	if len(os.Args) < 2 {
		cli.ShowHelp()
		osExit(1)
	}

	command := os.Args[1]

	// --help オプションのチェック
	if command == "--help" || command == "-h" || command == "help" {
		cli.ShowHelp()
		osExit(0)
	}

	var err error
	switch command {
	case "rewrite":
		rewriteCmd := commands.NewRewriteCommand()
		err = rewriteCmd.Execute(os.Args[2:])
	case "demo":
		demoCmd := commands.NewDemoCommand()
		err = demoCmd.Execute(os.Args[2:])
	case "test":
		err = runTests()
	default:
		fmt.Printf("不明なコマンド: %s\n", command)
		fmt.Println("")
		cli.ShowHelp()
		osExit(1)
	}

	if err != nil {
		fmt.Printf("エラー: %v\n", err)
		osExit(1)
	}
}

func runTests() error {
	fmt.Println("テスト機能を実行中...")

	if err := test.RunTests(); err != nil {
		return fmt.Errorf("テスト実行エラー: %v", err)
	}
	return nil
}
