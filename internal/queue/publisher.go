package queue

import (
	"context"
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/zishan044/url-shortener/internal/models"
)

const (
	ClicksExchange = "url.clicks"
	ClicksQueue    = "clicks.analytics"
	ClicksKey      = "click.created"
)

type Publisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewPublisher(amqpURL string) (*Publisher, error) {
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
		ClicksExchange,
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
		ClicksQueue,
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
		ClicksQueue,
		ClicksKey,
		ClicksExchange,
		false,
		nil,
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, err
	}

	return &Publisher{
		conn:    conn,
		channel: channel,
	}, nil
}

func (p *Publisher) PublishClick(ctx context.Context, click *models.Click) error {
	body, err := json.Marshal(click)
	if err != nil {
		return err
	}

	return p.channel.PublishWithContext(
		ctx,
		ClicksExchange,
		ClicksKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

func (p *Publisher) Close() error {
	if p.channel != nil {
		if err := p.channel.Close(); err != nil {
			log.Printf("Error closing channel: %v", err)
		}
	}
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}
