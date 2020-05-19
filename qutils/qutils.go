package qutils

import (
	"fmt"
	"github.com/streadway/amqp"
	"log"
)

const (
	SensorsDisabledExchange = "sensors:disabled"
	SensorsDiscoverExchange = "sensors:discover"
	SensorsListExchange     = "sensors:list"
	WebappDiscoverExchange  = "webapp:discovery"
	WebappSourcesExchange   = "webapp:sources"
	WebappValuesQueue       = "webappValues"
	SensorsListQueue        = "SensorList"
	url                     = "amqp://guest:guest@localhost:5672"
	QueuePersistentValues   = "persistentValues"
)

type exchange struct {
	Name string
	Type string
}

func DeclareFanoutExchange(ex string) {
	conn, ch := GetChannel()
	defer conn.Close()
	defer ch.Close()
	declareExchange(ch, exchange{ex, "fanout"})
}

func DeclareDirectExchange(ex string) {
	conn, ch := GetChannel()
	defer conn.Close()
	defer ch.Close()
	declareExchange(ch, exchange{ex, "direct"})
}

func declareExchange(ch *amqp.Channel, ex exchange) {
	ch.ExchangeDeclare(ex.Name, ex.Type, true, false, false, false, nil)
}

func PublishToFanout(ex string, body []byte) {
	exDto := exchange{ex, "fanout"}
	publish(exDto, "", body)
}

func PublishToDirect(ex string, routingKey string, body []byte) {
	exDto := exchange{ex, "direct"}
	publish(exDto, routingKey, body)
}

func publish(exchange exchange, routingKey string, body []byte) {
	conn, ch := GetChannel()
	defer conn.Close()
	defer ch.Close()

	if exchange.Name != "" {
		declareExchange(ch, exchange)
	}

	msg := amqp.Publishing{
		ContentType: "text/plain",
		Body:        body,
	}

	err := ch.Publish(
		exchange.Name,
		routingKey,
		false,
		false,
		msg,
	)

	if err != nil {
		fmt.Println(err)
	}
}

func DeclareQueue(ex string, name string, autoDelete bool, bindingKeys []string) *amqp.Queue {
	conn, ch := GetChannel()
	defer conn.Close()
	defer ch.Close()
	q := GetQueue(name, autoDelete, ch)

	for _, key := range bindingKeys {
		ch.QueueBind(q.Name, key, ex, false, nil)
	}

	return q
}

func Consume(queueName string, handler func(msg amqp.Delivery)) {
	conn, ch := GetChannel()
	defer conn.Close()
	defer ch.Close()

	msgs, err := ch.Consume(
		queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)

	failOnError(err, "Failed to register a consumer")

	for msg := range msgs {
		handler(msg)
		//fmt.Printf("Message: %s\n", msg.Body)
		err := msg.Ack(false)
		failOnError(err, "Failed to ack the message")
	}
}

func GetChannel() (*amqp.Connection, *amqp.Channel) {
	conn, err := amqp.Dial(url)
	failOnError(err, "Error connecting to message broker")

	channel, err := conn.Channel()
	failOnError(err, "Failed to open channel")

	return conn, channel
}

func GetQueue(name string, autoDelete bool, channel *amqp.Channel) *amqp.Queue {
	q, err := channel.QueueDeclare(
		name,
		false,
		autoDelete,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to declare a queue")

	return &q
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
