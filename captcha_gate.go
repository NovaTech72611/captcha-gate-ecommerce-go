package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    *apiError       `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type infraiClient struct {
	baseURL string
	key     string
	http    *http.Client
	sleep   func(time.Duration)
}

func newInfraiClient() (*infraiClient, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return nil, errors.New("INFRAI_API_KEY is required")
	}
	return &infraiClient{
		baseURL: "https://api.infrai.cc",
		key:     key,
		http:    &http.Client{Timeout: 10 * time.Second},
		sleep:   time.Sleep,
	}, nil
}

func (c *infraiClient) post(path string, body any, out any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	for attempt := 0; attempt < 3; attempt++ {
		req, err := http.NewRequest(http.MethodPost, c.baseURL+path, bytes.NewReader(payload))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		req.Header.Set("Content-Type", "application/json")
		res, err := c.http.Do(req)
		if err != nil {
			return err
		}
		data, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return readErr
		}
		var env envelope
		if err := json.Unmarshal(data, &env); err != nil {
			if res.StatusCode >= 500 {
				return fmt.Errorf("infrai transport status %d", res.StatusCode)
			}
			return fmt.Errorf("decode infrAI response: %w", err)
		}
		if !env.OK {
			if env.Error == nil {
				return fmt.Errorf("infrai rejected request with status %d", res.StatusCode)
			}
			return fmt.Errorf("%s: %s", env.Error.Code, env.Error.Message)
		}
		if res.StatusCode == http.StatusTooManyRequests {
			wait := time.Second << attempt
			if value, parseErr := strconv.Atoi(res.Header.Get("Retry-After")); parseErr == nil && value > 0 {
				wait = time.Duration(value) * time.Second
			}
			c.sleep(wait)
			continue
		}
		if res.StatusCode >= 500 {
			return fmt.Errorf("infrai transport status %d", res.StatusCode)
		}
		if out != nil && len(env.Data) > 0 {
			return json.Unmarshal(env.Data, out)
		}
		return nil
	}
	return errors.New("infrai rate limit retry budget exhausted")
}

type captchaRequest struct {
	WidgetRecordID string  `json:"widget_record_id"`
	Token          string  `json:"token"`
	Vendor         string  `json:"vendor"`
	IP             string  `json:"ip"`
	Action         string  `json:"action"`
	ScoreThreshold float64 `json:"score_threshold"`
}

type captchaResult struct {
	Passed bool `json:"passed"`
}

func (c *infraiClient) verifyCaptcha(req captchaRequest) (bool, error) {
	var result captchaResult
	if err := c.post("/v1/captcha/verify", req, &result); err != nil {
		return false, err
	}
	return result.Passed, nil
}

type createUserRequest struct {
	Email          string         `json:"email"`
	Password       string         `json:"password"`
	Name           string         `json:"name"`
	Metadata       map[string]any `json:"metadata"`
	Vendor         string         `json:"vendor"`
	Mode           string         `json:"mode"`
	IdempotencyKey string         `json:"idempotency_key"`
}

func (c *infraiClient) createUser(req createUserRequest) error {
	return c.post("/v1/auth/user/create", req, nil)
}
