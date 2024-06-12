package comments

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"net/url"
	configService "posts-service/pkg/config-client"
	requestuser "posts-service/pkg/hidepost-requestuser"
)

type Service struct {
	ServiceUrl string
}

func NewService(serviceUrl string, cfgService *configService.ConfigServiceManager) *Service {
	service := Service{ServiceUrl: serviceUrl}
	cfgService.SetUpdateHandler(func(ss configService.ServiceSetting) {
		service.ServiceUrl = ss.Value
	}, "COMMENTS_SERVICE_URL")
	return &service
}

type CountResponse struct {
	Amount int `json:"amount"`
}

func (s *Service) CountPostComments(postId uuid.UUID) (int, error) {

	reqUrl, _ := url.JoinPath(s.ServiceUrl, "parent/", postId.String(), "/count")

	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		return 0, fmt.Errorf("fail to create request cause %v", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set(requestuser.UserRoleHeaderKey, requestuser.UserRoleService)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("fail to send request cause %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected response status code: %d", resp.StatusCode)
	}

	var response CountResponse

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return 0, fmt.Errorf("fail to unmarshal response body cause %v", err)
	}

	return response.Amount, nil
}
