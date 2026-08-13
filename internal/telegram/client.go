package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"log/slog"
	"mime/multipart"
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

func (c *Client) SendCaptchaMessage(
	chatID int64,
	text string,
	replyMarkup *InlineKeyboardMarkup,
) (Message, error) {
	params := url.Values{}

	params.Set("chat_id", strconv.FormatInt(chatID, 10))
	params.Set("text", text)
	params.Set("parse_mode", "HTML")

	if replyMarkup != nil {
		markup, err := json.Marshal(replyMarkup)
		if err != nil {
			c.logger.Error(
				"failed to encode reply markup",
				"err", err,
				"chat_id", chatID,
			)

			return Message{}, fmt.Errorf("encode reply markup: %w", err)
		}

		params.Set("reply_markup", string(markup))
	}

	apiURL := fmt.Sprintf(
		"https://api.telegram.org/bot%s/sendMessage",
		c.token,
	)

	c.logger.Debug(
		"Telegram API request",
		"method", "sendMessage",
		"chat_id", chatID,
	)

	resp, err := c.client.PostForm(apiURL, params)
	if err != nil {
		c.logger.Error(
			"sendMessage request failed",
			"err", err,
			"chat_id", chatID,
		)

		return Message{}, fmt.Errorf(
			"sendMessage request: %w",
			err,
		)
	}

	defer resp.Body.Close()

	var result apiResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.logger.Error(
			"failed to decode sendMessage response",
			"err", err,
			"chat_id", chatID,
			"status", resp.StatusCode,
		)

		return Message{}, fmt.Errorf(
			"decode sendMessage response: %w",
			err,
		)
	}

	if !result.OK {
		c.logger.Error(
			"Telegram API error",
			"method", "sendMessage",
			"chat_id", chatID,
			"error_code", result.ErrorCode,
			"description", result.Description,
		)

		return Message{}, fmt.Errorf(
			"Telegram API error %d: %s",
			result.ErrorCode,
			result.Description,
		)
	}

	var message Message

	if err := json.Unmarshal(result.Result, &message); err != nil {
		c.logger.Error(
			"failed to decode sendMessage result",
			"err", err,
			"chat_id", chatID,
		)

		return Message{}, fmt.Errorf(
			"decode sendMessage result: %w",
			err,
		)
	}

	c.logger.Debug(
		"Telegram message sent",
		"chat_id", chatID,
		"message_id", message.MessageID,
	)

	return message, nil
}

func (c *Client) AnswerCallbackQuery(
	callbackID string,
	text string,
	showAlert bool,
) error {
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

	c.logger.Debug(
		"Telegram API request",
		"method", "answerCallbackQuery",
		"callback_id", callbackID,
		"show_alert", showAlert,
	)

	resp, err := c.client.PostForm(apiURL, params)
	if err != nil {
		c.logger.Error(
			"answerCallbackQuery request failed",
			"err", err,
			"callback_id", callbackID,
		)

		return fmt.Errorf("answerCallbackQuery request: %w", err)
	}

	defer resp.Body.Close()

	var result apiResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.logger.Error(
			"failed to decode answerCallbackQuery response",
			"err", err,
			"callback_id", callbackID,
		)

		return fmt.Errorf(
			"decode answerCallbackQuery response: %w",
			err,
		)
	}

	if !result.OK {
		c.logger.Error(
			"Telegram API error",
			"method", "answerCallbackQuery",
			"callback_id", callbackID,
			"error_code", result.ErrorCode,
			"description", result.Description,
		)

		return fmt.Errorf(
			"Telegram API error %d: %s",
			result.ErrorCode,
			result.Description,
		)
	}

	return nil
}

func (c *Client) RemoveChatMember(chatID, userID int64) error {
	if err := c.banChatMember(chatID, userID); err != nil {
		return err
	}

	if err := c.unbanChatMember(chatID, userID); err != nil {
		return err
	}

	c.logger.Info(
		"user removed from chat",
		"chat_id", chatID,
		"user_id", userID,
	)

	return nil
}

