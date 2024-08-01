package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
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
	router := httprouter.New()
	router.POST("/publish/:message", submitHandler)

	fmt.Println("Running...")
	log.Fatal(http.ListenAndServe(":5400", router))
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getRabbitMQConnectionString() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%s/", rabbitUser, rabbitPassword, rabbitHost, rabbitPort)
}

func submitHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	message := p.ByName("message")
	fmt.Println("Received message: " + message)

	if err := publishMessage(message); err != nil {
		log.Errorf("Failed to publish message: %s", err)
		http.Error(w, "Failed to publish message", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Message published successfully"))
}

func publishMessage(message string) error {
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

	err = ch.Publish(
		"",    // exchange
		q.Name, // routing key
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		})
	if err != nil {
		return fmt.Errorf("failed to publish a message: %w", err)
	}

	return nil
}
