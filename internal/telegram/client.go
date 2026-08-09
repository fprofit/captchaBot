package telegram

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	appLog "github.com/fprofit/telegram-captcha-bot/internal/log"
)

type Client struct {
	token       string
	pollTimeout int
	client      *http.Client
	logger      *appLog.Logger
}

func NewClient(token string, pollTimeout int, logger *appLog.Logger) *Client {
	return &Client{
		token:       token,
		pollTimeout: pollTimeout,
		client: &http.Client{
			Timeout: time.Duration(pollTimeout+10) * time.Second,
		},
		logger: logger,
	}
}

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

func (c *Client) SendCaptchaMessage(chatID int64, text string, replyMarkup *InlineKeyboardMarkup) (Message, error) {
	params := url.Values{}

	params.Set("chat_id", strconv.FormatInt(chatID, 10))
	params.Set("text", text)
	params.Set("parse_mode", "HTML")

	if replyMarkup != nil {
		markup, err := json.Marshal(replyMarkup)
		if err != nil {
			return Message{}, fmt.Errorf(
				"encode reply markup: %w",
				err,
			)
		}

		params.Set("reply_markup", string(markup))
	}

	apiURL := fmt.Sprintf(
		"https://api.telegram.org/bot%s/sendMessage",
		c.token,
	)

	c.logger.Debugf(
		"Telegram API request: sendMessage chat_id=%d",
		chatID,
	)

	resp, err := c.client.PostForm(apiURL, params)
	if err != nil {
		return Message{}, fmt.Errorf(
			"sendMessage request: %w",
			err,
		)
	}

	defer resp.Body.Close()

	var result apiResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return Message{}, fmt.Errorf(
			"decode sendMessage response: %w",
			err,
		)
	}

	if !result.OK {
		return Message{}, fmt.Errorf(
			"Telegram API error %d: %s",
			result.ErrorCode,
			result.Description,
		)
	}

	var message Message

	if err := json.Unmarshal(result.Result, &message); err != nil {
		return Message{}, fmt.Errorf(
			"decode sendMessage result: %w",
			err,
		)
	}

	c.logger.Debugf(
		"Telegram API response: message_id=%d",
		message.MessageID,
	)

	return message, nil
}

func (c *Client) AnswerCallbackQuery(callbackID string, text string, showAlert bool) error {
	params := url.Values{}

	params.Set("callback_query_id", callbackID)
	params.Set("show_alert", strconv.FormatBool(showAlert))

	if text != "" {
		params.Set("text", text)
	}

	apiURL := fmt.Sprintf(
		"https://api.telegram.org/bot%s/answerCallbackQuery",
		c.token,
	)

	c.logger.Debugf(
		"Telegram API request: answerCallbackQuery id=%s", callbackID)

	resp, err := c.client.PostForm(apiURL, params)
	if err != nil {
		return fmt.Errorf("answerCallbackQuery request: %w", err)
	}

	defer resp.Body.Close()

	var result apiResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf(
			"decode answerCallbackQuery response: %w", err)
	}

	if !result.OK {
		return fmt.Errorf(
			"Telegram API error %d: %s",
			result.ErrorCode, result.Description)
	}

	return nil
}

func (c *Client) RemoveChatMember(chatID int64, userID int64) error {
	if err := c.banChatMember(chatID, userID); err != nil {
		return err
	}

	if err := c.unbanChatMember(chatID, userID); err != nil {
		return err
	}

	c.logger.Infof(
		"user removed from chat: chat_id=%d user_id=%d",
		chatID,
		userID,
	)

	return nil
}

func (c *Client) banChatMember(chatID int64, userID int64) error {
	params := url.Values{}

	params.Set(
		"chat_id",
		strconv.FormatInt(chatID, 10),
	)

	params.Set(
		"user_id",
		strconv.FormatInt(userID, 10),
	)

	apiURL := fmt.Sprintf(
		"https://api.telegram.org/bot%s/banChatMember",
		c.token,
	)

	c.logger.Debugf(
		"Telegram API request: banChatMember chat_id=%d user_id=%d",
		chatID,
		userID,
	)

	resp, err := c.client.PostForm(apiURL, params)
	if err != nil {
		return fmt.Errorf(
			"banChatMember request: %w",
			err,
		)
	}

	defer resp.Body.Close()

	var result apiResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf(
			"decode banChatMember response: %w",
			err,
		)
	}

	if !result.OK {
		return fmt.Errorf(
			"Telegram API error %d: %s",
			result.ErrorCode,
			result.Description,
		)
	}

	return nil
}

