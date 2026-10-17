package booprecognition

// personResponse é o retorno de POST /persons e POST /identify na API Boop.
type personResponse struct {
	ID             int64   `json:"id"`
	Identification int64   `json:"identification"`
	Precision      float64 `json:"precision"`
	History        []int64 `json:"history"`
	Created        bool    `json:"created"`
}
