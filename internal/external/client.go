package external

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

func NewClient() *Client {
	base := os.Getenv("EXTERNAL_API_BASE_URL")
	token := os.Getenv("EXTERNAL_API_TOKEN")
	return &Client{
		baseURL: base,
		token:   token,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

type ExternalUser struct {
	ID         int    `json:"id"`
	Surname    string `json:"surname"`
	Name       string `json:"name"`
	Patronymic string `json:"patronymic"`
}

type ExternalUsersResponse struct {
	Success     bool           `json:"success"`
	TotalCount  int            `json:"total_count"`
	PerPage     int            `json:"per_page"`
	CurrentPage int            `json:"current_page"`
	TotalPages  int            `json:"total_pages"`
	Users       []ExternalUser `json:"users"`
}

func (c *Client) GetUsers(ctx context.Context, page, perPage int) (*ExternalUsersResponse, error) {
	if c.baseURL == "" {
		return nil, fmt.Errorf("EXTERNAL_API_BASE_URL is empty")
	}

	u, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base url: %w", err)
	}

	q := u.Query()
	q.Set("page", fmt.Sprintf("%d", page))
	q.Set("per_page", fmt.Sprintf("%d", perPage))
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Basic "+c.token)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.doWithRetry(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
		return nil, fmt.Errorf("external api status: %s body: %s", resp.Status, strings.TrimSpace(string(b)))
	}

	var out ExternalUsersResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

type ExternalUsersRawResponse struct {
	Success     bool              `json:"success"`
	TotalCount  int               `json:"total_count"`
	PerPage     int               `json:"per_page"`
	CurrentPage int               `json:"current_page"`
	TotalPages  int               `json:"total_pages"`
	Users       []json.RawMessage `json:"users"`
}

func (c *Client) doWithRetry(req *http.Request) (*http.Response, error) {
	const maxRetries = 3
	base := 500 * time.Millisecond

	for i := 0; i <= maxRetries; i++ {
		resp, err := c.httpClient.Do(req)
		if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return resp, nil
		}
		retry := false
		if err != nil {
			retry = true
		} else {
			switch resp.StatusCode {
			case http.StatusTooManyRequests, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
				retry = true
			}
		}
		if !retry || i == maxRetries {
			return resp, err
		}
		if resp != nil && resp.Body != nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
		time.Sleep(base << i)
	}
	return nil, fmt.Errorf("unreachable")
}

func (c *Client) GetUsersRaw(ctx context.Context, page, perPage int) (*ExternalUsersRawResponse, error) {
	if c.baseURL == "" {
		return nil, fmt.Errorf("EXTERNAL_API_BASE_URL is empty")
	}

	u, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base url: %w", err)
	}
	q := u.Query()
	q.Set("page", fmt.Sprintf("%d", page))
	q.Set("per_page", fmt.Sprintf("%d", perPage))
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Basic "+c.token)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.doWithRetry(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
		return nil, fmt.Errorf("external api status: %s body: %s", resp.Status, strings.TrimSpace(string(b)))
	}

	var out ExternalUsersRawResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}
