package wallet

import (
	"bytes"
	"encoding/json"
	"fmt"
	configService "github.com/llc-ldbit/go-cloud-config-client"
	"net/http"
	"net/url"
)

const (
	UrlConfigKey       = "WALLET_SERVICE_URL"
	GetBalanceEndpoint = "/api/v1/walletBalance"
	CreateEndpoint     = "/api/v1/create"
)

type Service struct {
	url string
}

func NewService(serviceUrl string, cfgService *configService.ConfigServiceManager) (*Service, error) {

	service := &Service{
		url: serviceUrl,
	}

	_, err := url.JoinPath(serviceUrl, CreateEndpoint)
	if err != nil {
		return nil, err
	}

	_, err = url.JoinPath(serviceUrl, GetBalanceEndpoint)
	if err != nil {
		return nil, err
	}

	cfgService.SetUpdateHandler(func(ss configService.ServiceSetting) {
		service.url = ss.Value
	}, UrlConfigKey)

	return service, nil
}

func (s *Service) GetBalance(address string) (float64, error) {

	getUrl, err := url.JoinPath(s.url, GetBalanceEndpoint)
	if err != nil {
		return 0, err
	}

	reqBody, err := json.Marshal(GetBalanceRequest{Address: address})
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequest("POST", getUrl, bytes.NewBuffer(reqBody))
	if err != nil {
		return 0, err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var respBody GetBalanceResponse
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return 0, err
	}

	return respBody.Balance, nil
}

func (s *Service) CreateWallet() (*CreateWalletResponse, error) {

	createUrl, err := url.JoinPath(s.url, CreateEndpoint)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", createUrl, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected wallet service http status code: %d", resp.StatusCode)
	}

	var respBody CreateWalletResponse
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return nil, err
	}

	return &respBody, nil
}
