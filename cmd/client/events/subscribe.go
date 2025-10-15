package events

import (
	"fmt"
	"log"
	"time"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub/result"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

func (e *EventsData) subscribeToPlayingState(state *gamelogic.GameState) error {
	return pubsub.SubscribeJSON(
		e.Connection,
		routing.ExchangePerilDirect,
		fmt.Sprintf("%s.%s", routing.PauseKey, e.postfix),
		routing.PauseKey,
		pubsub.TransientQueue,
		func(plyingState routing.PlayingState) result.Type {
			defer fmt.Print("> ")
			state.HandlePause(plyingState)
			return result.Success
		})
}

func (e *EventsData) subscribeToMoves(state *gamelogic.GameState) error {
	return pubsub.SubscribeJSON(
		e.Connection,
		routing.ExchangePerilTopic,
		fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, e.postfix),
		routing.ArmyMovesPrefix+".*",
		pubsub.TransientQueue,
		func(move gamelogic.ArmyMove) result.Type {
			defer fmt.Print("> ")
			outcome := state.HandleMove(move)
			if outcome == gamelogic.MoveOutComeSafe {
				return result.Success
			}

			if outcome != gamelogic.MoveOutcomeMakeWar {
				return result.Discard
			}

			err := e.PublishWar(gamelogic.RecognitionOfWar{
				Attacker: move.Player,
				Defender: state.Player,
			})
			if err != nil {
				log.Printf("Failed to publish war: %s", err)
				return result.RetryRequire
			}

			return result.Success

		})
}

func getLogMessage(outcome gamelogic.WarOutcome, winner, loser string) string {
	switch outcome {
	case gamelogic.WarOutcomeOpponentWon, gamelogic.WarOutcomeYouWon:
		return fmt.Sprintf("%v won a war against %v", winner, loser)
	case gamelogic.WarOutcomeDraw:
		return fmt.Sprintf("A war between %v and %v resulted in a draw", winner, loser)
	default:
		log.Printf("Unexpected war outcome for log: %v", outcome)
		return ""
	}
}

func (e *EventsData) subscribeToWars(state *gamelogic.GameState) error {
	return pubsub.SubscribeJSON(
		e.Connection,
		routing.ExchangePerilTopic,
		routing.WarRecognitionsPrefix,
		routing.WarRecognitionsPrefix+".*",
		pubsub.DurableQueue,
		func(war gamelogic.RecognitionOfWar) result.Type {
			defer fmt.Print("> ")
			outcome, winner, loser := state.HandleWar(war)
			switch outcome {
			case gamelogic.WarOutcomeNotInvolved:
				return result.RetryRequire
			case gamelogic.WarOutcomeNoUnits:
				return result.Discard
			case gamelogic.WarOutcomeOpponentWon, gamelogic.WarOutcomeYouWon, gamelogic.WarOutcomeDraw:
				if war.Attacker.Username != state.Player.Username {
					return result.Success
				}

				err := e.PublishLog(routing.GameLog{
					CurrentTime: time.Now(),
					Message:     getLogMessage(outcome, winner, loser),
					Username:    state.Player.Username,
				})

				if err != nil {
					log.Printf("Failed to publish war log: %s", err)
					return result.RetryRequire
				}
				return result.Success
			default:
				log.Printf("Unknown war outcome: %d", outcome)
				return result.Discard
			}
		})
}

func (e *EventsData) SubscribeToQueues(state *gamelogic.GameState) {
	if err := e.subscribeToPlayingState(state); err != nil {
		log.Fatalf("Failed to subscribe to playing state: %s", err)
	}

	if err := e.subscribeToMoves(state); err != nil {
		log.Fatalf("Failed to subscribe to moves: %s", err)
	}

	if err := e.subscribeToWars(state); err != nil {
		log.Fatalf("Failed to subscribe to wars: %s", err)
	}
}
