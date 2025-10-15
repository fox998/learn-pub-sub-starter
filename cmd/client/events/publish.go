package events

import (
	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	"github.com/rabbitmq/amqp091-go"
)

type publishMethod[T any] func(ch *amqp091.Channel, exchange, key string, val T) error

func publishWithMethod[T any](eventData *EventsData, channelData ChannelData, method publishMethod[T], val T) error {
	return method(
		channelData.Channel,
		channelData.Exchange,
		channelData.Queue.Name,
		val,
	)
}

func (eventData *EventsData) PublishMove(move gamelogic.ArmyMove) error {
	return publishWithMethod(
		eventData,
		eventData.MoveChannel,
		pubsub.PublishJSON[gamelogic.ArmyMove],
		move)
}

func (eventData *EventsData) PublishWar(war gamelogic.RecognitionOfWar) error {
	return publishWithMethod(
		eventData,
		eventData.WarChannel,
		pubsub.PublishJSON[gamelogic.RecognitionOfWar],
		war)
}

func (eventData *EventsData) PublishLog(gameLog routing.GameLog) error {
	return publishWithMethod(
		eventData,
		eventData.LogChannel,
		pubsub.PublishGob,
		gameLog)
}
