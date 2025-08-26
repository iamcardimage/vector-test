package external

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
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

type HTTPUsersResponse struct {
	Success     bool              `json:"success"`
	TotalCount  int               `json:"total_count"`
	PerPage     int               `json:"per_page"`
	CurrentPage int               `json:"current_page"`
	TotalPages  int               `json:"total_pages"`
	Users       []json.RawMessage `json:"users"`
}

// Retry механизм для HTTP запросов
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
			case http.StatusTooManyRequests, http.StatusBadGateway,
				http.StatusServiceUnavailable, http.StatusGatewayTimeout:
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

// GetUsersRaw возвращает HTTP transport структуру
func (c *Client) GetUsersRaw(ctx context.Context, page, perPage int) (*HTTPUsersResponse, error) {
	if c.baseURL == "" {
		return nil, fmt.Errorf("EXTERNAL_API_BASE_URL is empty")
	}

	// Добавляем путь для users к базовому URL
	usersURL := c.baseURL + "/users"
	u, err := url.Parse(usersURL)
	if err != nil {
		return nil, fmt.Errorf("invalid users url: %w", err)
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

	var out HTTPUsersResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}

	return &out, nil
}

type HTTPContractsResponse struct {
	Success     bool              `json:"success"`
	TotalCount  int               `json:"total_count"`
	PerPage     int               `json:"per_page"`
	CurrentPage int               `json:"current_page"`
	TotalPages  int               `json:"total_pages"`
	Contracts   []json.RawMessage `json:"contracts"`
}

// GetContractsRaw возвращает HTTP transport структуру для договоров
func (c *Client) GetContractsRaw(ctx context.Context, page, perPage int) (*HTTPContractsResponse, error) {
	if c.baseURL == "" {
		return nil, fmt.Errorf("EXTERNAL_API_BASE_URL is empty")
	}

	// Используем базовый URL и добавляем путь для contracts
	contractsURL := c.baseURL + "/contracts"
	u, err := url.Parse(contractsURL)
	if err != nil {
		return nil, fmt.Errorf("invalid contracts url: %w", err)
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

	log.Printf("[contracts] requesting: %s", u.String())

	resp, err := c.doWithRetry(req)
	if err != nil {
		log.Printf("[contracts] request error: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	log.Printf("[contracts] response status: %s", resp.Status)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
		log.Printf("[contracts] error response body: %s", string(b))
		return nil, fmt.Errorf("external api status: %s body: %s", resp.Status, strings.TrimSpace(string(b)))
	}

	var out HTTPContractsResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		log.Printf("[contracts] json decode error: %v", err)
		return nil, err
	}

	log.Printf("[contracts] success: total_count=%d, contracts_count=%d", out.TotalCount, len(out.Contracts))

	return &out, nil
}
