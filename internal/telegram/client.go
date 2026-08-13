package telegram

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Client struct {
	token       string
	pollTimeout int
	client      *http.Client
	logger      *slog.Logger
}

func NewClient(token string, pollTimeout int, logger *slog.Logger) *Client {
	return &Client{
		token:       token,
		pollTimeout: pollTimeout,
		client: &http.Client{
			Timeout: time.Duration(pollTimeout+15) * time.Second,
		},
		logger: logger,
	}
}

func (c *Client) GetMe() (User, error) {
	apiURL := fmt.Sprintf(
		"https://api.telegram.org/bot%s/getMe",
		c.token,
	)

	c.logger.Debug("Telegram API request", "method", "getMe")

	resp, err := c.client.Get(apiURL)
	if err != nil {
		c.logger.Error(
			"getMe request failed",
			"err", err,
		)

		return User{}, fmt.Errorf("getMe request: %w", err)
	}

	defer resp.Body.Close()

	c.logger.Debug(
		"Telegram API response",
		"method", "getMe",
		"status", resp.StatusCode,
	)

	var result apiResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.logger.Error(
			"failed to decode getMe response",
			"err", err,
			"status", resp.StatusCode,
		)

		return User{}, fmt.Errorf("decode getMe response: %w", err)
	}

	if !result.OK {
		c.logger.Error(
			"Telegram API error",
			"method", "getMe",
			"error_code", result.ErrorCode,
			"description", result.Description,
		)

		return User{}, fmt.Errorf(
			"Telegram API error %d: %s",
			result.ErrorCode,
			result.Description,
		)
	}

	var user User

	if err := json.Unmarshal(result.Result, &user); err != nil {
		c.logger.Error(
			"failed to decode getMe result",
			"err", err,
		)

		return User{}, fmt.Errorf("decode getMe result: %w", err)
	}

	c.logger.Info(
		"Telegram bot authenticated",
		"username", user.Username,
		"id", user.ID,
	)

	return user, nil
}

func (c *Client) AnswerCallbackQuery(callbackID string, text string, showAlert bool, cacheTime int) error {
	params := url.Values{}

	params.Set("callback_query_id", callbackID)
	params.Set("show_alert", strconv.FormatBool(showAlert))
	params.Set("cache_time", strconv.Itoa(cacheTime))

	if text != "" {
		params.Set("text", text)
	}

	apiURL := fmt.Sprintf(
		"https://api.telegram.org/bot%s/answerCallbackQuery",
		c.token,
	)

	c.logger.Debug("Telegram API request", "method", "answerCallbackQuery", "callback_id", callbackID, "show_alert", showAlert)

	resp, err := c.client.PostForm(apiURL, params)
	if err != nil {
		c.logger.Error("answerCallbackQuery request failed", "err", err, "callback_id", callbackID)
		return fmt.Errorf("answerCallbackQuery request: %w", err)
	}

	defer resp.Body.Close()

	var result apiResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.logger.Error("failed to decode answerCallbackQuery response", "err", err, "callback_id", callbackID)
		return fmt.Errorf("decode answerCallbackQuery response: %w", err)
	}

	if !result.OK {
		c.logger.Error("Telegram API error", "method", "answerCallbackQuery", "callback_id", callbackID, "error_code", result.ErrorCode, "description", result.Description)

		return fmt.Errorf("Telegram API error %d: %s", result.ErrorCode, result.Description)
	}

	return nil
}

func (c *Client) DeleteMessage(chatID, messageID int64) error {
	params := url.Values{}
	params.Set("chat_id", strconv.FormatInt(chatID, 10))
	params.Set("message_id", strconv.FormatInt(messageID, 10))

	apiURL := fmt.Sprintf(
		"https://api.telegram.org/bot%s/deleteMessage",
		c.token,
	)

	c.logger.Debug("Telegram API request", "method", "deleteMessage", "chat_id", chatID, "message_id", messageID)

	resp, err := c.client.PostForm(apiURL, params)
	if err != nil {
		c.logger.Error("deleteMessage request failed", "err", err, "chat_id", chatID, "message_id", messageID)
		return fmt.Errorf("deleteMessage request: %w", err)
	}

	defer resp.Body.Close()

	var result apiResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.logger.Error("failed to decode deleteMessage response", "err", err, "chat_id", chatID, "message_id", messageID)
		return fmt.Errorf("decode deleteMessage response: %w", err)
	}

	if !result.OK {
		c.logger.Error("Telegram API error", "method", "deleteMessage", "chat_id", chatID, "message_id", messageID, "error_code", result.ErrorCode, "description", result.Description)
		return fmt.Errorf("Telegram API error %d: %s", result.ErrorCode, result.Description)
	}

	return nil
}
