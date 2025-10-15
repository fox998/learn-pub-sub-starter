package state

import (
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/cmd/client/events"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
)

type ClientGameState struct {
	*gamelogic.GameState
	Events *events.EventsData
}

func NewClientGameState() *ClientGameState {
	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("Failed to get username: %s", err)
	}

	internalState := gamelogic.NewGameState(username)
	eventData := events.NewEnventsData(internalState)

	return &ClientGameState{
		GameState: internalState,
		Events:    eventData,
	}
}