func (c *Client) banChatMember(chatID, userID int64) error {
	params := url.Values{}

	params.Set("chat_id", strconv.FormatInt(chatID, 10))
	params.Set("user_id", strconv.FormatInt(userID, 10))

	apiURL := fmt.Sprintf(
		"https://api.telegram.org/bot%s/banChatMember",
		c.token,
	)

	c.logger.Debug(
		"Telegram API request",
		"method", "banChatMember",
		"chat_id", chatID,
		"user_id", userID,
	)

	resp, err := c.client.PostForm(apiURL, params)
	if err != nil {
		c.logger.Error(
			"banChatMember request failed",
			"err", err,
			"chat_id", chatID,
			"user_id", userID,
		)

		return fmt.Errorf("banChatMember request: %w", err)
	}

	defer resp.Body.Close()

	var result apiResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.logger.Error(
			"failed to decode banChatMember response",
			"err", err,
			"chat_id", chatID,
			"user_id", userID,
		)

		return fmt.Errorf(
			"decode banChatMember response: %w",
			err,
		)
	}

	if !result.OK {
		c.logger.Error(
			"Telegram API error",
			"method", "banChatMember",
			"chat_id", chatID,
			"user_id", userID,
			"error_code", result.ErrorCode,
			"description", result.Description,
		)

		return fmt.Errorf(
			"Telegram API error %d: %s",
			result.ErrorCode,
			result.Description,
		)
	}

	return nil
}

func (c *Client) unbanChatMember(chatID, userID int64) error {
	params := url.Values{}

	params.Set("chat_id", strconv.FormatInt(chatID, 10))
	params.Set("user_id", strconv.FormatInt(userID, 10))

	apiURL := fmt.Sprintf(
		"https://api.telegram.org/bot%s/unbanChatMember",
		c.token,
	)

	c.logger.Debug(
		"Telegram API request",
		"method", "unbanChatMember",
		"chat_id", chatID,
		"user_id", userID,
	)

	resp, err := c.client.PostForm(apiURL, params)
	if err != nil {
		c.logger.Error(
			"unbanChatMember request failed",
			"err", err,
			"chat_id", chatID,
			"user_id", userID,
		)

		return fmt.Errorf("unbanChatMember request: %w", err)
	}

	defer resp.Body.Close()

	var result apiResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.logger.Error(
			"failed to decode unbanChatMember response",
			"err", err,
			"chat_id", chatID,
			"user_id", userID,
		)

		return fmt.Errorf(
			"decode unbanChatMember response: %w",
			err,
		)
	}

	if !result.OK {
		c.logger.Error(
			"Telegram API error",
			"method", "unbanChatMember",
			"chat_id", chatID,
			"user_id", userID,
			"error_code", result.ErrorCode,
			"description", result.Description,
		)

		return fmt.Errorf(
			"Telegram API error %d: %s",
			result.ErrorCode,
			result.Description,
		)
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

	c.logger.Debug(
		"Telegram API request",
		"method", "deleteMessage",
		"chat_id", chatID,
		"message_id", messageID,
	)

	resp, err := c.client.PostForm(apiURL, params)
	if err != nil {
		c.logger.Error(
			"deleteMessage request failed",
			"err", err,
			"chat_id", chatID,
			"message_id", messageID,
		)

		return fmt.Errorf("deleteMessage request: %w", err)
	}

	defer resp.Body.Close()

	var result apiResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.logger.Error(
			"failed to decode deleteMessage response",
			"err", err,
			"chat_id", chatID,
			"message_id", messageID,
		)

		return fmt.Errorf(
			"decode deleteMessage response: %w",
			err,
		)
	}

	if !result.OK {
		c.logger.Error(
			"Telegram API error",
			"method", "deleteMessage",
			"chat_id", chatID,
			"message_id", messageID,
			"error_code", result.ErrorCode,
			"description", result.Description,
		)

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
		c.logger.Error(
			"failed to marshal chat permissions",
			"err", err,
			"chat_id", chatID,
			"user_id", userID,
		)

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

	c.logger.Debug(
		"Telegram API request",
		"method", "restrictChatMember",
		"chat_id", chatID,
		"user_id", userID,
	)

	resp, err := c.client.PostForm(apiURL, params)
	if err != nil {
		c.logger.Error(
			"restrictChatMember request failed",
			"err", err,
			"chat_id", chatID,
			"user_id", userID,
		)

		return fmt.Errorf("restrictChatMember request: %w", err)
	}

	defer resp.Body.Close()

	var result apiResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.logger.Error(
			"failed to decode restrictChatMember response",
			"err", err,
			"chat_id", chatID,
			"user_id", userID,
		)

		return fmt.Errorf(
			"decode restrictChatMember response: %w",
			err,
		)
	}

	if !result.OK {
		c.logger.Error(
			"Telegram API error",
			"method", "restrictChatMember",
			"chat_id", chatID,
			"user_id", userID,
			"error_code", result.ErrorCode,
			"description", result.Description,
		)

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
		c.logger.Error(
			"failed to marshal chat permissions",
			"err", err,
			"chat_id", chatID,
			"user_id", userID,
		)

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

	c.logger.Debug("Telegram API request", "method", "restrictChatMember", "action", "unrestrict", "chat_id", chatID, "user_id", userID)

	resp, err := c.client.PostForm(apiURL, params)
	if err != nil {
		c.logger.Error("unrestrictChatMember request failed", "err", err, "chat_id", chatID, "user_id", userID)

		return fmt.Errorf("unrestrictChatMember request: %w", err)
	}

	defer resp.Body.Close()

	var result apiResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.logger.Error("failed to decode unrestrictChatMember response", "err", err, "chat_id", chatID, "user_id", userID)

		return fmt.Errorf("decode unrestrictChatMember response: %w", err)
	}

	if !result.OK {
		c.logger.Error("Telegram API error", "method", "restrictChatMember", "action", "unrestrict", "chat_id", chatID, "user_id", userID, "error_code", result.ErrorCode, "description", result.Description)
		return fmt.Errorf("Telegram API error %d: %s", result.ErrorCode, result.Description)
	}

	return nil
}

