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

func handlerPause(game_state *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {
	return func(ps routing.PlayingState) pubsub.AckType {
		defer fmt.Print("> ")
		game_state.HandlePause(ps)
		return pubsub.Ack
	}
}

func handlerMove(game_state *gamelogic.GameState, channel *amqp.Channel) func(move gamelogic.ArmyMove) pubsub.AckType {
	return func(move gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")
		outcome := game_state.HandleMove(move)
		switch outcome {
		case gamelogic.MoveOutComeSafe:
			return pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			err := pubsub.PublishJSON(channel, routing.ExchangePerilTopic, routing.WarRecognitionsPrefix+"."+move.Player.Username, gamelogic.RecognitionOfWar{Attacker: move.Player, Defender: game_state.Player})
			if err != nil {
				log.Fatal("Couldn't publish 'war' message: ", err)
			}
			return pubsub.Ack
		case gamelogic.MoveOutcomeSamePlayer:
			return pubsub.NackDiscard
		default:
			return pubsub.NackDiscard
		}
	}
}

func handlerWar(game_state *gamelogic.GameState) func(rw gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(rw gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")
		outcome, _, _ := game_state.HandleWar(rw)
		switch outcome {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard
		case gamelogic.WarOutcomeOpponentWon:
			return pubsub.Ack
		case gamelogic.WarOutcomeYouWon:
			return pubsub.Ack
		case gamelogic.WarOutcomeDraw:
			return pubsub.Ack
		default:
			fmt.Println("Unknown outcome.")
			return pubsub.NackDiscard
		}
	}
}

func main() {
	fmt.Println("Starting Peril client...")

	conn_url := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(conn_url)
	if err != nil {
		log.Fatal("Couldn't connect to RabbitMQ: ", err)
	}
	defer conn.Close()
	fmt.Println("Connection to RabbitMQ server successful.")

	channel, err := conn.Channel()
	if err != nil {
		log.Fatal("Couldn't open channel: ", err)
	}

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatal(err)
	}

	_, _, err = pubsub.DeclareAndBind(conn, "peril_direct", routing.PauseKey+"."+username, routing.PauseKey, 1)
	if err != nil {
		log.Fatal("Couldn't declare and bind queue: ", err)
	}

	game_state := gamelogic.NewGameState(username)

	err = pubsub.SubscribeJSON(conn, string(routing.ExchangePerilDirect), string(routing.PauseKey)+"."+username, string(routing.PauseKey), 1, handlerPause(game_state))
	if err != nil {
		log.Fatal("Couldn't subscribe to 'pause.*' queue: ", err)
	}

	err = pubsub.SubscribeJSON(conn, string(routing.ExchangePerilTopic), string(routing.ArmyMovesPrefix)+"."+username, string(routing.ArmyMovesPrefix)+".*", 1, handlerMove(game_state, channel))
	if err != nil {
		log.Fatal("Couldn't subscribe to 'army_moves.*' queue: ", err)
	}

	err = pubsub.SubscribeJSON(conn, string(routing.ExchangePerilTopic), string(routing.WarRecognitionsPrefix), string(routing.WarRecognitionsPrefix)+".*", 0, handlerWar(game_state))
	if err != nil {
		log.Fatal("Couldn't subscribe to 'war' queue: ", err)
	}

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
			move, err := game_state.CommandMove(input)
			if err != nil {
				fmt.Println(err)
			}

			err = pubsub.PublishJSON(channel, string(routing.ExchangePerilTopic), string(routing.ArmyMovesPrefix)+"."+username, move)
			if err != nil {
				log.Fatal("Couldn't publish 'move' message: ", err)
			}
			fmt.Println("Move published successfully.")
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
