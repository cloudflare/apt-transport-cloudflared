package main

import (
	"bufio"
	"os"

	"github.com/cloudflare/apt-transport-cloudflared/apt"
)

func main() {
	cfd, err := apt.NewCloudflaredMethod(os.Stdout, bufio.NewReader(os.Stdin), "")
	if err != nil {
		return
	}
	err = cfd.Run()
	if err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
