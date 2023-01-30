package data

type TapOptions struct {
	Name       string `json:"name"`
	Code       string `json:"code"`
	WebhookURL string `json:"webhook_url"`
	BatchLimit int    `json:"batch_limit"`
}

type TapID struct {
	ID string `json:"id"`
}
