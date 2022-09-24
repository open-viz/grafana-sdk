/*
Copyright AppsCode Inc. and Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package grafana_sdk

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/go-resty/resty/v2"
	"gomodules.xyz/pointer"
	"gomodules.xyz/x/crypto/rand"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	GrafanaTestServerURL = "http://localhost:3000/"
)

var (
	validAuth = &AuthConfig{
		BasicAuth: &BasicAuth{
			Username: "admin",
			Password: "admin",
		},
	}
	invalidAuth = &AuthConfig{
		BasicAuth: &BasicAuth{
			Username: "admin",
			Password: "wrong-pass",
		},
	}
)

func TestClient_CreateDatasource(t *testing.T) {
	type fields struct {
		baseURL string
		auth    *AuthConfig
		client  *resty.Client
	}
	type args struct {
		ctx context.Context
		ds  *Datasource
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *GrafanaResponse
		wantErr bool
	}{
		{
			name: "Create Datasource",
			fields: fields{
				baseURL: GrafanaTestServerURL,
				auth:    validAuth,
				client:  resty.New(),
			},
			args: args{
				ctx: context.TODO(),
				ds: &Datasource{
					OrgID:  1,
					Name:   "sample-ds",
					Type:   "prometheus",
					Access: "proxy",
					URL:    "http://localhost:9090/",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				baseURL: tt.fields.baseURL,
				auth:    tt.fields.auth,
				client:  tt.fields.client,
			}
			got, err := c.CreateDatasource(tt.args.ctx, tt.args.ds)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateDatasource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && *got.Message != "Datasource added" {
				t.Errorf("CreateDatasource() got = %v, want = Datasource added", *got.Message)
				return
			}
			if got.ID != nil {
				_ = deleteDS(*got.ID)
			}
		})
	}
}

func TestClient_UpdateDatasource(t *testing.T) {
	name := rand.WithUniqSuffix("update-datasource")
	ds, err := createDS(name)
	if err != nil {
		t.Errorf("failed to prepare grafana server for TestClient_CreateDatasource, reason: %v", err)
		return
	}
	defer func() {
		_ = deleteDS(pointer.Int(ds.ID))
	}()
	type fields struct {
		baseURL string
		auth    *AuthConfig
		client  *resty.Client
	}
	type args struct {
		ctx context.Context
		ds  Datasource
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *GrafanaResponse
		wantErr bool
	}{
		{
			name: "Test Update Datasource",
			fields: fields{
				baseURL: GrafanaTestServerURL,
				auth:    validAuth,
				client:  resty.New(),
			},
			args: args{
				ctx: nil,
				ds: Datasource{
					OrgID:  1,
					ID:     uint(pointer.Int(ds.ID)),
					Name:   name,
					Type:   "prometheus",
					Access: "proxy",
					URL:    "http://127.0.0.1:9090/",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				baseURL: tt.fields.baseURL,
				auth:    tt.fields.auth,
				client:  tt.fields.client,
			}
			got, err := c.UpdateDatasource(tt.args.ctx, tt.args.ds)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateDatasource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && *got.Message != "Datasource updated" {
				t.Errorf("UpdateDatasource() got = %v, want Datasource updated", *got.Message)
			}
		})
	}
}

func TestClient_DeleteDatasource(t *testing.T) {
	name := rand.WithUniqSuffix("delete-datasource")
	ds, err := createDS(name)
	if err != nil {
		t.Errorf("failed to prepare grafana server for TestClient_CreateDatasource, reason: %v", err)
		return
	}
	defer func() {
		_ = deleteDS(pointer.Int(ds.ID))
	}()
	type fields struct {
		baseURL string
		auth    *AuthConfig
		client  *resty.Client
	}
	type args struct {
		ctx context.Context
		id  int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *GrafanaResponse
		wantErr bool
	}{
		{
			name: "Delete Datasource",
			fields: fields{
				baseURL: GrafanaTestServerURL,
				auth:    validAuth,
				client:  resty.New(),
			},
			args: args{
				ctx: context.TODO(),
				id:  pointer.Int(ds.ID),
			},
			wantErr: false,
		},
		{
			name: "Delete Datasource with same id",
			fields: fields{
				baseURL: GrafanaTestServerURL,
				auth:    validAuth,
				client:  resty.New(),
			},
			args: args{
				ctx: context.TODO(),
				id:  pointer.Int(ds.ID),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				baseURL: tt.fields.baseURL,
				auth:    tt.fields.auth,
				client:  tt.fields.client,
			}
			got, err := c.DeleteDatasource(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				fmt.Println(err.Error())
				t.Errorf("DeleteDatasource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && *got.Message != "Data source deleted" {
				t.Errorf("DeleteDatasource() got = %v, want = Data source deleted", *got.Message)
				return
			}
		})
	}
}

func TestClient_GetCurrentOrg(t *testing.T) {
	type fields struct {
		baseURL string
		auth    *AuthConfig
		client  *resty.Client
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Org
		wantErr bool
	}{
		{
			name: "Get CurrentOrg",
			fields: fields{
				baseURL: GrafanaTestServerURL,
				auth:    validAuth,
				client:  resty.New(),
			},
			args: args{
				ctx: context.TODO(),
			},
			want: &Org{
				ID:   pointer.IntP(1),
				Name: pointer.StringP("Main Org."),
			},
			wantErr: false,
		},
		{
			name: "Get CurrentOrg with invalid auth",
			fields: fields{
				baseURL: GrafanaTestServerURL,
				auth:    invalidAuth,
				client:  resty.New(),
			},
			args: args{
				ctx: context.TODO(),
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				baseURL: tt.fields.baseURL,
				auth:    tt.fields.auth,
				client:  tt.fields.client,
			}
			got, err := c.GetCurrentOrg(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCurrentOrg() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetCurrentOrg() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_GetHealth(t *testing.T) {
	type fields struct {
		baseURL string
		auth    *AuthConfig
		client  *resty.Client
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *HealthResponse
		wantErr bool
	}{
		{
			name: "Test Get Health API",
			fields: fields{
				baseURL: GrafanaTestServerURL,
				auth:    validAuth,
				client:  resty.New(),
			},
			args: args{
				ctx: context.TODO(),
			},
			wantErr: false,
		},
		{
			name: "Test Get Health API without auth",
			fields: fields{
				baseURL: GrafanaTestServerURL,
				client:  resty.New(),
			},
			args: args{
				ctx: context.TODO(),
			},
			wantErr: false, // as health api is public api and don't need authentication
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				baseURL: tt.fields.baseURL,
				auth:    tt.fields.auth,
				client:  tt.fields.client,
			}
			got, err := c.GetHealth(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetHealth() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.Database != "ok" {
				t.Errorf("GetHealth() got = %v, want ok", got.Database)
			}
		})
	}
}

func TestClient_SetDashboard(t *testing.T) {
	type fields struct {
		baseURL string
		auth    *AuthConfig
		client  *resty.Client
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		wantErr      bool
		jsonFilePath string
	}{
		{
			name: "Test Set Dashboard by creating new dashboard",
			fields: fields{
				baseURL: GrafanaTestServerURL,
				auth:    validAuth,
				client:  resty.New(),
			},
			args: args{
				ctx: context.TODO(),
			},
			wantErr:      false,
			jsonFilePath: "./testdata/dashboard.yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				baseURL: tt.fields.baseURL,
				auth:    tt.fields.auth,
				client:  tt.fields.client,
			}
			model, err := os.ReadFile(tt.jsonFilePath)
			if err != nil {
				t.Errorf("failed to read json model, reason: %v", err)
				return
			}
			db := &GrafanaDashboard{
				Dashboard: &runtime.RawExtension{Raw: model},
				FolderId:  0,
				Overwrite: true,
			}
			got, err := c.SetDashboard(tt.args.ctx, db)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetDashboard() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && *got.Status != "success" {
				t.Errorf("SetDashboard() Status got = %v, want success", *got.Status)
			}
		})
	}
}

func TestClient_DeleteDashboardByUID(t *testing.T) {
	db, err := createDB("./testdata/dashboard.yaml")
	if err != nil {
		t.Errorf("failed to prepare grafana server for TestClient_CreateDatasource, reason: %v", err)
		return
	}
	type fields struct {
		baseURL string
		auth    *AuthConfig
		client  *resty.Client
	}
	type args struct {
		ctx context.Context
		uid string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *GrafanaResponse
		wantErr bool
	}{
		{
			name: "Delete Dashboard",
			fields: fields{
				baseURL: GrafanaTestServerURL,
				auth:    validAuth,
				client:  resty.New(),
			},
			args: args{
				ctx: context.TODO(),
				uid: pointer.String(db.UID),
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				baseURL: tt.fields.baseURL,
				auth:    tt.fields.auth,
				client:  tt.fields.client,
			}
			_, err := c.DeleteDashboardByUID(tt.args.ctx, tt.args.uid)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteDashboardByUID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func createDS(name string) (*GrafanaResponse, error) {
	client, err := NewClient(GrafanaTestServerURL, validAuth)
	if err != nil {
		return nil, err
	}
	ds, err := client.CreateDatasource(context.TODO(), &Datasource{
		OrgID:     1,
		Name:      name,
		Type:      "prometheus",
		Access:    "proxy",
		URL:       "http://localhost:9090",
		IsDefault: true,
	})
	if err != nil {
		return nil, err
	}
	return ds, nil
}

func deleteDS(id int) error {
	client, err := NewClient(GrafanaTestServerURL, validAuth)
	if err != nil {
		return err
	}
	_, err = client.DeleteDatasource(context.TODO(), id)
	return err
}

func createDB(dbFilePath string) (*GrafanaResponse, error) {
	client, err := NewClient(GrafanaTestServerURL, validAuth)
	if err != nil {
		return nil, err
	}
	model, err := os.ReadFile(dbFilePath)
	if err != nil {
		return nil, err
	}
	db := &GrafanaDashboard{
		Dashboard: &runtime.RawExtension{Raw: model},
		FolderId:  0,
		Overwrite: true,
	}
	resp, err := client.SetDashboard(context.TODO(), db)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
