package main

import (
	"fmt"
	"os"

	"aliyun-exporter/cmd"
)

func main() {
	if err := cmd.NewRootCommand().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
