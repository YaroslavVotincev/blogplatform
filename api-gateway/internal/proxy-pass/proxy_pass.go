package proxypass

import (
	requestuser "api-gateway/pkg/hidepost-requestuser"
	"fmt"
	"io"
	"net/http"
)

func ToServiceWithHost(w http.ResponseWriter, req *http.Request, serviceHost string) error {

	req.URL.Scheme = "http"
	req.URL.Host = serviceHost

	serviceRequest, err := http.NewRequest(req.Method, req.URL.String(), req.Body)
	if err != nil {
		return fmt.Errorf("failed to create service request: %v", err)
	}

	serviceRequest.Header.Set(requestuser.UserIdHeaderKey, req.Header.Get(requestuser.UserIdHeaderKey))
	serviceRequest.Header.Set(requestuser.UserRoleHeaderKey, req.Header.Get(requestuser.UserRoleHeaderKey))
	serviceRequest.Header.Set(requestuser.UserIsBannedHeaderKey, req.Header.Get(requestuser.UserIsBannedHeaderKey))
	serviceRequest.Header.Set(requestuser.AuthorizationHeaderKey, req.Header.Get(requestuser.AuthorizationHeaderKey))

	serviceResponse, err := http.DefaultClient.Do(serviceRequest)
	if err != nil {
		return fmt.Errorf("failed to do service request: %v", err)
	}

	defer serviceResponse.Body.Close()

	for key, values := range serviceResponse.Header {
		for _, value := range values {
			w.Header().Set(key, value)
		}
	}

	w.WriteHeader(serviceResponse.StatusCode)

	_, err = io.Copy(w, serviceResponse.Body)
	if err != nil {
		return fmt.Errorf("failed to copy service response body to client request: %v", err)
	}

	return nil
}
