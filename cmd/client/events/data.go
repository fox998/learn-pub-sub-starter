package events

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

type ChannelData struct {
	Channel  *amqp.Channel
	Queue    amqp.Queue
	Exchange string
}

type EventsData struct {
	Connection  *amqp.Connection
	MoveChannel ChannelData
	WarChannel  ChannelData
	LogChannel  ChannelData
	postfix     string
}

func (eventData *EventsData) Close() {
	eventData.Connection.Close()

	eventData.MoveChannel.Channel.Close()
	eventData.WarChannel.Channel.Close()
	eventData.LogChannel.Channel.Close()
}

func (eventData *EventsData) createChannel(prefix string) ChannelData {
	key := fmt.Sprintf("%s.%s", prefix, eventData.postfix)
	exchage := routing.ExchangePerilTopic
	channel, queue, err := pubsub.DeclareAndBind(
		eventData.Connection,
		exchage,
		key,
		key,
		pubsub.TransientQueue,
	)

	if err != nil {
		log.Fatalf("Failed to declare and bind a queue fot %v key: %s", key, err)
	}

	return ChannelData{
		Channel:  channel,
		Queue:    queue,
		Exchange: exchage,
	}
}

func NewEnventsData(state *gamelogic.GameState) *EventsData {

	url := "amqp://guest:guest@localhost:5672/"
	connection, err := amqp.Dial(url)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err)
	}

	eventData := &EventsData{
		Connection: connection,
		postfix:    state.GetUsername(),
	}

	eventData.MoveChannel = eventData.createChannel(routing.ArmyMovesPrefix)
	eventData.WarChannel = eventData.createChannel(routing.WarRecognitionsPrefix)
	eventData.LogChannel = eventData.createChannel(routing.GameLogSlug)

	eventData.SubscribeToQueues(state)

	return eventData
}
