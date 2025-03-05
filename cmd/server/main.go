package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

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

	err = pubsub.PublishJSON(channel, string(routing.ExchangePerilDirect), string(routing.PauseKey), routing.PlayingState{IsPaused: true})

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	if <-signalChan != nil {
		fmt.Println("\nClosing RabbitMQ connection.")
	}
}
