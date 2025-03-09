package main

import (
	"fmt"
	"log"
	"os"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func gameLogsHandler() func(game_log routing.GameLog) pubsub.AckType {
	return func(game_log routing.GameLog) pubsub.AckType {
		defer fmt.Print("> ")
		err := gamelogic.WriteLog(game_log)
		if err != nil {
			return pubsub.NackRequeue
		}
		return pubsub.Ack
	}
}

func main() {
	fmt.Println("Starting Peril server...")

	conn_url := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(conn_url)
	if err != nil {
		log.Fatal("Couldn't connect to RabbitMQ: ", err)
	}
	defer conn.Close()
	fmt.Println("Connection to RabbitMQ server successful.")

	channel, err := conn.Channel()
	if err != nil {
		log.Fatal("Couldn't create channel: ", err)
	}

	err = pubsub.DeclareExchange(channel, "peril_direct", "direct")
	if err != nil {
		log.Fatal("Error declaring 'peril_direct' exchange:", err)
	}
	err = pubsub.DeclareExchange(channel, "peril_topic", "topic")
	if err != nil {
		log.Fatal("Error declaring 'peril_topic' exchange:", err)
	}
	err = pubsub.DeclareExchange(channel, "peril_dlx", "fanout")
	if err != nil {
		log.Fatal("Error declaring 'peril_dlx' exchange:", err)
	}

	_, _, err = pubsub.DeclareAndBindQueue(conn, "peril_dlx", "peril_dlq", "", 0)
	if err != nil {
		log.Fatal("Error declaring and binding 'peril_dlq' queue:", err)
	}
	_, _, err = pubsub.DeclareAndBindQueue(conn, routing.ExchangePerilTopic, routing.GameLogSlug, routing.GameLogSlug+".*", 0)
	if err != nil {
		log.Fatal("Error declaring and binding 'game_logs' queue: ", err)
	}

	err = pubsub.SubscribeGob(conn, string(routing.ExchangePerilTopic), string(routing.GameLogSlug), string(routing.GameLogSlug)+".*", 0, gameLogsHandler())
	if err != nil {
		log.Fatal("Error subscribing to 'game_logs' queue: ", err)
	}

	gamelogic.PrintServerHelp()

	for {
		fmt.Println()
		input := gamelogic.GetInput()
		if input == nil || len(input) == 0 {
			continue
		}
		if input[0] == "pause" {
			fmt.Println("Pausing the game...")
			err = pubsub.PublishJSON(channel, string(routing.ExchangePerilDirect), string(routing.PauseKey), routing.PlayingState{IsPaused: true})
			if err != nil {
				log.Fatal("Error sending 'pause' message: ", err)
			}
		} else if input[0] == "resume" {
			fmt.Println("Resuming the game...")
			err = pubsub.PublishJSON(channel, string(routing.ExchangePerilDirect), string(routing.PauseKey), routing.PlayingState{IsPaused: false})
			if err != nil {
				log.Fatal("Error sending 'resume' message: ", err)
			}
		} else if input[0] == "quit" {
			fmt.Println("Quiting the game...")
			fmt.Println("\nShutting down Peril server.")
			os.Exit(0)
		} else if input[0] == "help" {
			gamelogic.PrintServerHelp()
		} else {
			fmt.Println("Unknown command.")
		}
	}
}
