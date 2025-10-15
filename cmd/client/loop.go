package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bootdotdev/learn-pub-sub-starter/cmd/client/state"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

type commandsMap map[string]func(*state.ClientGameState, []string) error

func NewCommandsMap() commandsMap {
	return commandsMap{
		"spawn": func(gameState *state.ClientGameState, words []string) error {
			return gameState.CommandSpawn(words)
		},
		"move": func(gameState *state.ClientGameState, words []string) error {
			move, err := gameState.CommandMove(words)
			if err != nil {
				return err
			}

			return gameState.Events.PublishMove(move)
		},
		"spam": func(gameState *state.ClientGameState, words []string) error {
			if len(words) < 2 {
				return fmt.Errorf("second integer argument required")
			}

			messageNumber, err := strconv.Atoi(words[1])
			if err != nil {
				return fmt.Errorf("second argument must be an integer")
			}

			for i := 0; i < messageNumber; i++ {
				err = gameState.Events.PublishLog(routing.GameLog{
					CurrentTime: time.Now(),
					Message:     gamelogic.GetMaliciousLog(),
					Username:    gameState.Player.Username,
				})

				if err != nil {
					return fmt.Errorf("failed to publish log: %s", err)
				}
			}

			return nil
		},

		"help": func(gameState *state.ClientGameState, words []string) error {
			gamelogic.PrintClientHelp()
			return nil
		},
	}
}

func gameLoop(state *state.ClientGameState) {

	commandsMap := NewCommandsMap()
	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}

		command := strings.ToLower(words[0])
		if command == "quit" {
			gamelogic.PrintQuit()
			return
		}

		handler, found := commandsMap[command]
		if !found {
			fmt.Println("Unknown command")
			continue
		}

		err := handler(state, words)
		if err != nil {
			fmt.Println("Failed to process command:", err)
			continue
		}
	}
}
