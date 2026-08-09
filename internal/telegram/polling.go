package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
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

	c.logger.Debugf(
		"Telegram API request: getUpdates offset=%d timeout=%d",
		offset,
		c.pollTimeout,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create getUpdates request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		c.logger.Errorf(
			"getUpdates request failed: %v",
			err,
		)

		return nil, fmt.Errorf(
			"getUpdates request: %w",
			err,
		)
	}

	defer resp.Body.Close()

	var result apiResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.logger.Errorf(
			"decode getUpdates response failed: %v",
			err,
		)

		return nil, fmt.Errorf(
			"decode getUpdates response: %w",
			err,
		)
	}

	if !result.OK {
		c.logger.Errorf(
			"Telegram API error: %d %s",
			result.ErrorCode,
			result.Description,
		)

		return nil, fmt.Errorf(
			"Telegram API error %d: %s",
			result.ErrorCode,
			result.Description,
		)
	}

	var updates []Update

	if err := json.Unmarshal(result.Result, &updates); err != nil {
		c.logger.Errorf(
			"decode getUpdates result failed: %v",
			err,
		)

		return nil, fmt.Errorf(
			"decode getUpdates result: %w",
			err,
		)
	}

	c.logger.Debugf(
		"Telegram API response: received %d updates",
		len(updates),
	)

	return updates, nil
}

func (c *Client) Poll(ctx context.Context, handler func(Update)) error {
	offset := 0
	c.logger.Info("Telegram long polling started")
	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Telegram long polling stopped")
			return ctx.Err()
		default:
		}
		updates, err := c.GetUpdates(ctx, offset)
		if err != nil {
			c.logger.Errorf("getUpdates failed: %v", err)
			continue
		}
		for _, update := range updates {
			c.logger.Debugf("processing update: update_id=%d", update.UpdateID)
			handler(update)
			offset = update.UpdateID + 1
		}
	}
}
