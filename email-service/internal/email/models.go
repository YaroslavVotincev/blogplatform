package email

type QueueMessage struct {
	From    *string `json:"from,omitempty"`
	Name    *string `json:"name,omitempty"`
	Subject string  `json:"subject"`
	To      string  `json:"to"`
	ToName  *string `json:"to_name,omitempty"`
	Html    string  `json:"html"`
}
