package main

import (
	"os"

	"github.com/txtweet/test_velo_b1/cmd"

	_ "github.com/txtweet/test_velo_b1/include"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
