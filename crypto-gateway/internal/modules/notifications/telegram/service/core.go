package service

import (
	"crypto-gateway/internal/modules/notifications/telegram/repo"
	"fmt"

	"github.com/valyala/fasthttp"
)

func sendNotificationIfNeeded(send bool, NotificationId int) (error, bool) {
	if send == false {
		return nil, true
	}
	var telegramData repo.TelegramNotificationInfo
	telegramData, _ = repo.GetTelegramNotificationInfo(NotificationId)

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", telegramData.Token)

	args := &fasthttp.Args{}
	args.Set("chat_id", fmt.Sprintf("%d", telegramData.ChatId))
	args.Set("text", telegramData.Message)

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(apiURL)
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetContentType("application/x-www-form-urlencoded")
	req.SetBody(args.QueryString())

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	client := &fasthttp.Client{}
	if err := client.Do(req, resp); err != nil {
		return fmt.Errorf("failed to send request: %w", err), false
	}

	if resp.StatusCode() != fasthttp.StatusOK {
		return fmt.Errorf("bad response: %s", resp.Body()), false
	}

	return nil, true
}
