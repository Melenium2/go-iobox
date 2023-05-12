package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/streadway/amqp"
)

func main() {
	useAmqpAddress := flag.String("amqp", "amqp://guest:guest@localhost:5677/", "address to amqp broker")

	flag.Parse()

	broker, err := NewRandomPublisher(*useAmqpAddress)
	if err != nil {
		log.Fatal(err)
	}

	broker.Cycle()
}

type RandomPublisher struct {
	cc *amqp.Channel
}

func NewRandomPublisher(amqpAddress string) (*RandomPublisher, error) {
	conn, err := amqp.Dial(amqpAddress)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	return &RandomPublisher{
		cc: ch,
	}, nil
}

func (p *RandomPublisher) Publish(ctx context.Context, subject string, body []byte) error {
	return p.cc.Publish(
		"",
		subject,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

func (p *RandomPublisher) Cycle() {
	ticker := time.NewTicker(5 * time.Second)

	for range ticker.C {
		p.publishFirstSubject()
		p.publishSecondSubject()
	}
}

type payload struct {
	ID      uuid.UUID      `json:"id"`
	Payload map[string]any `json:"payload"`
}

func (p *RandomPublisher) publishFirstSubject() {
	body := payload{
		ID: uuid.New(),
		Payload: map[string]any{
			"1": "1",
		},
	}

	raw, err := json.Marshal(body)
	if err != nil {
		log.Println(err.Error())
	}

	log.Printf("sent event to multy-handlers %s", raw)

	if err := p.Publish(context.Background(), "multy-handlers", raw); err != nil {
		log.Println(err.Error())
	}
}

func (p *RandomPublisher) publishSecondSubject() {
	body := payload{
		ID: uuid.New(),
		Payload: map[string]any{
			"2": "2",
		},
	}

	raw, err := json.Marshal(body)
	if err != nil {
		log.Println(err.Error())
	}

	log.Printf("sent event to single-handler %s", raw)

	if err := p.Publish(context.Background(), "single-handler", raw); err != nil {
		log.Println(err.Error())
	}
}
