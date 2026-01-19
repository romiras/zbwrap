package main

import (
	"zbwrap/internal/commands"
	"zbwrap/internal/initializers"
)

func main() {
	initializers.Load()
	commands.Execute()
}
