package main

import (
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/cmd/client/state"
)

func main() {
	fmt.Println("Starting Peril client...")

	state := state.NewClientGameState()
	defer state.Events.Close()

	gameLoop(state)

	// subscriptions are automatically cancelled when the connection is closed
}
