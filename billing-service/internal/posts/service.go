package posts

import (
	requestuser "billing-service/pkg/hidepost-requestuser"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	configService "github.com/llc-ldbit/go-cloud-config-client"
	"net/http"
	"net/url"
	"strings"
)

type Service struct {
	BaseUrl            string
	GrantPaidAccessUrl string
}

var GrantSubscriptionPath = "/paid-access/grant"

func NewService(serviceUrl string, cfgService *configService.ConfigServiceManager) *Service {
	service := Service{BaseUrl: serviceUrl}
	tempStr, err := url.JoinPath(serviceUrl, GrantSubscriptionPath)
	if err != nil {
		panic(err)
	}
	service.GrantPaidAccessUrl = tempStr
	cfgService.SetUpdateHandler(func(ss configService.ServiceSetting) {
		service.BaseUrl = ss.Value
		service.GrantPaidAccessUrl, _ = url.JoinPath(service.BaseUrl, GrantSubscriptionPath)
	}, "POSTS_SERVICE_URL")
	return &service
}

type GrantItemServiceRequest struct {
	ItemId   uuid.UUID `json:"item_id" validate:"required"`
	UserId   uuid.UUID `json:"user_id" validate:"required"`
	Value    float64   `json:"value" validate:"required"`
	Currency string    `json:"currency" validate:"required"`
}

func (s *Service) GrantPostPaidAccess(postId, userId uuid.UUID, value float64, currency string) error {

	requestBody := GrantItemServiceRequest{
		ItemId:   postId,
		UserId:   userId,
		Value:    value,
		Currency: currency,
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("fail to marshal request body cause %v", err)
	}

	req, err := http.NewRequest("POST", s.GrantPaidAccessUrl, strings.NewReader(string(body)))
	if err != nil {
		return fmt.Errorf("fail to create request cause %v", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set(requestuser.UserRoleHeaderKey, requestuser.UserRoleService)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("fail to send request cause %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response status code: %d", resp.StatusCode)
	}

	return nil
}