func (c *Client) unbanChatMember(chatID int64, userID int64) error {
	params := url.Values{}

	params.Set(
		"chat_id",
		strconv.FormatInt(chatID, 10),
	)

	params.Set(
		"user_id",
		strconv.FormatInt(userID, 10),
	)

	apiURL := fmt.Sprintf(
		"https://api.telegram.org/bot%s/unbanChatMember",
		c.token,
	)

	c.logger.Debugf(
		"Telegram API request: unbanChatMember chat_id=%d user_id=%d",
		chatID,
		userID,
	)

	resp, err := c.client.PostForm(apiURL, params)
	if err != nil {
		return fmt.Errorf(
			"unbanChatMember request: %w",
			err,
		)
	}

	defer resp.Body.Close()

	var result apiResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf(
			"decode unbanChatMember response: %w",
			err,
		)
	}

	if !result.OK {
		return fmt.Errorf(
			"Telegram API error %d: %s",
			result.ErrorCode,
			result.Description,
		)
	}

	return nil
}

func (c *Client) DeleteMessage(chatID int64, messageID int64) error {
	params := url.Values{}

	params.Set("chat_id", strconv.FormatInt(chatID, 10))
	params.Set("message_id", strconv.FormatInt(messageID, 10))

	apiURL := fmt.Sprintf(
		"https://api.telegram.org/bot%s/deleteMessage",
		c.token,
	)

	c.logger.Debugf(
		"Telegram API request: deleteMessage chat_id=%d message_id=%d",
		chatID,
		messageID,
	)

	resp, err := c.client.PostForm(apiURL, params)
	if err != nil {
		return fmt.Errorf(
			"deleteMessage request: %w",
			err,
		)
	}

	defer resp.Body.Close()

	var result apiResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf(
			"decode deleteMessage response: %w",
			err,
		)
	}

	if !result.OK {
		return fmt.Errorf(
			"Telegram API error %d: %s",
			result.ErrorCode,
			result.Description,
		)
	}

	return nil
}

func (c *Client) RestrictChatMember(chatID, userID int64) error {
	permissions := ChatPermissions{}

	data, err := json.Marshal(permissions)
	if err != nil {
		return fmt.Errorf("marshal chat permissions: %w", err)
	}

	params := url.Values{}
	params.Set("chat_id", strconv.FormatInt(chatID, 10))
	params.Set("user_id", strconv.FormatInt(userID, 10))
	params.Set("permissions", string(data))
	params.Set("use_independent_chat_permissions", "true")

	apiURL := fmt.Sprintf(
		"https://api.telegram.org/bot%s/restrictChatMember",
		c.token,
	)

	c.logger.Debugf(
		"Telegram API request: restrictChatMember chat_id=%d user_id=%d",
		chatID,
		userID,
	)

	resp, err := c.client.PostForm(apiURL, params)
	if err != nil {
		return fmt.Errorf(
			"restrictChatMember request: %w",
			err,
		)
	}

	defer resp.Body.Close()

	var result apiResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf(
			"decode restrictChatMember response: %w",
			err,
		)
	}

	if !result.OK {
		return fmt.Errorf(
			"Telegram API error %d: %s",
			result.ErrorCode,
			result.Description,
		)
	}

	return nil
}

func (c *Client) UnrestrictChatMember(chatID, userID int64) error {
	permissions := ChatPermissions{
		CanSendMessages:       true,
		CanSendAudios:         true,
		CanSendDocuments:      true,
		CanSendPhotos:         true,
		CanSendVideos:         true,
		CanSendVideoNotes:     true,
		CanSendVoiceNotes:     true,
		CanSendPolls:          true,
		CanSendOtherMessages:  true,
		CanAddWebPagePreviews: true,
		CanChangeInfo:         true,
		CanInviteUsers:        true,
		CanPinMessages:        true,
		CanManageTopics:       true,
		CanReactToMessages:    true,
	}

	data, err := json.Marshal(permissions)
	if err != nil {
		return fmt.Errorf("marshal chat permissions: %w", err)
	}

	params := url.Values{}
	params.Set("chat_id", strconv.FormatInt(chatID, 10))
	params.Set("user_id", strconv.FormatInt(userID, 10))
	params.Set("permissions", string(data))
	params.Set("use_independent_chat_permissions", "true")

	apiURL := fmt.Sprintf(
		"https://api.telegram.org/bot%s/restrictChatMember",
		c.token,
	)

	c.logger.Debugf(
		"Telegram API request: unrestrictChatMember chat_id=%d user_id=%d",
		chatID,
		userID,
	)

	resp, err := c.client.PostForm(apiURL, params)
	if err != nil {
		return fmt.Errorf(
			"unrestrictChatMember request: %w",
			err,
		)
	}

	defer resp.Body.Close()

	var result apiResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf(
			"decode unrestrictChatMember response: %w",
			err,
		)
	}

	if !result.OK {
		return fmt.Errorf(
			"Telegram API error %d: %s",
			result.ErrorCode,
			result.Description,
		)
	}

	return nil
}
