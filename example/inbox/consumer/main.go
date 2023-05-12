package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/streadway/amqp"

	"github.com/Melenium2/go-iobox/inbox"
)

func main() {
	useAmqpAddress := flag.String("amqp", "amqp://guest:guest@localhost:5677/", "address to amqp broker")
	useHost := flag.String("host", "localhost", "host of database")
	usePort := flag.String("port", "5437", "port of database")
	useUser := flag.String("user", "postgres", "database use")
	usePass := flag.String("pass", "postgres", "password to databsae user")
	useDatabase := flag.String("db", "outbox", "which database use")

	flag.Parse()

	address := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		*useHost, *usePort, *useUser, *usePass, *useDatabase,
	)

	db, err := sql.Open("postgres", address)
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	consumer, err := NewRandomConsumer(*useAmqpAddress)
	if err != nil {
		log.Fatal(err)
	}

	consumer.Declare()

	registy := inbox.NewRegistry()
	{
		registy.On("multy-handlers", &firstMultiHandler{}, &secondMultiHandler{})
		registy.On("single-handler", &singleHandler{})
	}

	inboxStorage := inbox.NewInbox(registy, db)

	inboxStorage.Start(context.Background())

	consumer.Consume(inboxStorage.Writer())
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

func (c *RandomConsumer) Declare() {
	_ = c.cc.ExchangeDeclare("inbox-exch", amqp.ExchangeTopic, true, false, false, false, nil)

	queue1, _ := c.cc.QueueDeclare("multy-handlers", true, false, false, false, nil)

	_ = c.cc.QueueBind(queue1.Name, "multy-handlers-subject", "inbox-exch", false, nil)

	queue2, _ := c.cc.QueueDeclare("single-handler", true, false, false, false, nil)

	_ = c.cc.QueueBind(queue2.Name, "single-handler-subject", "inbox-exch", false, nil)
}

type rawBody struct {
	ID uuid.UUID `json:"id"`
}

func (c *RandomConsumer) Consume(writer *inbox.Client) {
	ch1, _ := c.cc.Consume("multy-handlers", "", false, false, false, false, nil)
	ch2, _ := c.cc.Consume("single-handler", "", false, false, false, false, nil)

	for {
		select {
		case body := <-ch1:
			event := c.parse(body.Body)

			record, _ := inbox.NewRecord(event.ID, "multy-handlers", body.Body)

			log.Printf("write event from multy-handlers")

			writer.WriteInbox(context.Background(), record)

			body.Ack(false)
		case body := <-ch2:
			event := c.parse(body.Body)

			record, _ := inbox.NewRecord(event.ID, "single-handler", body.Body)

			log.Printf("write event from single-handler")

			writer.WriteInbox(context.Background(), record)

			body.Ack(false)
		}
	}
}

func (c *RandomConsumer) parse(body []byte) rawBody {
	var event rawBody

	if err := json.Unmarshal(body, &event); err != nil {
		log.Printf("%s", err)
	}

	return event
}

type firstMultiHandler struct{}

func (h *firstMultiHandler) Process(ctx context.Context, body []byte) error {
	log.Printf("firstMultyHandler: %s", body)

	return nil
}

func (h *firstMultiHandler) Key() string {
	return "fistMultyHandler"
}

type secondMultiHandler struct{}

func (h *secondMultiHandler) Process(ctx context.Context, body []byte) error {
	return errors.New("error")
}

func (h *secondMultiHandler) Key() string {
	return "secondMultiHandler"
}

type singleHandler struct{}

func (h *singleHandler) Process(ctx context.Context, body []byte) error {
	return errors.New("error")
}

func (h *singleHandler) Key() string {
	return "singleHandler"
}
