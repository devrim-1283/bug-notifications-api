package model

import "time"

type Category string

const (
	CategoryDesign        Category = "design"
	CategoryFunctionality Category = "functionality"
	CategoryPerformance   Category = "performance"
	CategoryContent       Category = "content"
	CategoryMobile        Category = "mobile"
	CategorySecurity      Category = "security"
	CategoryOther         Category = "other"
)

var ValidCategories = map[Category]bool{
	CategoryDesign:        true,
	CategoryFunctionality: true,
	CategoryPerformance:   true,
	CategoryContent:       true,
	CategoryMobile:        true,
	CategorySecurity:      true,
	CategoryOther:         true,
}

type ContactType string

const (
	ContactPhone     ContactType = "phone"
	ContactEmail     ContactType = "email"
	ContactTelegram  ContactType = "telegram"
	ContactInstagram ContactType = "instagram"
)

var ValidContactTypes = map[ContactType]bool{
	ContactPhone:     true,
	ContactEmail:     true,
	ContactTelegram:  true,
	ContactInstagram: true,
}

// ReportRequest is the incoming API request body.
type ReportRequest struct {
	SiteID       string   `json:"site_id"`
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	Category     Category `json:"category"`
	PageURL      *string  `json:"page_url,omitempty"`
	ContactType  *string  `json:"contact_type,omitempty"`
	ContactValue *string  `json:"contact_value,omitempty"`
	FirstName    *string  `json:"first_name,omitempty"`
	LastName     *string  `json:"last_name,omitempty"`
	ImageURLs    []string `json:"image_urls,omitempty"`
}

// QueueMessage is what gets pushed to Redis.
type QueueMessage struct {
	EventID      string   `json:"event_id"`
	SiteID       string   `json:"site_id"`
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	Category     Category `json:"category"`
	PageURL      *string  `json:"page_url,omitempty"`
	ContactType  *string  `json:"contact_type,omitempty"`
	ContactValue *string  `json:"contact_value,omitempty"`
	FirstName    *string  `json:"first_name,omitempty"`
	LastName     *string  `json:"last_name,omitempty"`
	ImageURLs    []string `json:"image_urls,omitempty"`
	ReceivedAt   string   `json:"received_at"`
	RetryCount   int      `json:"retry_count"`
}

// BugReport is the database row.
type BugReport struct {
	ID           string    `json:"id"`
	SiteID       string    `json:"site_id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Category     string    `json:"category"`
	PageURL      *string   `json:"page_url,omitempty"`
	ContactType  *string   `json:"contact_type,omitempty"`
	ContactValue *string   `json:"contact_value,omitempty"`
	FirstName    *string   `json:"first_name,omitempty"`
	LastName     *string   `json:"last_name,omitempty"`
	ImageURLs    []string  `json:"image_urls,omitempty"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

// ReportResponse is the API response.
type ReportResponse struct {
	EventID string `json:"event_id"`
	Queued  bool   `json:"queued"`
}

// ErrorResponse is the standard error format.
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
}
