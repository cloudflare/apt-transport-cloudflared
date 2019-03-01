package main

import (
	"bufio"
	"os"
)

func main() {
	cfd, err := NewCloudflaredMethod(os.Stdout, bufio.NewReader(os.Stdin), "")
	if err != nil {
		return
	}
	cfd.Run()
}
