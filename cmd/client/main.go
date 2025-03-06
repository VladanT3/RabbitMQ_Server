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
	fmt.Println("Starting Peril client...")

	conn_url := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(conn_url)
	if err != nil {
		log.Fatal("Couldn't connect to RabbitMQ: ", err)
	}
	defer conn.Close()
	fmt.Println("Connection to RabbitMQ server successful.")

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatal(err)
	}

	_, _, err = pubsub.DeclareAndBind(conn, "peril_direct", routing.PauseKey+"."+username, routing.PauseKey, 1)
	if err != nil {
		log.Fatal("Couldn't declare and bind queue: ", err)
	}

	game_state := gamelogic.NewGameState(username)

	for {
		fmt.Println()
		input := gamelogic.GetInput()
		if input == nil || len(input) == 0 {
			continue
		}
		if input[0] == "spawn" {
			err := game_state.CommandSpawn(input)
			if err != nil {
				fmt.Println(err)
			}
		} else if input[0] == "move" {
			_, err := game_state.CommandMove(input)
			if err != nil {
				fmt.Println(err)
			}
		} else if input[0] == "status" {
			game_state.CommandStatus()
		} else if input[0] == "help" {
			gamelogic.PrintClientHelp()
		} else if input[0] == "spam" {
			fmt.Println("Spamming not allowed yet.")
		} else if input[0] == "quit" {
			gamelogic.PrintQuit()
			fmt.Println("\nClosing Peril client.")
			os.Exit(0)
		} else {
			fmt.Println("Unknown command.")
		}
	}
}
