package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
	"os"
	"strconv"
	"time"
)

type Message struct {
	Message string
	Channel string
}

const (
	CI_SUCCESS_CHANNEL = "ci-success"
	CI_FAILURE_CHANNEL = "ci-failure"
)

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
	log.Info().Msgf("send message with content %s to topic %s", string(data), topic())

	m := kafka.Message{
		Key:   []byte(key),
		Value: data,
		Topic: topic(),
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

func message(c *Config) Message {

	prefix := "[SUCCESS]"
	if !pipelineSuccessful() {
		prefix = "[FAILURE]"
	}

	text := fmt.Sprintf("%s %s/%s updated", prefix, c.Project, c.Organization)

	m := Message{
		Message: text,
		Channel: channel(),
	}

	return m
}

func channel() string {
	if pipelineSuccessful() {
		return CI_SUCCESS_CHANNEL
	}
	return CI_FAILURE_CHANNEL
}

func topic() string {

	res := os.Getenv("TOPIC_DEV_CHAT")
	if res == "" {
		log.Panic().Msgf("TOPIC_DEV_SPAM_CHAT env var is not set")
	}

	return res
}

func pipelineSuccessful() bool {
	s := os.Getenv("SUCCESS")
	if s == "" {
		log.Panic().Msgf("SUCCESS env var is not set")
	}

	success, err := strconv.ParseBool(s)

	if err != nil {
		log.Panic().Err(err).Msgf("SUCCESS env var is not a boolean")
	}

	return success
}
