package main

import (
	"bufio"
	"os"

	"github.com/cloudflare/apt-transport-cloudflared/apt"
)

func run() int {
	cfd, err := apt.NewCloudflaredMethod(os.Stdout, bufio.NewReader(os.Stdin))
	if err != nil {
		return 1
	}
	defer cfd.Close()

	err = cfd.Run()

	if err != nil {
		return 1
	}
	return 0
}

func main() {
	os.Exit(run())
}
