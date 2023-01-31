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

type TapListEntry struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	WebhookURL string `json:"webhook_url"`
}

type TapListPage struct {
	Entries []TapListEntry `json:"entries"`
	Cursor  string         `json:"cursor,omitempty"`
}
