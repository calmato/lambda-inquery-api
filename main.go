package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/calmato/lambda-inquiry-api/pkg/line"
	"github.com/calmato/lambda-inquiry-api/pkg/sendgrid"
	"golang.org/x/sync/errgroup"
)

type CreateInquiryRequest struct {
	Name        string `json:"name,omitempty"`
	CompanyName string `json:"companyName,omitempty"`
	Email       string `json:"email,omitempty"`
	PhoneNumber string `json:"phoneNumber,omitempty"`
	Subject     string `json:"subject,omitempty"`
	Content     string `json:"content,omitempty"`
}

type CreateInquiryResponse struct {
	Code int         `json:"code"`
	Body interface{} `json:"body"`
}

var (
	apiKey    = ""
	lineToken = ""
	fromEmail = ""
)

func init() {
	apiKey = os.Getenv("SENDGRID_API_KEY")
	fromEmail = os.Getenv("SENDGRID_EMAIL")
	lineToken = os.Getenv("LINE_API_TOKEN")
}

func main() {
	lambda.Start(lambdaHandler)
}

/**
 * お問い合わせ受付メールの送信
 */
func lambdaHandler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	in, err := getRequest(req.Body)
	if err != nil {
		res := newErrorResponse(err)
		return res, err
	}

	eg, ectx := errgroup.WithContext(ctx)
	var res *sendgrid.SendEmailResponse
	eg.Go(func() (err error) {
		sendgridReq := newSendgridRequest(in)
		res, err = sendgrid.SendEmail(ectx, apiKey, fromEmail, sendgridReq)
		return
	})
	eg.Go(func() (err error) {
		lineReq := newLINERequest(in)
		_, err = line.SendNotify(ectx, lineToken, lineReq)
		return
	})
	if err := eg.Wait(); err != nil {
		res := newErrorResponse(err)
		return res, err
	}

	out := newResponse(res)
	return out, nil
}

func getRequest(body string) (*CreateInquiryRequest, error) {
	req := &CreateInquiryRequest{}
	buf := []byte(body)

	err := json.Unmarshal(buf, req)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func newSendgridRequest(in *CreateInquiryRequest) *sendgrid.SendEmailRequest {
	return &sendgrid.SendEmailRequest{
		Name:        in.Name,
		CompanyName: in.CompanyName,
		Email:       in.Email,
		PhoneNumber: in.PhoneNumber,
		Subject:     in.Subject,
		Content:     in.Content,
	}
}

func newLINERequest(in *CreateInquiryRequest) *line.NotifyRequest {
	return &line.NotifyRequest{
		Name:        in.Name,
		CompanyName: in.CompanyName,
		Email:       in.Email,
		PhoneNumber: in.PhoneNumber,
		Subject:     in.Subject,
		Content:     in.Content,
	}
}

func newResponse(out *sendgrid.SendEmailResponse) events.APIGatewayProxyResponse {
	b, _ := json.Marshal(out.Body)

	return events.APIGatewayProxyResponse{
		Headers: map[string]string{
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "*",
			"Access-Control-Allow-Headers": "*",
			"Content-Type":                 "application/json",
		},
		StatusCode: out.Code,
		Body:       string(b),
	}
}

func newErrorResponse(err error) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		Headers: map[string]string{
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "*",
			"Access-Control-Allow-Headers": "*",
			"Content-Type":                 "application/json",
		},
		StatusCode: http.StatusInternalServerError,
		Body:       err.Error(),
	}
}
