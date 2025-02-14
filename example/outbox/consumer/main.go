package main

import (
	"flag"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	useAmqpAddress := flag.String("amqp", "amqp://guest:guest@localhost:5677/", "address to amqp broker")

	flag.Parse()

	consumer, err := NewRandomConsumer(*useAmqpAddress)
	if err != nil {
		log.Fatal(err)
	}

	consumer.Consume()
}

type RandomConsumer struct {
	cc *amqp.Channel
}

func NewRandomConsumer(amqpAddress string) (*RandomConsumer, error) {
	conn, err := amqp.Dial(amqpAddress)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	return &RandomConsumer{
		cc: ch,
	}, nil
}

func (c *RandomConsumer) Consume() {
	_ = c.cc.ExchangeDeclare("rnd", amqp.ExchangeTopic, true, false, false, false, nil)

	queue, _ := c.cc.QueueDeclare("outbox", true, false, false, false, nil)

	_ = c.cc.QueueBind(queue.Name, "outbox-subject", "rnd", false, nil)

	ch, _ := c.cc.Consume("outbox", "", false, false, false, false, nil)

	for next := range ch {
		log.Printf("next arrived message: %s", next.Body)

		_ = next.Ack(false)
	}
}
