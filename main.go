package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type SendEmailRequest struct {
	Name        string `json:"name,omitempty"`
	CompanyName string `json:"companyName,omitempty"`
	Email       string `json:"email,omitempty"`
	PhoneNumber string `json:"phoneNumber,omitempty"`
	Subject     string `json:"subject,omitempty"`
	Content     string `json:"content,omitempty"`
}

type SendEmailResponse struct {
	Code int    `json:"code,omitempty"`
	Body string `json:"body,omitempty"`
}

type LambdaHandler struct{}

const (
	mailFormatWithText = `※このメールはシステムからの自動返信です

%s 様
Calmatoへお問い合わせありがとうございます。

以下の内容でお問い合わせを受付致しました。
後日、担当者よりご連絡いたしますので今しばらくお待ちくださいませ。

-------□■□ お問い合わせ内容 □■□-------
お名前　　　　: %s
貴社名　　　　: %s
メールアドレス: %s
電話番号　　　: %s

件名: %s
日時: %s
内容: %s
-----------------------------------

$ %s
$ Email: %s
`
	mailFormatWithHTML = `
<p>※このメールはシステムからの自動返信です<p>
<p>%s 様</p>
<p>
	Calmatoへお問い合わせありがとうございます。<br />
	以下の内容でお問い合わせを受付致しました。<br />
	後日、担当者よりご連絡いたしますので今しばらくお待ちくださいませ。
</p>
<h3>-------□■□ お問い合わせ内容 □■□-------</h3>
<table>
	<tbody>
		<tr>
			<td>お名前</td>
			<td>%s</td>
		</tr>
		<tr>
			<td>貴社名</td>
			<td>%s</td>
		</tr>
		<tr>
			<td>メールアドレス</td>
			<td>%s</td>
		</tr>
		<tr>
			<td>電話番号</td>
			<td>%s</td>
		</tr>
		<tr>
			<td>件名</td>
			<td>%s</td>
		</tr>
		<tr>
			<td>日時</td>
			<td>%s</td>
		</tr>
		<tr>
			<td>内容</td>
			<td>%s</td>
		</tr>
	</tbody>
</table>
<h3>-----------------------------------</h3>
$ %s<br />
$ email: %s<br />
$ url: %s`
)

var (
	apiKey    = ""
	fromEmail = ""
	fromName  = "Calmato 担当者"
	homepage  = "https://www.calmato.jp"
)

func init() {
	apiKey = os.Getenv("SENDGRID_API_KEY")
	fromEmail = os.Getenv("SENDGRID_EMAIL")
}

func main() {
	lambda.Start(lambdaHandler)
}

/**
 * お問い合わせ受付メールの送信
 */
func lambdaHandler(ctx context.Context, req SendEmailRequest) (SendEmailResponse, error) {
	var res SendEmailResponse
	now := time.Now()

	client := newSendgridClient()
	message := newSendgridMessage(&req, now)
	out, err := client.Send(message)
	if err != nil {
		return res, err
	}

	res.Code = out.StatusCode
	res.Body = out.Body
	return res, nil
}

func newSendgridClient() *sendgrid.Client {
	return sendgrid.NewSendClient(apiKey)
}

func newSendgridMessage(req *SendEmailRequest, now time.Time) *mail.SGMailV3 {
	// Common
	to := mail.NewEmail(fromName, fromEmail)
	from := mail.NewEmail(req.Name, req.Email)
	subject := newSubject()
	contentWithText := mail.NewContent("text/plain", newContentWithText(req, now))
	contentWithHTML := mail.NewContent("text/html", newContentWithHTML(req, now))

	// Personalization
	personalization := mail.NewPersonalization()
	personalization.AddTos(to)
	personalization.AddBCCs(from)
	personalization.SetSubstitution("%fullname%", req.Name)
	personalization.SetSendAt(int(now.Unix()))

	// Mail
	message := mail.NewV3Mail()
	message.Subject = subject
	message.SetFrom(from)
	message.AddPersonalizations(personalization)
	message.AddContent(contentWithText, contentWithHTML)
	return message
}

func newSubject() string {
	return fmt.Sprintf("[Calmato] お問い合わせありがとうございます")
}

func newContentWithText(req *SendEmailRequest, now time.Time) string {
	return fmt.Sprintf(
		mailFormatWithText,
		req.Name,
		req.Name,
		req.CompanyName,
		req.Email,
		req.PhoneNumber,
		req.Subject,
		now.Format("2006/01/02 15:04:05"),
		req.Content,
		fromName,
		fromEmail,
	)
}

func newContentWithHTML(req *SendEmailRequest, now time.Time) string {
	return fmt.Sprintf(
		mailFormatWithHTML,
		req.Name,
		req.Name,
		req.CompanyName,
		req.Email,
		req.PhoneNumber,
		req.Subject,
		now.Format("2006/01/02 15:04:05"),
		req.Content,
		fromName,
		fromEmail,
		homepage,
	)
}
