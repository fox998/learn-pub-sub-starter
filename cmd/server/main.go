package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub/result"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func setPause(channel *amqp.Channel, isPaused bool) error {
	return pubsub.PublishJSON(
		channel,
		routing.ExchangePerilDirect,
		routing.PauseKey,
		routing.PlayingState{IsPaused: isPaused})
}

func gameLoop(connection *amqp.Connection) {
	chanel, err := connection.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %s", err)
	}
	defer chanel.Close()

	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}

		command := strings.ToLower(words[0])
		var err error = nil
		switch command {
		case "pause":
			fmt.Println("Pausing the game...")
			err = setPause(chanel, true)
		case "resume":
			fmt.Println("Unpausing the game...")
			err = setPause(chanel, false)
		case "quit":
			fmt.Println("Quitting the game...")
			return
		default:
			fmt.Println("Unknown command")
			continue
		}

		if err != nil {
			fmt.Println("Failed to execute command", command)
			fmt.Println(err.Error())
		}
	}
}

func main() {
	fmt.Println("Starting Peril server...")

	url := "amqp://guest:guest@localhost:5672/"
	connection, err := amqp.Dial(url)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err)
	}
	defer connection.Close()
	fmt.Println("Connected to RabbitMQ")

	err = pubsub.SubscribeGob(
		connection,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		fmt.Sprintf("%s.*", routing.GameLogSlug),
		pubsub.DurableQueue,
		func(gameLog routing.GameLog) result.Type {
			defer fmt.Print("> ")
			gamelogic.WriteLog(gameLog)
			return result.Success
		},
	)
	if err != nil {
		log.Fatalf("Failed to subscribe to game logs: %s", err)
	}

	gamelogic.PrintServerHelp()
	gameLoop(connection)
}
