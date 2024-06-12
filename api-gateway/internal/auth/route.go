package auth

import (
	proxypass "api-gateway/internal/proxy-pass"
	"api-gateway/pkg/filelogger"
	requestuser "api-gateway/pkg/hidepost-requestuser"
	"fmt"
	"github.com/gorilla/mux"
	configService "github.com/llc-ldbit/go-cloud-config-client"
	"io"
	"net/http"
	"net/url"
	"sync"
)

const (
	ServiceHostConfigKey  = "AUTHENTICATION_SERVICE"
	AuthorizationEndpoint = "/api/v1/auth/service/authorize"
)

type ServiceRoute struct {
	host string
	mu   *sync.Mutex
}

func (s *ServiceRoute) setServiceHost(host string) {
	s.mu.Lock()
	s.host = host
	s.mu.Unlock()
}

func (s *ServiceRoute) getServiceHost() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.host
}

func RegisterServiceRoute(
	route *mux.Route,
	serviceHost string,
	fileLogger *filelogger.FileLogger,
	cfgService *configService.ConfigServiceManager) *ServiceRoute {

	s := &ServiceRoute{
		host: serviceHost,
		mu:   new(sync.Mutex),
	}

	route.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		err := proxypass.ToServiceWithHost(w, req, s.getServiceHost())

		if err != nil {
			fileLogger.Error("failed to proxy request to service", map[string]interface{}{
				"error":        err.Error(),
				"request_path": req.URL.String(),
				"service_host": s.getServiceHost(),
				"service_name": ServiceHostConfigKey,
			})
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

	cfgService.SetUpdateHandler(func(ss configService.ServiceSetting) {
		s.setServiceHost(ss.Value)
	}, ServiceHostConfigKey)

	return s
}

func (s *ServiceRoute) SetAuthorizationHeaders(req *http.Request) error {

	reqAuthHeader := req.Header.Get(requestuser.AuthorizationHeaderKey)
	if reqAuthHeader == "" {
		return nil
	}

	authorizationUrl := url.URL{
		Scheme: "http",
		Host:   s.getServiceHost(),
		Path:   AuthorizationEndpoint,
	}

	serviceRequest, err := http.NewRequest("HEAD", authorizationUrl.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to create auth service authorization request: %v", err)
	}
	serviceRequest.Header.Set(requestuser.AuthorizationHeaderKey, reqAuthHeader)
	serviceRequest.Header.Set("USER-REQUEST-URL", req.URL.String())

	serviceResponse, err := http.DefaultClient.Do(serviceRequest)
	if err != nil {
		return fmt.Errorf("failed to send auth service authorization request: %v", err)
	}
	defer serviceResponse.Body.Close()

	if serviceResponse.StatusCode != 200 && serviceResponse.StatusCode != 401 {
		return fmt.Errorf("unexpected auth service authorization response status code: %d", serviceResponse.StatusCode)
	}

	req.Header.Set(requestuser.UserIdHeaderKey, serviceResponse.Header.Get(requestuser.UserIdHeaderKey))
	req.Header.Set(requestuser.UserRoleHeaderKey, serviceResponse.Header.Get(requestuser.UserRoleHeaderKey))
	req.Header.Set(requestuser.UserIsBannedHeaderKey, serviceResponse.Header.Get(requestuser.UserIsBannedHeaderKey))

	_, _ = io.Copy(io.Discard, serviceResponse.Body)
	return nil
}
