package main

import (
	"fmt"
	"github.com/streadway/amqp"
	"log"
	"os"
)

func main() {

}


func testRabbit() {
	go client()
	go publish(100)

	var a string

	for {
		fmt.Scanln(&a)

		if a == "e" || a == "exit" {
			os.Exit(1)
		}
	}
}

func publish(i int) {
	conn, ch, q := getQueue()
	defer conn.Close()
	defer ch.Close()

	msg := amqp.Publishing{
		ContentType: "text/plain",
	}

	for j := 0; j < i; j++ {
		data := fmt.Sprintf("Hello RabbitMq %v", j)
		msg.Body = []byte(data)
		ch.Publish(
			"",
			q.Name,
			false,
			false,
			msg,
		)
	}
}

func client() {
	conn, ch, q := getQueue()
	defer conn.Close()
	defer ch.Close()

	msgs, err := ch.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)

	failOnError(err, "Failed to register a consumer")

	for msg := range msgs {
		fmt.Printf("Message: %s\n", msg.Body)
		err := msg.Ack(false)
		failOnError(err, "Failed to ack the message")
	}
}

func getQueue() (*amqp.Connection, *amqp.Channel, *amqp.Queue) {
	conn, err := amqp.Dial("amqp://guest@localhost:5672")
	failOnError(err, "Error connecting to message broker")

	channel, err := conn.Channel()
	failOnError(err, "Failed to open channel")

	q, err := channel.QueueDeclare(
		"hello",
		true,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to declare a queue")

	return conn, channel, &q
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}
