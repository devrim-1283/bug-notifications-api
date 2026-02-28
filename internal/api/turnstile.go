package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const turnstileVerifyURL = "https://challenges.cloudflare.com/turnstile/v0/siteverify"

type turnstileResponse struct {
	Success bool `json:"success"`
}

// verifyTurnstile validates a Cloudflare Turnstile token server-side.
func verifyTurnstile(secretKey, token, remoteIP string) error {
	if token == "" {
		return fmt.Errorf("turnstile token is required")
	}

	resp, err := (&http.Client{Timeout: 10 * time.Second}).PostForm(turnstileVerifyURL, url.Values{
		"secret":   {secretKey},
		"response": {token},
		"remoteip": {remoteIP},
	})
	if err != nil {
		return fmt.Errorf("turnstile verify request failed: %w", err)
	}
	defer resp.Body.Close()

	var result turnstileResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("turnstile verify decode failed: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("turnstile verification failed")
	}
	return nil
}
