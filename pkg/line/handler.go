package line

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type NotifyRequest struct {
	Name        string
	CompanyName string
	Email       string
	PhoneNumber string
	Subject     string
	Content     string
}

type NotifyResponse struct {
	Status  int64  `json:"status"`
	Message string `json:"message"`
}

const (
	lineNotifyAPI = "https://notify-api.line.me/api/notify"
	messageFormat = `Calmatoホームページより以下お問い合わせがありました。
-------□■□ お問い合わせ内容 □■□-------
お名前　　　　: %s
貴社名　　　　: %s
メールアドレス: %s
電話番号　　　: %s

件名: %s
日時: %s
内容: %s
-----------------------------------
`
)

func SendNotify(ctx context.Context, token string, in *NotifyRequest) (*NotifyResponse, error) {
	now := time.Now()
	message := fmt.Sprintf(
		messageFormat,
		in.Name,
		in.CompanyName,
		in.Email,
		in.PhoneNumber,
		in.Subject,
		now.Format("2006/01/02 15:04:05"),
		in.Content,
	)

	vals := url.Values{}
	vals.Add("message", message)
	vals.Add("notificationDisabled", "false")

	body := strings.NewReader(vals.Encode())
	req, err := http.NewRequest(http.MethodPost, lineNotifyAPI, body)
	if err != nil {
		return nil, err
	}

	authorization := fmt.Sprintf("Bearer %s", token)
	req.Header.Add("Authorization", authorization)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	res := &NotifyResponse{}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(res)
	if err != nil {
		return nil, err
	}

	return res, nil
}
