package comments

type CommentCreateRequest struct {
	Content string `json:"content"`
}

type HealthCheckResponse struct {
	Status string `json:"status"`
	Up     bool   `json:"up"`
}
