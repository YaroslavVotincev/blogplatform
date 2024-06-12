package users

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
	}, "USERS_SERVICE_URL")
	return &service
}

type AddRequest struct {
	Value float64 `json:"value"`
}

func (s *Service) AddRubToBalanceOfUser(userId uuid.UUID, value float64) error {

	reqUrl, _ := url.JoinPath(s.ServiceUrl, "id/", userId.String(), "/wallet/add-rub")

	requestBody := AddRequest{
		Value: value,
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("fail to marshal request body cause %v", err)
	}

	req, err := http.NewRequest("PUT", reqUrl, strings.NewReader(string(body)))
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

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("unexpected response status code: %d", resp.StatusCode)
	}

	return nil
}
