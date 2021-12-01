package grafana_sdk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/go-resty/resty/v2"
	"k8s.io/apimachinery/pkg/runtime"
)

type Client struct {
	baseURL     string
	key         string
	isBasicAuth bool
	username    string
	password    string
	client      *resty.Client
}

type GrafanaDashboard struct {
	Dashboard *runtime.RawExtension `json:"dashboard,omitempty"`
	FolderId  int                   `json:"folderId,omitempty"`
	FolderUid string                `json:"FolderUid,omitempty"`
	Message   string                `json:"message,omitempty"`
	Overwrite bool                  `json:"overwrite,omitempty"`
}

type GrafanaResponse struct {
	ID      *int    `json:"id,omitempty"`
	UID     *string `json:"uid,omitempty"`
	URL     *string `json:"url,omitempty"`
	Title   *string `json:"title,omitempty"`
	Message *string `json:"message,omitempty"`
	Status  *string `json:"status,omitempty"`
	Version *int    `json:"version,omitempty"`
	Slug    *string `json:"slug,omitempty"`
}

type Org struct {
	ID   *int    `json:"id,omitempty"`
	Name *string `json:"name,omitempty"`
}

type HealthResponse struct {
	Commit   string `json:"commit,omitempty"`
	Database string `json:"database,omitempty"`
	Version  string `json:"version,omitempty"`
}

// NewClient initializes client for interacting with an instance of Grafana server;
// apiKeyOrBasicAuth accepts either 'username:password' basic authentication credentials,
// or a Grafana API key. If it is an empty string then no authentication is used.
func NewClient(hostURL string, keyOrBasicAuth string) (*Client, error) {
	isBasicAuth := strings.Contains(keyOrBasicAuth, ":")
	baseURL, err := url.Parse(hostURL)
	if err != nil {
		return nil, err
	}
	client := &Client{
		baseURL:     baseURL.String(),
		key:         "",
		isBasicAuth: isBasicAuth,
		username:    "",
		password:    "",
		client:      resty.New(),
	}
	if len(keyOrBasicAuth) > 0 {
		if !isBasicAuth {
			client.key = keyOrBasicAuth
		} else {
			auths := strings.Split(keyOrBasicAuth, ":")
			if len(auths) != 2 {
				return nil, errors.New("given basic auth format is invalid. expected format: <username>:<password>")
			}
			client.username = auths[0]
			client.password = auths[1]
		}
	}
	return client, nil
}

// SetDashboard will create or update grafana dashboard
func (c *Client) SetDashboard(ctx context.Context, db *GrafanaDashboard) (*GrafanaResponse, error) {
	u, _ := url.Parse(c.baseURL)
	u.Path = path.Join(u.Path, "api/dashboards/db")
	resp, err := c.do(ctx, http.MethodPost, u.String(), db)
	if err != nil {
		return nil, err
	}
	gResp := &GrafanaResponse{}
	err = json.Unmarshal(resp.Body(), gResp)
	if err != nil {
		return nil, err
	}
	return gResp, nil
}

// DeleteDashboardByUID will delete the grafana dashboard with the given uid
func (c *Client) DeleteDashboardByUID(ctx context.Context, uid string) (*GrafanaResponse, error) {
	u, _ := url.Parse(c.baseURL)
	u.Path = path.Join(u.Path, fmt.Sprintf("api/dashboards/uid/%v", uid))
	resp, err := c.do(ctx, http.MethodDelete, u.String(), nil)
	if err != nil {
		return nil, err
	}
	gResp := &GrafanaResponse{}
	err = json.Unmarshal(resp.Body(), gResp)
	if err != nil {
		return nil, err
	}
	return gResp, nil
}

// GetCurrentOrg gets current organization.
// It reflects GET /api/org/ API call.
func (c *Client) GetCurrentOrg(ctx context.Context) (*Org, error) {
	u, _ := url.Parse(c.baseURL)
	u.Path = path.Join(u.Path, "api/org/")
	resp, err := c.do(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	org := &Org{}
	err = json.Unmarshal(resp.Body(), org)
	if err != nil {
		return nil, err
	}
	return org, nil
}

// GetHealth returns the current health status
func (c *Client) GetHealth(ctx context.Context) (*HealthResponse, error) {
	u, _ := url.Parse(c.baseURL)
	u.Path = path.Join(u.Path, "api/health")
	resp, err := c.do(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	health := &HealthResponse{}
	err = json.Unmarshal(resp.Body(), health)
	if err != nil {
		return nil, err
	}
	return health, nil
}

func (c *Client) do(ctx context.Context, method string, url string, body interface{}) (*resty.Response, error) {
	req := c.client.R().SetContext(ctx).SetBody(body)
	if c.isBasicAuth {
		req = req.SetBasicAuth(c.username, c.password)
	} else {
		req = req.SetAuthToken(c.key)
	}

	var resp *resty.Response
	var err error
	switch method {
	case http.MethodGet:
		resp, err = req.Get(url)
	case http.MethodPost:
		resp, err = req.Post(url)
	case http.MethodDelete:
		resp, err = req.Delete(url)
	case http.MethodPut:
		resp, err = req.Put(url)
	default:
		return nil, errors.New("unsupported http method")
	}
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func ReplaceDatasource(model []byte, ds string) ([]byte, error) {
	val := make(map[string]interface{})
	err := json.Unmarshal(model, &val)
	if err != nil {
		return nil, err
	}
	panels, ok := val["panels"].([]interface{})
	if !ok {
		return model, nil
	}
	var updatedPanels []interface{}
	for _, p := range panels {
		panel, ok := p.(map[string]interface{})
		if !ok {
			continue
		}
		panel["datasource"] = ds
		updatedPanels = append(updatedPanels, panel)
	}
	val["panels"] = updatedPanels

	templateList, ok := val["templating"].(map[string]interface{})
	if !ok {
		return json.Marshal(val)
	}
	templateVars, ok := templateList["list"].([]interface{})
	if !ok {
		return json.Marshal(val)
	}

	var newVars []interface{}
	for _, v := range templateVars {
		vr, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		ty, ok := vr["type"].(string)
		if !ok {
			continue
		}
		vr["datasource"] = ds
		if ty != "datasource" {
			newVars = append(newVars, vr)
		}
	}
	templateList["list"] = newVars

	return json.Marshal(val)
}
