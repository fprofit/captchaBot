package telegram

import (
	"encoding/json"
	"fmt"
)

func (c *Client) GetMe() (User, error) {
	url := fmt.Sprintf(
		"https://api.telegram.org/bot%s/getMe",
		c.token,
	)

	c.logger.Debugf("Telegram API request: getMe")

	resp, err := c.client.Get(url)
	if err != nil {
		c.logger.Errorf("getMe request failed: %v", err)
		return User{}, fmt.Errorf("getMe request: %w", err)
	}
	defer resp.Body.Close()

	var result apiResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.logger.Errorf("decode getMe response failed: %v", err)
		return User{}, fmt.Errorf("decode getMe response: %w", err)
	}

	if !result.OK {
		c.logger.Errorf(
			"Telegram API error: %d %s",
			result.ErrorCode,
			result.Description,
		)

		return User{}, fmt.Errorf(
			"Telegram API error %d: %s",
			result.ErrorCode,
			result.Description,
		)
	}

	var user User

	if err := json.Unmarshal(result.Result, &user); err != nil {
		c.logger.Errorf("decode getMe result failed: %v", err)
		return User{}, fmt.Errorf("decode getMe result: %w", err)
	}

	c.logger.Debugf(
		"Telegram bot authenticated: @%s",
		user.Username,
	)

	return user, nil
}
