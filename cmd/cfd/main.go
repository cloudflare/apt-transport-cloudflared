package main

import (
	"bufio"
	"io"
	"os"

	"github.com/cloudflare/apt-transport-cloudflared/apt"
)

func run(outfp io.Writer, infp io.Reader) int {
	cfd, err := apt.NewCloudflaredMethod(outfp, bufio.NewReader(infp))
	if err != nil {
		return 1
	}

	if cfd.Run() {
		return 0
	}
	return 1
}

func main() {
	os.Exit(run(os.Stdout, os.Stdin))
}
