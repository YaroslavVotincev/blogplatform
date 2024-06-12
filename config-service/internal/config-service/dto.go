package configservice

type CreateServiceRequest struct {
	Service string `json:"service" validate:"required,max=64"`
}

type UpdateServiceRequest struct {
	Service string `json:"service" validate:"required,max=64"`
}

type CreateSettingRequest struct {
	Key   string `json:"key" validate:"required,max=64"`
	Value string `json:"value" validate:"required"`
}

type HealthCheckResponse struct {
	Status string `json:"status"`
	Up     bool   `json:"up"`
}
