package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

func (c *Client) GetUpdates(ctx context.Context, offset int) ([]Update, error) {
	params := url.Values{}
	params.Set("offset", strconv.Itoa(offset))
	params.Set("timeout", strconv.Itoa(c.pollTimeout))

	apiURL := fmt.Sprintf(
		"https://api.telegram.org/bot%s/getUpdates?%s",
		c.token,
		params.Encode(),
	)

	c.logger.Debug(
		"Telegram API request",
		"method", http.MethodGet,
		"endpoint", "getUpdates",
		"offset", offset,
		"timeout", c.pollTimeout,
	)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		apiURL,
		nil,
	)
	if err != nil {
		c.logger.Error(
			"failed to create getUpdates request",
			"err", err,
			"offset", offset,
		)

		return nil, fmt.Errorf("create getUpdates request: %w", err)
	}

	start := time.Now()

	resp, err := c.client.Do(req)
	if err != nil {
		c.logger.Error(
			"getUpdates request failed",
			"err", err,
			"offset", offset,
			"duration", time.Since(start),
		)

		return nil, fmt.Errorf("getUpdates request: %w", err)
	}

	defer resp.Body.Close()

	c.logger.Debug(
		"Telegram API response received",
		"endpoint", "getUpdates",
		"status", resp.StatusCode,
		"duration", time.Since(start),
	)

	var result apiResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.logger.Error(
			"failed to decode getUpdates response",
			"err", err,
			"status", resp.StatusCode,
			"duration", time.Since(start),
		)

		return nil, fmt.Errorf(
			"decode getUpdates response: %w",
			err,
		)
	}

	if !result.OK {
		c.logger.Error(
			"Telegram API returned an error",
			"error_code", result.ErrorCode,
			"description", result.Description,
			"offset", offset,
			"duration", time.Since(start),
		)

		return nil, fmt.Errorf(
			"Telegram API error %d: %s",
			result.ErrorCode,
			result.Description,
		)
	}

	var updates []Update

	if err := json.Unmarshal(result.Result, &updates); err != nil {
		c.logger.Error(
			"failed to decode getUpdates result",
			"err", err,
			"duration", time.Since(start),
		)

		return nil, fmt.Errorf(
			"decode getUpdates result: %w",
			err,
		)
	}

	c.logger.Debug(
		"Telegram API getUpdates completed",
		"updates", len(updates),
		"next_offset", nextOffset(offset, updates),
		"duration", time.Since(start),
	)

	return updates, nil
}

func (c *Client) Poll(ctx context.Context, handler func(Update)) error {
	offset := 0

	c.logger.Info("Telegram long polling started")

	for {
		select {
		case <-ctx.Done():
			c.logger.Info(
				"Telegram long polling stopped",
				"reason", ctx.Err(),
			)

			return ctx.Err()

		default:
		}

		updates, err := c.GetUpdates(ctx, offset)
		if err != nil {
			if ctx.Err() != nil {
				c.logger.Debug(
					"getUpdates interrupted by context",
					"err", ctx.Err(),
				)

				return ctx.Err()
			}

			c.logger.Error(
				"getUpdates failed",
				"err", err,
				"offset", offset,
			)

			// Не крутим цикл слишком быстро при постоянной ошибке.
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Second):
			}

			continue
		}

		if len(updates) == 0 {
			c.logger.Debug(
				"no Telegram updates received",
				"offset", offset,
			)

			continue
		}

		c.logger.Debug(
			"Telegram updates received",
			"count", len(updates),
			"offset", offset,
		)

		for _, update := range updates {
			c.logger.Debug(
				"processing Telegram update",
				"update_id", update.UpdateID,
			)

			handler(update)

			offset = update.UpdateID + 1

			c.logger.Debug(
				"Telegram update processed",
				"update_id", update.UpdateID,
				"next_offset", offset,
			)
		}
	}
}

func nextOffset(offset int, updates []Update) int {
	if len(updates) == 0 {
		return offset
	}

	return updates[len(updates)-1].UpdateID + 1
}
