package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"mime/multipart"
	"net/http"
	"strconv"
)

func (c *Client) SendCaptchaMessageImg(chatID int64, text string, img image.Image, replyMarkup *InlineKeyboardMarkup) (Message, error) {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendPhoto", c.token)

	c.logger.Debug("Telegram API request", "method", "sendPhoto", "chat_id", chatID)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	if err := writer.WriteField("chat_id", strconv.FormatInt(chatID, 10)); err != nil {
		return Message{}, err
	}

	if err := writer.WriteField("caption", text); err != nil {
		return Message{}, err
	}

	if err := writer.WriteField("parse_mode", "HTML"); err != nil {
		return Message{}, err
	}

	if err := writeReplyMarkup(writer, replyMarkup); err != nil {
		return Message{}, err
	}

	if err := writeImage(writer, "photo", "captcha.png", img); err != nil {
		return Message{}, err
	}

	if err := writer.Close(); err != nil {
		return Message{}, fmt.Errorf("close multipart writer: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, apiURL, &body)
	if err != nil {
		return Message{}, fmt.Errorf("create sendPhoto request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	message, err := c.doMessageRequest(req)
	if err != nil {
		c.logger.Error("sendPhoto request failed", "err", err, "chat_id", chatID)
		return Message{}, err
	}
	c.logger.Debug("Telegram API response", "method", "sendPhoto", "chat_id", chatID, "message_id", message.MessageID)
	return message, nil
}

func (c *Client) EditCaptchaMessageImg(chatID, messageID int64, caption string, img image.Image, replyMarkup *InlineKeyboardMarkup) (Message, error) {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/editMessageMedia", c.token)

	c.logger.Debug("Telegram API request", "method", "editMessageMedia", "chat_id", chatID, "message_id", messageID)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	if err := writer.WriteField("chat_id", strconv.FormatInt(chatID, 10)); err != nil {
		return Message{}, err
	}

	if err := writer.WriteField("message_id", strconv.FormatInt(messageID, 10)); err != nil {
		return Message{}, err
	}
	media := map[string]string{
		"type":       "photo",
		"media":      "attach://photo",
		"caption":    caption,
		"parse_mode": "HTML",
	}

	data, err := json.Marshal(media)
	if err != nil {
		return Message{}, fmt.Errorf("encode media: %w", err)
	}
	if err := writer.WriteField("media", string(data)); err != nil {
		return Message{}, err
	}
	if err := writeReplyMarkup(writer, replyMarkup); err != nil {
		return Message{}, err
	}
	if err := writeImage(writer, "photo", "captcha.png", img); err != nil {
		return Message{}, err
	}
	if err := writer.Close(); err != nil {
		return Message{}, fmt.Errorf("close multipart writer: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, apiURL, &body)
	if err != nil {
		return Message{}, fmt.Errorf("create sendPhoto request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	message, err := c.doMessageRequest(req)
	if err != nil {
		c.logger.Error("editMessageMedia request failed", "err", err, "chat_id", chatID, "message_id", messageID)
		return Message{}, err
	}
	c.logger.Debug("Telegram API response", "method", "editMessageMedia", "chat_id", chatID, "message_id", messageID, "new_message_id", message.MessageID)
	return message, nil
}

func (c *Client) doMessageRequest(req *http.Request) (Message, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return Message{}, fmt.Errorf("execute request: %w", err)
	}

	defer resp.Body.Close()

	var result apiResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return Message{}, fmt.Errorf("decode Telegram API response: %w", err)
	}

	if !result.OK {
		return Message{}, fmt.Errorf("Telegram API error %d: %s", result.ErrorCode, result.Description)
	}

	var message Message

	if err := json.Unmarshal(result.Result, &message); err != nil {
		return Message{}, fmt.Errorf("decode Telegram message: %w", err)
	}

	return message, nil
}

func writeReplyMarkup(writer *multipart.Writer, replyMarkup *InlineKeyboardMarkup) error {
	if replyMarkup == nil {
		return nil
	}

	markup, err := json.Marshal(replyMarkup)
	if err != nil {
		return fmt.Errorf("encode reply markup: %w", err)
	}

	if err := writer.WriteField("reply_markup", string(markup)); err != nil {
		return fmt.Errorf("write reply_markup: %w", err)
	}

	return nil
}

func writeImage(writer *multipart.Writer, fieldName, fileName string, img image.Image) error {
	part, err := writer.CreateFormFile(fieldName, fileName)
	if err != nil {
		return fmt.Errorf("create photo field: %w", err)
	}

	if err := png.Encode(part, img); err != nil {
		return fmt.Errorf("encode captcha image: %w", err)
	}

	return nil
}
