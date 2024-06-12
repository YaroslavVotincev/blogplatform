package files

type HealthCheckResponse struct {
	Status string `json:"status"`
	Up     bool   `json:"up"`
}
