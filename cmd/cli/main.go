package main

import (
	"github.com/Raihanarrasyid/iacctl/cmd/cli/command"
	_ "github.com/lib/pq"
)


func main() {
	command.Execute()
}