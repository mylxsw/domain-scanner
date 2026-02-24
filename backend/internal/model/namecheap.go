package model

// NamecheapConfig holds the configuration for Namecheap API
type NamecheapConfig struct {
	ApiUser  string
	ApiKey   string
	Username string
	ClientIp string
	ApiBase  string
}

// DomainCheckResult represents the result of a domain check
type DomainCheckResult struct {
	Domain                   string `json:"Domain"`
	Available                string `json:"Available"`
	IsPremiumName            string `json:"IsPremiumName"`
	PremiumRegistrationPrice string `json:"PremiumRegistrationPrice"`
	IcannFee                 string `json:"IcannFee"`
	EapFee                   string `json:"EapFee"`
}

// ErrorInfo represents an API error
type ErrorInfo struct {
	Number string `json:"number"`
	Text   string `json:"text"`
}

// ApiCallError represents an API call error
type ApiCallError struct {
	Message    string
	HTTPStatus int
	Errors     []ErrorInfo
}

func (e *ApiCallError) Error() string {
	return e.Message
}
