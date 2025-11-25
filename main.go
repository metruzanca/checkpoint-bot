package main

//go:generate sqlc generate

import "github.com/metruzanca/checkpoint-bot/cmd"

func main() {
	cmd.Execute()
}
