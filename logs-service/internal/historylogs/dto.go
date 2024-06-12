package historylogs

type HealthCheckResponse struct {
	Status string `json:"status"`
	Up     bool   `json:"up"`
}
