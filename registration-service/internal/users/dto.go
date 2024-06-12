package users

type SignupRequest struct {
	Login    string `json:"login" validate:"required,min=2,max=30,login"`
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required,password"`
}

type HealthCheckResponse struct {
	Status string `json:"status"`
	Up     bool   `json:"up"`
}
