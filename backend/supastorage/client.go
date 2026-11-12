package supastorage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	BaseURL    string
	ServiceKey string
	Bucket     string
	HTTPClient *http.Client
}

func (c *Client) httpClient() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return &http.Client{Timeout: 30 * time.Second}
}

func (c *Client) validate() error {
	if c.BaseURL == "" || c.ServiceKey == "" || c.Bucket == "" {
		return fmt.Errorf("supabase storage configuration is incomplete")
	}
	return nil
}

func (c *Client) objectURL(objectPath string) string {
	parts := strings.Split(strings.Trim(objectPath, "/"), "/")
	for i := range parts {
		parts[i] = url.PathEscape(parts[i])
	}

	return strings.TrimRight(c.BaseURL, "/") +
		"/storage/v1/object/" +
		url.PathEscape(c.Bucket) + "/" +
		strings.Join(parts, "/")
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("apikey", c.ServiceKey)

	// Legacy service_role เป็น JWT จึงส่ง Authorization ได้
	// Secret Key รูปแบบ sb_secret_ ไม่ใช่ JWT
	if !strings.HasPrefix(c.ServiceKey, "sb_secret_") {
		req.Header.Set("Authorization", "Bearer "+c.ServiceKey)
	}
}

// Upload uploads a private object to Supabase Storage.
func (c *Client) Upload(
	ctx context.Context,
	objectPath string,
	data []byte,
	contentType string,
) error {
	if err := c.validate(); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.objectURL(objectPath),
		bytes.NewReader(data),
	)
	if err != nil {
		return err
	}

	c.setHeaders(req)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("x-upsert", "false")

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return fmt.Errorf("upload slip: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf(
			"supabase upload returned %d: %s",
			resp.StatusCode,
			string(body),
		)
	}

	return nil
}

// Remove deletes an object. Used for cleanup if database insertion fails.
func (c *Client) Remove(ctx context.Context, objectPath string) error {
	if err := c.validate(); err != nil {
		return err
	}

	payload, err := json.Marshal(map[string][]string{
		"prefixes": {objectPath},
	})
	if err != nil {
		return err
	}

	endpoint := strings.TrimRight(c.BaseURL, "/") +
		"/storage/v1/object/" + url.PathEscape(c.Bucket)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodDelete,
		endpoint,
		bytes.NewReader(payload),
	)
	if err != nil {
		return err
	}

	c.setHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return fmt.Errorf("remove storage object: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf(
			"supabase delete returned %d: %s",
			resp.StatusCode,
			string(body),
		)
	}

	return nil
}

// SignedURL creates a temporary URL for viewing a private object.
func (c *Client) SignedURL(
	ctx context.Context,
	objectPath string,
) (string, error) {
	if err := c.validate(); err != nil {
		return "", err
	}

	payload, err := json.Marshal(map[string]int{
		"expiresIn": 3600,
	})
	if err != nil {
		return "", err
	}

	endpoint := c.objectURL(objectPath)
	endpoint = strings.Replace(
		endpoint,
		"/storage/v1/object/",
		"/storage/v1/object/sign/",
		1,
	)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		endpoint,
		bytes.NewReader(payload),
	)
	if err != nil {
		return "", err
	}

	c.setHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return "", fmt.Errorf("create signed URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", fmt.Errorf(
			"supabase signed URL returned %d: %s",
			resp.StatusCode,
			string(body),
		)
	}

	var result struct {
		SignedURL string `json:"signedURL"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if result.SignedURL == "" {
		return "", fmt.Errorf("supabase returned an empty signed URL")
	}

	if strings.HasPrefix(result.SignedURL, "http://") ||
		strings.HasPrefix(result.SignedURL, "https://") {
		return result.SignedURL, nil
	}

	return strings.TrimRight(c.BaseURL, "/") +
		"/storage/v1" + result.SignedURL, nil
}
