package main

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

var (
	rabbitHost     = getEnv("RABBIT_HOST", "localhost")
	rabbitPort     = getEnv("RABBIT_PORT", "5672")
	rabbitUser     = getEnv("RABBIT_USER", "user")
	rabbitPassword = getEnv("RABBIT_PASSWORD", "password")
)

func main() {
	err := consumer()
	if err != nil {
		log.Fatalf("Failed to start consumer: %s", err)
	}
	fmt.Println("Consumer Running...")
}

func getRabbitMQConnectionString() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%s/", rabbitUser, rabbitPassword, rabbitHost, rabbitPort)
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func consumer() error {
	conn, err := amqp.Dial(getRabbitMQConnectionString())
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	defer conn.Close()
	log.Info("Successfully connected to RabbitMQ")

	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %w", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"publisher",
		false, // durable
		false, // delete when used
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare a queue: %w", err) 
	}

	msgs, err := ch.Consume(
		q.Name,	// queue
		"", 		// consumer
		false,	// auto-check
		false,	// exclusive
		false,	// no-local
		false,	// no-wait
		nil,		// args
	)
	if err != nil {
		log.Fatalf("%s: %s", "failed to publish a message", err)
	}

	forever := make(chan bool)

	go func(){
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)

			d.Ack(false)
		}
	}()

	fmt.Println("Running...")
	<-forever

	return nil
}