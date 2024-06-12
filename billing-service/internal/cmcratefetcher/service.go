package cmcratefetcher

import (
	"encoding/json"
	"fmt"
	configService "github.com/llc-ldbit/go-cloud-config-client"
	"net/http"
	"net/url"
	"time"
)

type Service struct {
	ServiceUrl            string
	LastTonPriceQueryTime time.Time
	LastTonPrice          float64
}

func NewService(serviceUrl string, cfgService *configService.ConfigServiceManager) *Service {
	service := Service{
		ServiceUrl:            serviceUrl,
		LastTonPriceQueryTime: time.Now().UTC().Add(-2 * time.Hour),
		LastTonPrice:          0,
	}
	cfgService.SetUpdateHandler(func(ss configService.ServiceSetting) {
		service.ServiceUrl = ss.Value
	}, "CMCRATE_SERVICE_URL")
	return &service
}

type TonPriceResponse struct {
	Price float64 `json:"price"`
}

func (s *Service) TonPrice() (float64, error) {

	if time.Since(s.LastTonPriceQueryTime) < 1*time.Hour {
		return s.LastTonPrice, nil
	}

	reqUrl, _ := url.JoinPath(s.ServiceUrl, "toncoin/rub")

	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		return 0, fmt.Errorf("fail to create request cause %v", err)
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("fail to send request cause %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected response status code: %d", resp.StatusCode)
	}

	var response TonPriceResponse

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return 0, fmt.Errorf("fail to unmarshal response body cause %v", err)
	}

	s.LastTonPrice = response.Price
	s.LastTonPriceQueryTime = time.Now().UTC()

	return response.Price, nil
}
