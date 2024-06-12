package basicroute

import (
	"api-gateway/internal/auth"
	proxypass "api-gateway/internal/proxy-pass"
	"api-gateway/pkg/filelogger"
	"github.com/gorilla/mux"
	configService "github.com/llc-ldbit/go-cloud-config-client"
	"net/http"
	"sync"
)

type ServiceRoute struct {
	host string
	mu   *sync.RWMutex
}

func (s *ServiceRoute) setServiceHost(host string) {
	s.mu.Lock()
	s.host = host
	s.mu.Unlock()
}

func (s *ServiceRoute) getServiceHost() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.host
}

func RegisterServiceRoute(
	route *mux.Route,
	serviceHost string,
	ServiceHostConfigKey string,
	fileLogger *filelogger.FileLogger,
	cfgService *configService.ConfigServiceManager,
	authRoute *auth.ServiceRoute) {

	s := &ServiceRoute{
		host: serviceHost,
		mu:   new(sync.RWMutex),
	}

	route.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		if authRoute != nil {
			err := authRoute.SetAuthorizationHeaders(req)
			if err != nil {
				fileLogger.Error("failed to proxy request to service, authorization request failed", map[string]interface{}{
					"error":        err.Error(),
					"request_url":  req.URL.String(),
					"service_host": s.getServiceHost(),
					"service_name": ServiceHostConfigKey,
				})
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		err := proxypass.ToServiceWithHost(w, req, s.getServiceHost())
		if err != nil {
			fileLogger.Error("failed to proxy request to service", map[string]interface{}{
				"error":        err.Error(),
				"request_url":  req.URL.String(),
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

}
