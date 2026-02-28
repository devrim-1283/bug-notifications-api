package validate

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/devrimsoft/bug-notifications-api/internal/model"
)

const (
	MaxTitleLen       = 200
	MaxDescriptionLen = 5000
	MaxURLLen         = 2048
	MaxContactLen     = 200
	MaxNameLen        = 100
)

// htmlTagPattern matches HTML tags and common XSS vectors.
var htmlTagPattern = regexp.MustCompile(`<[^>]*>`)

// dangerousPatterns matches javascript: URIs and event handlers that could execute scripts.
var dangerousPatterns = regexp.MustCompile(`(?i)(javascript\s*:|on\w+\s*=)`)

// sanitizeString strips HTML tags and dangerous patterns from user input to prevent stored XSS.
func sanitizeString(s string) string {
	s = htmlTagPattern.ReplaceAllString(s, "")
	s = dangerousPatterns.ReplaceAllString(s, "")
	return s
}

// sanitizePtr applies sanitization to an optional string pointer.
func sanitizePtr(s *string) {
	if s != nil && *s != "" {
		*s = sanitizeString(*s)
	}
}

// ReportRequest validates the incoming report request and returns a list of errors.
// It also sanitizes text fields in-place to strip HTML/script tags (stored XSS prevention).
func ReportRequest(r *model.ReportRequest) []string {
	var errs []string

	// Sanitize all free-text fields before validation
	r.Title = sanitizeString(r.Title)
	r.Description = sanitizeString(r.Description)
	sanitizePtr(r.ContactValue)
	sanitizePtr(r.FirstName)
	sanitizePtr(r.LastName)

	// Required fields
	if r.ReportType == "" {
		r.ReportType = model.ReportTypeBug // default
	}
	if !model.ValidReportTypes[r.ReportType] {
		errs = append(errs, fmt.Sprintf("invalid report_type %q, must be 'bug' or 'request'", r.ReportType))
	}
	if strings.TrimSpace(r.SiteID) == "" {
		errs = append(errs, "site_id is required")
	}
	if strings.TrimSpace(r.Title) == "" {
		errs = append(errs, "title is required")
	} else if len(r.Title) > MaxTitleLen {
		errs = append(errs, fmt.Sprintf("title must be at most %d characters", MaxTitleLen))
	}
	if strings.TrimSpace(r.Description) == "" {
		errs = append(errs, "description is required")
	} else if len(r.Description) > MaxDescriptionLen {
		errs = append(errs, fmt.Sprintf("description must be at most %d characters", MaxDescriptionLen))
	}
	if r.Category == "" {
		errs = append(errs, "category is required")
	} else if !model.ValidCategories[r.Category] {
		errs = append(errs, fmt.Sprintf("invalid category %q", r.Category))
	}

	// Optional fields
	if r.PageURL != nil && *r.PageURL != "" {
		if len(*r.PageURL) > MaxURLLen {
			errs = append(errs, fmt.Sprintf("page_url must be at most %d characters", MaxURLLen))
		} else if u, err := url.ParseRequestURI(*r.PageURL); err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
			errs = append(errs, "page_url must be a valid http or https URL")
		}
	}

	if r.ContactType != nil && *r.ContactType != "" {
		ct := model.ContactType(*r.ContactType)
		if !model.ValidContactTypes[ct] {
			errs = append(errs, fmt.Sprintf("invalid contact_type %q", *r.ContactType))
		}
	}

	if r.ContactValue != nil && len(*r.ContactValue) > MaxContactLen {
		errs = append(errs, fmt.Sprintf("contact_value must be at most %d characters", MaxContactLen))
	}

	if r.FirstName != nil && len(*r.FirstName) > MaxNameLen {
		errs = append(errs, fmt.Sprintf("first_name must be at most %d characters", MaxNameLen))
	}

	if r.LastName != nil && len(*r.LastName) > MaxNameLen {
		errs = append(errs, fmt.Sprintf("last_name must be at most %d characters", MaxNameLen))
	}

	return errs
}