func (c *Client) SendCaptchaMessageImg(chatID int64, text string, img image.Image, replyMarkup *InlineKeyboardMarkup) (Message, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	if err := writer.WriteField("chat_id", strconv.FormatInt(chatID, 10)); err != nil {
		c.logger.Error("failed to write chat_id", "err", err, "chat_id", chatID)

		return Message{}, fmt.Errorf("write chat_id: %w", err)
	}

	if err := writer.WriteField("caption", text); err != nil {
		c.logger.Error("failed to write caption", "err", err, "chat_id", chatID)

		return Message{}, fmt.Errorf("write caption: %w", err)
	}

	if err := writer.WriteField("parse_mode", "HTML"); err != nil {
		c.logger.Error("failed to write parse mode", "err", err, "chat_id", chatID)

		return Message{}, fmt.Errorf("write parse_mode: %w", err)
	}

	if replyMarkup != nil {
		markup, err := json.Marshal(replyMarkup)
		if err != nil {
			c.logger.Error("failed to encode reply markup", "err", err, "chat_id", chatID)

			return Message{}, fmt.Errorf("encode reply markup: %w", err)
		}

		if err := writer.WriteField("reply_markup", string(markup)); err != nil {
			c.logger.Error("failed to write reply markup", "err", err, "chat_id", chatID)

			return Message{}, fmt.Errorf("write reply_markup: %w", err)
		}
	}

	part, err := writer.CreateFormFile("photo", "captcha.png")
	if err != nil {
		c.logger.Error("failed to create photo field", "err", err, "chat_id", chatID)

		return Message{}, fmt.Errorf("create photo field: %w", err)
	}

	if err := png.Encode(part, img); err != nil {
		c.logger.Error("failed to encode captcha image", "err", err, "chat_id", chatID)

		return Message{}, fmt.Errorf("encode captcha image: %w", err)
	}

	if err := writer.Close(); err != nil {
		c.logger.Error("failed to close multipart writer", "err", err, "chat_id", chatID)

		return Message{}, fmt.Errorf("close multipart writer: %w", err)
	}

	apiURL := fmt.Sprintf(
		"https://api.telegram.org/bot%s/sendPhoto",
		c.token,
	)

	c.logger.Debug(
		"Telegram API request",
		"method", "sendPhoto",
		"chat_id", chatID,
	)

	req, err := http.NewRequest(http.MethodPost, apiURL, &body)
	if err != nil {
		c.logger.Error("failed to create sendPhoto request", "err", err, "chat_id", chatID)

		return Message{}, fmt.Errorf(
			"create sendPhoto request: %w",
			err,
		)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.client.Do(req)
	if err != nil {
		c.logger.Error("sendPhoto request failed", "err", err, "chat_id", chatID)

		return Message{}, fmt.Errorf("sendPhoto request: %w", err)
	}

	defer resp.Body.Close()

	var result apiResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.logger.Error("failed to decode sendPhoto response", "err", err, "chat_id", chatID, "status", resp.StatusCode)

		return Message{}, fmt.Errorf(
			"decode sendPhoto response: %w",
			err,
		)
	}

	if !result.OK {
		c.logger.Error("Telegram API error", "method", "sendPhoto", "chat_id", chatID, "error_code", result.ErrorCode, "description", result.Description)

		return Message{}, fmt.Errorf(
			"Telegram API error %d: %s",
			result.ErrorCode,
			result.Description,
		)
	}

	var message Message

	if err := json.Unmarshal(result.Result, &message); err != nil {
		c.logger.Error("failed to decode sendPhoto result", "err", err, "chat_id", chatID)

		return Message{}, fmt.Errorf(
			"decode sendPhoto result: %w",
			err,
		)
	}

	c.logger.Debug("Telegram photo sent", "chat_id", chatID, "message_id", message.MessageID)

	return message, nil
}

