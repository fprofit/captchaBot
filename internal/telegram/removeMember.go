package telegram

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

func (c *Client) RemoveChatMember(chatID, userID int64) error {
	if err := c.banChatMember(chatID, userID); err != nil {
		return err
	}

	if err := c.unbanChatMember(chatID, userID); err != nil {
		return err
	}

	c.logger.Info("user removed from chat", "chat_id", chatID, "user_id", userID)
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

	c.logger.Debug("Telegram API request", "method", "banChatMember", "chat_id", chatID, "user_id", userID)

	resp, err := c.client.PostForm(apiURL, params)
	if err != nil {
		c.logger.Error("banChatMember request failed", "err", err, "chat_id", chatID, "user_id", userID)
		return fmt.Errorf("banChatMember request: %w", err)
	}

	defer resp.Body.Close()

	var result apiResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.logger.Error("failed to decode banChatMember response", "err", err, "chat_id", chatID, "user_id", userID)
		return fmt.Errorf("decode banChatMember response: %w", err)
	}

	if !result.OK {
		c.logger.Error("Telegram API error", "method", "banChatMember", "chat_id", chatID, "user_id", userID, "error_code", result.ErrorCode, "description", result.Description)
		return fmt.Errorf("Telegram API error %d: %s", result.ErrorCode, result.Description)
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

	c.logger.Debug("Telegram API request", "method", "unbanChatMember", "chat_id", chatID, "user_id", userID)

	resp, err := c.client.PostForm(apiURL, params)
	if err != nil {
		c.logger.Error("unbanChatMember request failed", "err", err, "chat_id", chatID, "user_id", userID)
		return fmt.Errorf("unbanChatMember request: %w", err)
	}

	defer resp.Body.Close()

	var result apiResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.logger.Error("failed to decode unbanChatMember response", "err", err, "chat_id", chatID, "user_id", userID)

		return fmt.Errorf("decode unbanChatMember response: %w", err)
	}

	if !result.OK {
		c.logger.Error("Telegram API error", "method", "unbanChatMember", "chat_id", chatID, "user_id", userID, "error_code", result.ErrorCode, "description", result.Description)

		return fmt.Errorf("Telegram API error %d: %s", result.ErrorCode, result.Description)
	}

	return nil
}

func (c *Client) RestrictChatMember(chatID, userID int64) error {
	permissions := ChatPermissions{}

	data, err := json.Marshal(permissions)
	if err != nil {
		c.logger.Error("failed to marshal chat permissions", "err", err, "chat_id", chatID, "user_id", userID)
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

	c.logger.Debug("Telegram API request", "method", "restrictChatMember", "chat_id", chatID, "user_id", userID)

	resp, err := c.client.PostForm(apiURL, params)
	if err != nil {
		c.logger.Error("restrictChatMember request failed", "err", err, "chat_id", chatID, "user_id", userID)
		return fmt.Errorf("restrictChatMember request: %w", err)
	}

	defer resp.Body.Close()

	var result apiResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.logger.Error("failed to decode restrictChatMember response", "err", err, "chat_id", chatID, "user_id", userID)

		return fmt.Errorf("decode restrictChatMember response: %w", err)
	}

	if !result.OK {
		c.logger.Error("Telegram API error", "method", "restrictChatMember", "chat_id", chatID, "user_id", userID, "error_code", result.ErrorCode, "description", result.Description)
		return fmt.Errorf("Telegram API error %d: %s", result.ErrorCode, result.Description)
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
		c.logger.Error("failed to marshal chat permissions", "err", err, "chat_id", chatID, "user_id", userID)
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
