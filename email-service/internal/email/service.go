package email

import (
	"fmt"
	configService "github.com/llc-ldbit/go-cloud-config-client"
	"net/http"
	"net/url"
	"strings"
)

const (
	SmtpBzApiKeyConfigKey            = "SMTP_BZ_API_KEY"
	DefaultEmailAddressConfigKey     = "DEFAULT_EMAIL_ADDRESS"
	DefaultEmailDisplayNameConfigKey = "DEFAULT_EMAIL_DISPLAY_NAME"
	SendEmailEndpointUrl             = "https://api.smtp.bz/v1/smtp/send"
)

type Service struct {
	smtpBzApiKey            string
	DefaultEmailAddress     string
	DefaultEmailDisplayName string
}

func NewService(smtpBzApiKey string, defaultEmailAddress string, defaultEmailDisplayName string,
	cfgService *configService.ConfigServiceManager) *Service {

	service := &Service{
		smtpBzApiKey:            smtpBzApiKey,
		DefaultEmailAddress:     defaultEmailAddress,
		DefaultEmailDisplayName: defaultEmailDisplayName,
	}

	cfgService.SetUpdateHandler(func(ss configService.ServiceSetting) {
		service.smtpBzApiKey = ss.Value
	}, SmtpBzApiKeyConfigKey)

	cfgService.SetUpdateHandler(func(ss configService.ServiceSetting) {
		service.DefaultEmailAddress = ss.Value
	}, DefaultEmailAddressConfigKey)

	cfgService.SetUpdateHandler(func(ss configService.ServiceSetting) {
		service.DefaultEmailDisplayName = ss.Value
	}, DefaultEmailDisplayNameConfigKey)

	return service
}

func (s *Service) SendEmail(email QueueMessage) error {

	reqBody := url.Values{}
	if email.From != nil {
		reqBody.Set("from", *email.From)
	} else {
		reqBody.Set("from", s.DefaultEmailAddress)
	}
	if email.Name != nil {
		reqBody.Set("name", *email.Name)
	} else {
		reqBody.Set("name", s.DefaultEmailDisplayName)
	}
	reqBody.Set("subject", email.Subject)
	reqBody.Set("to", email.To)
	if email.ToName != nil {
		reqBody.Set("to_name", *email.ToName)
	}
	reqBody.Set("html", email.Html)

	req, err := http.NewRequest("POST", SendEmailEndpointUrl, strings.NewReader(reqBody.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request to smtp.bz: %v", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", s.smtpBzApiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to do request to smtp.bz: %v", err)
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case 200:
		return nil
	case 400:
		return fmt.Errorf("smtp.bz returned 400 status code")
	case 401:
		return fmt.Errorf("smtp.bz returned 401 status code")
	default:
		return fmt.Errorf("smtp.bz returned unexpected status code: %d", resp.StatusCode)
	}
}
