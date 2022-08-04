package tg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

type Client struct {
	APIKey    string
	APISecret string
	APIHost   string
}

func (tg *Client) Delete(ctx context.Context, url string, payload interface{}) error {
	body, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("couldn't marshal body: %s", err)
	}
	b := bytes.NewBuffer(body)

	req, err := http.NewRequest("DELETE", fmt.Sprintf("https://%s/%s", tg.APIHost, strings.TrimPrefix(url, "/")), b)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("trustgrid-token %s:%s", tg.APIKey, tg.APISecret))
	req.Header.Set("Accept", "application/json")

	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	if r.StatusCode != 200 {
		reply, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return fmt.Errorf("non-200 from portal: %d; couldn't read body: %s", r.StatusCode, err)
		}
		return fmt.Errorf("non-200 from portal: %d\npayload:\n%s\n\nreply:\n%s", r.StatusCode, string(body), reply)
	}

	return nil
}

func (tg *Client) Post(ctx context.Context, url string, payload interface{}) error {
	body, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("couldn't marshal body: %s", err)
	}
	b := bytes.NewBuffer(body)

	req, err := http.NewRequest("POST", fmt.Sprintf("https://%s/%s", tg.APIHost, strings.TrimPrefix(url, "/")), b)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("trustgrid-token %s:%s", tg.APIKey, tg.APISecret))
	req.Header.Set("Accept", "application/json")

	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	if r.StatusCode != 200 {
		reply, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return fmt.Errorf("non-200 from portal: %d; couldn't read body: %s", r.StatusCode, err)
		}
		return fmt.Errorf("non-200 from portal: %d\npayload:\n%s\n\nreply:\n%s", r.StatusCode, string(body), reply)
	}

	return nil
}

func (tg *Client) Put(ctx context.Context, url string, payload interface{}) error {
	body, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("couldn't marshal body: %s", err)
	}
	b := bytes.NewBuffer(body)

	req, err := http.NewRequest("PUT", fmt.Sprintf("https://%s/%s", tg.APIHost, strings.TrimPrefix(url, "/")), b)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("trustgrid-token %s:%s", tg.APIKey, tg.APISecret))
	req.Header.Set("Accept", "application/json")

	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	if r.StatusCode != 200 {
		reply, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return fmt.Errorf("non-200 from portal: %d; couldn't read body: %s", r.StatusCode, err)
		}
		return fmt.Errorf("non-200 from portal: %d\npayload:\n%s\n\nreply:\n%s", r.StatusCode, string(body), reply)
	}

	return nil
}

func (tg *Client) RawGet(ctx context.Context, url string) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://%s/%s", tg.APIHost, strings.TrimPrefix(url, "/")), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("trustgrid-token %s:%s", tg.APIKey, tg.APISecret))

	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if r.StatusCode != 200 {
		return r.Body, fmt.Errorf("non-200 from portal: %d; couldn't read body: %s", r.StatusCode, err)
	}

	return r.Body, nil
}

func (tg *Client) Get(ctx context.Context, url string, out interface{}) error {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://%s/%s", tg.APIHost, strings.TrimPrefix(url, "/")), nil)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("trustgrid-token %s:%s", tg.APIKey, tg.APISecret))
	req.Header.Set("Accept", "application/json")

	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	if r.StatusCode != 200 {
		reply, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return fmt.Errorf("non-200 from portal: %d; couldn't read body: %s", r.StatusCode, err)
		}
		return fmt.Errorf("non-200 from portal: %d - %s\n%s", r.StatusCode, req.URL.String(), reply)
	}

	reply, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("error reading reply: %s", err)
	}

	err = json.Unmarshal(reply, out)
	if err != nil {
		return fmt.Errorf("error decoding json: %s\n\nreply:\n%s", err, string(reply))
	}

	return nil
}
