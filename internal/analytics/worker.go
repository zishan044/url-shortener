package analytics

import (
	"context"
	"encoding/json"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/zishan044/url-shortener/internal/models"
	"github.com/zishan044/url-shortener/internal/queue"
)

type Worker struct {
	repo    Repository
	channel *amqp.Channel
	conn    *amqp.Connection
}

func NewWorker(amqpURL string, repo Repository) (*Worker, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, err
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	err = channel.ExchangeDeclare(
		queue.ClicksExchange,
		amqp.ExchangeTopic,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, err
	}

	_, err = channel.QueueDeclare(
		queue.ClicksQueue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, err
	}

	err = channel.QueueBind(
		queue.ClicksQueue,
		queue.ClicksKey,
		queue.ClicksExchange,
		false,
		nil,
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, err
	}

	return &Worker{
		repo:    repo,
		channel: channel,
		conn:    conn,
	}, nil
}

func (w *Worker) Start(ctx context.Context) error {
	err := w.channel.Qos(100, 0, false)
	if err != nil {
		return err
	}

	deliveries, err := w.channel.Consume(
		queue.ClicksQueue,
		"analytics-worker",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	log.Println("Analytics worker started, listening for click events...")

	batch := make([]*models.Click, 0, 100)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			if len(batch) > 0 {
				w.processBatch(ctx, batch)
			}
			w.Close()
			return ctx.Err()

		case <-ticker.C:
			if len(batch) > 0 {
				w.processBatch(ctx, batch)
				batch = make([]*models.Click, 0, 100)
			}

		case delivery := <-deliveries:
			if delivery.Body == nil {
				continue
			}

			var click models.Click
			err := json.Unmarshal(delivery.Body, &click)
			if err != nil {
				log.Printf("Failed to unmarshal click event: %v", err)
				delivery.Ack(false)
				continue
			}

			batch = append(batch, &click)

			if len(batch) >= 100 {
				w.processBatch(ctx, batch)
				batch = make([]*models.Click, 0, 100)
			}

			delivery.Ack(false)
		}
	}
}

func (w *Worker) processBatch(ctx context.Context, clicks []*models.Click) {
	err := w.repo.CreateClicksBatch(ctx, clicks)
	if err != nil {
		log.Printf("Failed to process batch of %d clicks: %v", len(clicks), err)
		return
	}
	log.Printf("Successfully processed batch of %d clicks", len(clicks))
}

func (w *Worker) Close() error {
	if w.channel != nil {
		if err := w.channel.Close(); err != nil {
			log.Printf("Error closing channel: %v", err)
		}
	}
	if w.conn != nil {
		return w.conn.Close()
	}
	return nil
}