func (c *Client) EditCaptchaMessageImg(chatID int64, messageID int64, img image.Image, replyMarkup *InlineKeyboardMarkup) (Message, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	if err := writer.WriteField("chat_id", strconv.FormatInt(chatID, 10)); err != nil {
		c.logger.Error("failed to write chat_id", "err", err, "chat_id", chatID, "message_id", messageID)

		return Message{}, fmt.Errorf("write chat_id: %w", err)
	}

	if err := writer.WriteField("message_id", strconv.FormatInt(messageID, 10)); err != nil {
		c.logger.Error("failed to write message_id", "err", err, "chat_id", chatID, "message_id", messageID)

		return Message{}, fmt.Errorf("write message_id: %w", err)
	}

	media := map[string]interface{}{
		"type":       "photo",
		"media":      "attach://captcha.png",
		"parse_mode": "HTML",
	}

	mediaJSON, err := json.Marshal(media)
	if err != nil {
		c.logger.Error("failed to encode media", "err", err, "chat_id", chatID, "message_id", messageID)

		return Message{}, fmt.Errorf("encode media: %w", err)
	}

	if err := writer.WriteField("media", string(mediaJSON)); err != nil {
		c.logger.Error("failed to write media", "err", err, "chat_id", chatID, "message_id", messageID)

		return Message{}, fmt.Errorf("write media: %w", err)
	}

	if replyMarkup != nil {
		markup, err := json.Marshal(replyMarkup)
		if err != nil {
			c.logger.Error("failed to encode reply markup", "err", err, "chat_id", chatID, "message_id", messageID)

			return Message{}, fmt.Errorf("encode reply markup: %w", err)
		}

		if err := writer.WriteField("reply_markup", string(markup)); err != nil {
			c.logger.Error("failed to write reply markup", "err", err, "chat_id", chatID, "message_id", messageID)

			return Message{}, fmt.Errorf("write reply_markup: %w", err)
		}
	}

	part, err := writer.CreateFormFile("captcha.png", "captcha.png")
	if err != nil {
		c.logger.Error("failed to create photo field", "err", err, "chat_id", chatID, "message_id", messageID)

		return Message{}, fmt.Errorf("create photo field: %w", err)
	}

	if err := png.Encode(part, img); err != nil {
		c.logger.Error("failed to encode captcha image", "err", err, "chat_id", chatID, "message_id", messageID)

		return Message{}, fmt.Errorf("encode captcha image: %w", err)
	}

	if err := writer.Close(); err != nil {
		c.logger.Error("failed to close multipart writer", "err", err, "chat_id", chatID, "message_id", messageID)

		return Message{}, fmt.Errorf("close multipart writer: %w", err)
	}

	apiURL := fmt.Sprintf(
		"https://api.telegram.org/bot%s/editMessageMedia",
		c.token,
	)

	c.logger.Debug("Telegram API request", "method", "editMessageMedia", "chat_id", chatID, "message_id", messageID)

	req, err := http.NewRequest(
		http.MethodPost,
		apiURL,
		&body,
	)
	if err != nil {
		c.logger.Error("failed to create editMessageMedia request", "err", err, "chat_id", chatID, "message_id", messageID)

		return Message{}, fmt.Errorf(
			"create editMessageMedia request: %w",
			err,
		)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.client.Do(req)
	if err != nil {
		c.logger.Error("editMessageMedia request failed", "err", err, "chat_id", chatID, "message_id", messageID)

		return Message{}, fmt.Errorf(
			"editMessageMedia request: %w",
			err,
		)
	}

	defer resp.Body.Close()

	var result apiResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.logger.Error("failed to decode editMessageMedia response", "err", err, "chat_id", chatID, "message_id", messageID, "status", resp.StatusCode)

		return Message{}, fmt.Errorf(
			"decode editMessageMedia response: %w",
			err,
		)
	}

	if !result.OK {
		c.logger.Error("Telegram API error", "method", "editMessageMedia", "chat_id", chatID, "message_id", messageID, "error_code", result.ErrorCode, "description", result.Description)

		return Message{}, fmt.Errorf(
			"Telegram API error %d: %s",
			result.ErrorCode,
			result.Description,
		)
	}

	var message Message

	if err := json.Unmarshal(result.Result, &message); err != nil {
		c.logger.Error("failed to decode editMessageMedia result", "err", err, "chat_id", chatID, "message_id", messageID)

		return Message{}, fmt.Errorf(
			"decode editMessageMedia result: %w",
			err,
		)
	}

	c.logger.Debug("Telegram photo message edited", "chat_id", chatID, "message_id", message.MessageID)

	return message, nil
}
