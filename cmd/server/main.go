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
