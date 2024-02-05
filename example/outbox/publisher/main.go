package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/Melenium2/go-iobox/outbox"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/streadway/amqp"
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

	broker, err := NewRandomPublisher(*useAmqpAddress)
	if err != nil {
		log.Fatal(err)
	}

	ob := outbox.NewOutbox(broker, db, outbox.EnableDebugMode(), outbox.EnableMetrics())

	if err = ob.Start(context.Background()); err != nil {
		log.Fatal(err)
	}

	cycle := NewInfinityPublishCycle(10*time.Second, ob.Writer(), db)

	cycle.Start()
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

type body struct {
	A int     `json:"a"`
	B float64 `json:"b"`
}

func (b *body) MarshalJSON() ([]byte, error) {
	type Alias body

	return json.Marshal(struct {
		*Alias
	}{
		Alias: (*Alias)(b),
	})
}

type InifinityPublishCycle struct {
	interval time.Duration
	client   outbox.Client
	db       *sql.DB
}

func NewInfinityPublishCycle(interval time.Duration, client outbox.Client, conn *sql.DB) *InifinityPublishCycle {
	return &InifinityPublishCycle{
		interval: interval,
		client:   client,
		db:       conn,
	}
}

func (p *InifinityPublishCycle) Start() {
	ticker := time.NewTicker(p.interval)

	for range ticker.C {
		ctx := context.Background()

		b := body{A: 1, B: float64(24)}

		rec := outbox.NewRecord(uuid.New(), "outbox", &b)

		err := p.client.WriteOutbox(ctx, p.db, rec)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("publish new record to outbox")
	}
}
