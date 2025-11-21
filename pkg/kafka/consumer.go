package kafka

import (
	"chatx-01-backend/pkg/errs"
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/IBM/sarama"
)

type Consumer struct {
	cfg            ConsumerConfig
	topic          string
	serviceName    string
	serviceVersion string
	saramaCfg      *sarama.Config
	consumerGroup  sarama.ConsumerGroup
	handleFn       HandleFunc
}

// HandleFunc is a delivery handler that should be injected into the consumer.
type HandleFunc func(context.Context, *sarama.ConsumerMessage) error

// NewConsumer creates a new kafka consumer.
func NewConsumer(
	cfg ConsumerConfig,
	topic string,
	serviceName string,
	serviceVersion string,
	handleFn HandleFunc,
) (*Consumer, error) {
	const op = "kafka.NewConsumer"

	saramaCfg, err := cfg.getSaramaConfig(serviceName)
	if err != nil {
		return nil, errs.Wrap(op, err)
	}

	// create a new consumer group
	consumerGroup, err := sarama.NewConsumerGroup(strings.Split(cfg.Brokers, ","), cfg.GroupID, saramaCfg)
	if err != nil {
		return nil, errs.Wrap(op, err)
	}

	return &Consumer{
		cfg:            cfg,
		topic:          topic,
		serviceName:    serviceName,
		serviceVersion: serviceVersion,
		saramaCfg:      saramaCfg,
		consumerGroup:  consumerGroup,
		handleFn:       handleFn,
	}, nil
}

// Start starts the consumer and begins consuming messages.
func (c *Consumer) Start() error {
	const op = "kafka.Consumer.Start"

	// the main consume loop, parent of the ConsumerClaim() partition consumer loop
	for {
		err := c.consumerGroup.Consume(context.Background(), []string{c.topic}, c)
		if err != nil {
			if errors.Is(err, sarama.ErrClosedConsumerGroup) {
				return nil
			}
			return errs.Wrap(op, err)
		}

		slog.Info("rebalancing occurred, waiting for new messages")
	}
}

func (c *Consumer) Stop() error {
	const op = "kafka.Consumer.Stop"

	if err := c.consumerGroup.Close(); err != nil {
		return errs.Wrap(op, err)
	}
	return nil
}

// Setup implements sarama.ConsumerGroupHandler contract.
func (c *Consumer) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup implements sarama.ConsumerGroupHandler contract.
func (c *Consumer) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine,
	// https://github.com/IBM/sarama/blob/main/consumer_group.go#L27-L29
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				// The channel is closed, exit the loop
				return nil
			}

			// Build the handler chain
			chain := c.buildHandlerChain()

			// ignore the error and move on to the next message
			// as the error is already handled in the handler chain
			_ = chain(context.Background(), message)

			// mask this message offset as consumed
			session.MarkMessage(message, "")

		// Should return when `session.Context()` is done
		// if not, will raise `ErrRebalanceInProgress` or `read tcp <ip>:<port>: i/o timeout` when kafka rebalance
		// https://github.com/IBM/sarama/issues/1192
		case <-session.Context().Done():
			return nil
		}
	}
}

func (c *Consumer) buildHandlerChain() HandleFunc {
	// start with the core business logic handler
	handler := c.handleFn

	// build the chain in reverse order (last wrapper first)
	handler = c.handlerWithLogging(handler)
	handler = c.handlerWithTimeout(handler)
	handler = c.handlerWithRecovery(handler)

	return handler
}
