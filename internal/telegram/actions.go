package telegram

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

func (c *Client) SendMessage(chatID int64, text string) error {
	params := url.Values{}
	params.Set("chat_id", strconv.FormatInt(chatID, 10))
	params.Set("text", text)
	params.Set("parse_mode", "HTML")

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", c.token)

	c.logger.Debug("Telegram API request", "method", "sendMessage", "chat_id", chatID)

	resp, err := c.client.PostForm(apiURL, params)
	if err != nil {
		return fmt.Errorf("sendMessage request: %w", err)
	}

	defer resp.Body.Close()

	var result apiResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("decode sendMessage response: %w", err)
	}

	if !result.OK {
		return fmt.Errorf("Telegram API error %d: %s", result.ErrorCode, result.Description)
	}

	c.logger.Debug("Telegram API response", "method", "sendMessage", "chat_id", chatID)

	return nil
}
