package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/scram"
	"golang.org/x/exp/slog"
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
	slog.Info(fmt.Sprintf("send message with content %s to topic %s", string(data), topic()))

	m := kafka.Message{
		Key:   []byte(key),
		Value: data,
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
	clientCertPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(KafkaCA())

	if err != nil {
		slog.Error("error reading ca cert", err)
		return nil, err
	}

	if ok := clientCertPool.AppendCertsFromPEM(ca); !ok {
		slog.Warn("error appending certs from pem")
	}

	mechanism, err := scram.Mechanism(scram.SHA256, KafkaUser(), KafkaPassword())
	if err != nil {
		slog.Error("error creating scram mechanism", err)
		return nil, err
	}

	cert, _ := tls.LoadX509KeyPair(KafkaCert(), KafkaKey())

	dialer := &kafka.Dialer{
		Timeout:       10 * time.Second,
		DualStack:     true,
		SASLMechanism: mechanism,
		TLS: &tls.Config{
			ClientCAs:          clientCertPool,
			InsecureSkipVerify: true,
			Certificates:       []tls.Certificate{cert},
		},
	}

	slog.Info(fmt.Sprintf("creating kafka writer with host %s, topic %s, tls %s %s %s", KafkaHost(), topic(), KafkaCA(), KafkaCert(), KafkaKey()))
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{KafkaHost()},
		Topic:   topic(),
		Dialer:  dialer,
	})
	writer.AllowAutoTopicCreation = true

	return writer, nil
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
		panic("TOPIC_DEV_SPAM_CHAT env var is not set")
	}

	return res
}

func pipelineSuccessful() bool {
	s := os.Getenv("SUCCESS")
	if s == "" {
		panic("SUCCESS env var is not set")
	}

	success, err := strconv.ParseBool(s)

	if err != nil {
		slog.Error("can not parse SUCCESS env var", err)
		panic("SUCCESS env var is not a boolean")
	}

	return success
}

func Env(key string) string {
	res := os.Getenv(key)
	if res == "" {
		panic(fmt.Sprintf("%s env var is not set", key))
	}

	return res
}

func KafkaHost() string {
	return Env("KAFKA_BROKERS")
}

func KafkaCA() string {
	return Env("KAFKA_TLS_CA_LOCATION")
}

func KafkaCert() string {
	return Env("KAFKA_TLS_CERTIFICATE_LOCATION")
}

func KafkaKey() string {
	return Env("KAFKA_TLS_KEY_LOCATION")
}

func KafkaUser() string {
	return Env("KAFKA_USERNAME")
}

func KafkaPassword() string {
	return Env("KAFKA_PASSWORD")
}
