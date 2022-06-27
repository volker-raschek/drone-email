package main

import (
	"git.cryptic.systems/volker.raschek/drone-email-docker/cmd"
)

var version string

func main() {
	_ = cmd.Execute(version)
}
