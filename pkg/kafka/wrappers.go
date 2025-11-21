package kafka

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"time"

	"github.com/IBM/sarama"
)

// handlerWithRecovery is a wrapper around the handler to add recovery support.
func (c *Consumer) handlerWithRecovery(next HandleFunc) HandleFunc {
	return func(ctx context.Context, msg *sarama.ConsumerMessage) (err error) {
		defer func() {
			if r := recover(); r != nil {
				stackTrace := make([]byte, 4096) // 4KB
				stackTrace = stackTrace[:runtime.Stack(stackTrace, false)]

				slog.Error(
					"panic recovered in recovery handler",
					"stack_trace",
					string(stackTrace),
					"panic_values",
					fmt.Sprintf("%v", r),
				)
			}
		}()
		return next(ctx, msg)
	}
}

// handlerWithTimeout is a wrapper around the handler to add timeout support.
func (c *Consumer) handlerWithTimeout(next HandleFunc) HandleFunc {
	return func(ctx context.Context, msg *sarama.ConsumerMessage) error {
		ctx, cancel := context.WithTimeout(ctx, c.cfg.HandlerTimeout)
		defer cancel()

		return next(ctx, msg)
	}
}

// handleWithLogging is a wrapper around the handler to add logging.
func (c *Consumer) handlerWithLogging(next HandleFunc) HandleFunc {
	return func(ctx context.Context, msg *sarama.ConsumerMessage) error {
		start := time.Now()

		// extra recovery for catching panic in earler staps of the handler
		withRecovery := c.handlerWithRecovery(next)
		err := withRecovery(ctx, msg)

		duration := time.Since(start)

		headers := make(map[string]string, len(msg.Headers))
		for _, h := range msg.Headers {
			headers[string(h.Key)] = string(h.Value)
		}

		logger := slog.With(
			"topic", msg.Topic,
			"partition", msg.Partition,
			"offset", msg.Offset,
			"key", string(msg.Key),
			"duration", duration.String(),
			"headers", headers,
		)

		logMsg := "consumed incoming kafka message"
		if err != nil {
			logger = logger.With("consume_error", err.Error())
			logger.Error(logMsg)
		} else {
			logger.Info(logMsg)
		}

		return err
	}
}
