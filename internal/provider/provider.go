package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func init() {
	schema.DescriptionKind = schema.StringMarkdown
}

func New(version string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				"api_key_id": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("TG_API_KEY_ID", nil),
				},
				"api_key_secret": {
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
					DefaultFunc: schema.EnvDefaultFunc("TG_API_KEY_SECRET", nil),
				},
				"api_host": {
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
					DefaultFunc: schema.EnvDefaultFunc("TG_API_HOST", "api.trustgrid.io"),
				},
			},
			DataSourcesMap: map[string]*schema.Resource{
				"tg_node": dataSourceNode(),
			},
			ResourcesMap: map[string]*schema.Resource{
				"tg_compute_limits": cpuLimitsResource(),
				"tg_snmp":           snmpResource(),
			},
		}

		p.ConfigureContextFunc = configure(version, p)

		return p
	}
}

type tgClient struct {
	APIKey    string
	APISecret string
	APIHost   string
}

func (tg *tgClient) put(ctx context.Context, url string, payload interface{}) error {
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

func (tg *tgClient) get(ctx context.Context, url string, out interface{}) error {
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

func configure(version string, p *schema.Provider) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		return &tgClient{
			APIKey:    d.Get("api_key_id").(string),
			APISecret: d.Get("api_key_secret").(string),
			APIHost:   d.Get("api_host").(string),
		}, nil
	}
}
