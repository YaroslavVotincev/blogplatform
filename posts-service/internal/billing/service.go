package billing

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"net/url"
	configService "posts-service/pkg/config-client"
	requestuser "posts-service/pkg/hidepost-requestuser"
	"strings"
)

type Service struct {
	ServiceUrl string
}

func NewService(serviceUrl string, cfgService *configService.ConfigServiceManager) *Service {
	service := Service{ServiceUrl: serviceUrl}
	cfgService.SetUpdateHandler(func(ss configService.ServiceSetting) {
		service.ServiceUrl = ss.Value
	}, "BILLING_SERVICE_URL")
	return &service
}

type RobokassaPaymentLinkRequest struct {
	ItemId      uuid.UUID `json:"item_id"`
	ItemType    string    `json:"item_type"`
	UserId      uuid.UUID `json:"user_id"`
	Sum         float64   `json:"sum"`
	Description string    `json:"description"`
}

type RobokassaPaymentLinkResponse struct {
	Url string `json:"url"`
}

func (s *Service) RobokassaPaymentLink(itemId, userId uuid.UUID, sum float64, itemType, description string) (string, error) {

	roboUrl, _ := url.JoinPath(s.ServiceUrl, "robokassa/payment-link")

	requestBody := RobokassaPaymentLinkRequest{
		ItemId:      itemId,
		ItemType:    itemType,
		UserId:      userId,
		Sum:         sum,
		Description: description,
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("fail to marshal request body cause %v", err)
	}

	req, err := http.NewRequest("POST", roboUrl, strings.NewReader(string(body)))
	if err != nil {
		return "", fmt.Errorf("fail to create request cause %v", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set(requestuser.UserRoleHeaderKey, requestuser.UserRoleService)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("fail to send request cause %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected response status code: %d", resp.StatusCode)
	}

	var response RobokassaPaymentLinkResponse

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return "", fmt.Errorf("fail to unmarshal response body cause %v", err)
	}

	return response.Url, nil
}
