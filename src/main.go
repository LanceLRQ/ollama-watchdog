package main

import "github.com/LanceLRQ/ollama-watchdog/cmd"

var version = "v0.0.0"

func main() {
	cmd.CommandEntry(version)
}
