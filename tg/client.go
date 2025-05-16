package tg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

type Client struct {
	Lock      sync.Mutex
	writeLock sync.Mutex

	APIHost   string
	APIKey    string
	APISecret string
	JWT       string

	Domain string
}

type NotFoundError struct {
	URL string
}

type ValidationError struct {
	URL string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s", e.URL)
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("not found: %s", e.URL)
}

type ClientParams struct {
	APIKey    string
	APISecret string
	APIHost   string
	JWT       string
	OrgID     string
}

func NewClient(ctx context.Context, params ClientParams) (*Client, error) {
	client := &Client{
		APIKey:    params.APIKey,
		APISecret: params.APISecret,
		APIHost:   params.APIHost,
		JWT:       params.JWT,
	}

	org := Org{}
	err := client.Get(ctx, "/org/mine", &org)
	if err != nil {
		return client, fmt.Errorf("error retrieving org info: %w", err)
	}
	if params.OrgID != "" && params.OrgID != org.UID {
		return client, fmt.Errorf("org id mismatch: %s != %s", params.OrgID, org.UID)
	}

	client.Domain = org.Domain

	return client, nil
}

func GetClient(meta any) *Client {
	tgc, ok := meta.(*Client)
	if !ok {
		panic("meta is not a *tg.Client")
	}
	return tgc
}

func (tg *Client) authHeader() string {
	if tg.JWT != "" {
		return fmt.Sprintf("Bearer %s", tg.JWT)
	}
	return fmt.Sprintf("trustgrid-token %s:%s", tg.APIKey, tg.APISecret)
}

func (tg *Client) Delete(ctx context.Context, url string, payload any) error {
	tg.writeLock.Lock()
	defer tg.writeLock.Unlock()

	body, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("couldn't marshal body: %w", err)
	}
	b := bytes.NewBuffer(body)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, fmt.Sprintf("https://%s/%s", tg.APIHost, strings.TrimPrefix(url, "/")), b)
	if err != nil {
		return err
	}

	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Authorization", tg.authHeader())
	req.Header.Set("Accept", "application/json")

	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		reply, err := io.ReadAll(r.Body)
		if err != nil {
			return fmt.Errorf("[DELETE]: non-200 from portal (%s): %d; couldn't read body: %w", url, r.StatusCode, err)
		}
		return fmt.Errorf("[DELETE]: non-200 from portal (%s): %d\npayload:\n%s\n\nreply:\n%s", url, r.StatusCode, string(body), reply)
	}

	return nil
}

func (tg *Client) Post(ctx context.Context, url string, payload any) ([]byte, error) {
	tg.writeLock.Lock()
	defer tg.writeLock.Unlock()

	body, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("couldn't marshal body: %w", err)
	}
	b := bytes.NewBuffer(body)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("https://%s/%s", tg.APIHost, strings.TrimPrefix(url, "/")), b)
	if err != nil {
		return nil, err
	}

	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Authorization", tg.authHeader())
	req.Header.Set("Accept", "application/json")

	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	reply, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("[POST] couldn't read body: %w", err)
	}
	if r.StatusCode != http.StatusOK {
		return reply, fmt.Errorf("[POST] non-200 from portal (%s): %d\npayload:\n%s\n\nreply:\n%s", url, r.StatusCode, string(body), reply)
	}

	return reply, nil
}

func (tg *Client) Put(ctx context.Context, url string, payload any) ([]byte, error) {
	tg.writeLock.Lock()
	defer tg.writeLock.Unlock()

	body, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("couldn't marshal body: %w", err)
	}
	b := bytes.NewBuffer(body)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, fmt.Sprintf("https://%s/%s", tg.APIHost, strings.TrimPrefix(url, "/")), b)
	if err != nil {
		return nil, err
	}

	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Authorization", tg.authHeader())
	req.Header.Set("Accept", "application/json")

	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	reply, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("[PUT] couldn't read body: %w", err)
	}
	if r.StatusCode != http.StatusOK {
		return reply, fmt.Errorf("[PUT] non-200 from portal (%s): %d\npayload:\n%s\n\nreply:\n%s", url, r.StatusCode, string(body), reply)
	}

	return reply, nil
}

func (tg *Client) RawGet(ctx context.Context, url string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("https://%s/%s", tg.APIHost, strings.TrimPrefix(url, "/")), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", tg.authHeader())

	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if r.StatusCode != http.StatusOK {
		switch r.StatusCode {
		case http.StatusNotFound:
			return r.Body, &NotFoundError{URL: url}
		case http.StatusUnprocessableEntity:
			return r.Body, &ValidationError{URL: url}
		default:
			return r.Body, fmt.Errorf("[RAWGET] non-200 from portal (%s): %d; couldn't read body: %w", url, r.StatusCode, err)
		}
	}

	return r.Body, nil
}

func (tg *Client) Get(ctx context.Context, url string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("https://%s/%s", tg.APIHost, strings.TrimPrefix(url, "/")), nil)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", tg.authHeader())
	req.Header.Set("Accept", "application/json")

	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		reply, err := io.ReadAll(r.Body)
		if err != nil {
			return fmt.Errorf("[GET] non-200 from portal (%s): %d; couldn't read body: %w", url, r.StatusCode, err)
		}
		if r.StatusCode == http.StatusNotFound {
			return &NotFoundError{URL: url}
		}
		return fmt.Errorf("[GET] non-200 from portal (%s): %d - %s\n%s", url, r.StatusCode, req.URL.String(), reply)
	}

	reply, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("error reading reply: %w", err)
	}

	err = json.Unmarshal(reply, out)
	if err != nil {
		return fmt.Errorf("error decoding json: %w\n\nreply:\n%s", err, string(reply))
	}

	return nil
}
