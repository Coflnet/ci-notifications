package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"os"
	"time"
)

type Message struct {
	Message string
}

func writeMessage(c *Config) error {
	w, err := writer()
	if err != nil {
		return err
	}

	msg := message(c)
	key := fmt.Sprintf("%s-%s", c.Project, c.Organization)

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	m := kafka.Message{
		Key:   []byte(key),
		Value: data,
		Topic: os.Getenv("TOPIC_DEV_CHAT"),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = w.WriteMessages(ctx, m)
	if err != nil {
		return err
	}

	return nil
}

func writer() (*kafka.Writer, error) {
	kHosts := os.Getenv("KAFKA_HOST")
	if kHosts == "" {
		return nil, fmt.Errorf("KAFKA_HOST env var is not set")
	}

	t := os.Getenv("TOPIC_DEV_CHAT")
	if t == "" {
		return nil, fmt.Errorf("TOPIC_DEV_CHAT env var is not set")
	}

	w := &kafka.Writer{
		Addr:                   kafka.TCP(kHosts),
		Topic:                  os.Getenv(t),
		AllowAutoTopicCreation: true,
	}

	return w, nil
}

func message(c *Config) *Message {
	m := &Message{
		Message: fmt.Sprintf("project %s in organization %s was updated at %v", c.Project, c.Organization, time.Now()),
	}

	return m
}
