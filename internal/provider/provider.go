// Copyright (c) Predell Services
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/arangodb/go-driver/v2/arangodb"
	"github.com/arangodb/go-driver/v2/connection"
	"net"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure ArangoProvider satisfies various provider interfaces.
var _ provider.Provider = &ArangoProvider{}
var _ provider.ProviderWithFunctions = &ArangoProvider{}

// ArangoProvider defines the provider implementation.
type ArangoProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// ArangoProviderModel describes the provider data model.
type ArangoProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	Password types.String `tfsdk:"password"`
	Tls      types.Bool   `tfsdk:"tls"`
	Username types.String `tfsdk:"username"`
}

func (p *ArangoProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "arangodb"
	resp.Version = p.version
}

func (p *ArangoProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "Endpoint url",
				Required:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password",
				Optional:            true,
			},
			"tls": schema.BoolAttribute{
				MarkdownDescription: "Enable TLS, defaults to true",
				Optional:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Username",
				Required:            true,
			},
		},
	}
}

func (p *ArangoProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data ArangoProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := connection.NewRoundRobinEndpoints([]string{data.Endpoint.ValueString()})
	conn := connection.NewHttpConnection(jsonHttpConnectionConfig(endpoint, data.Tls.IsNull() || data.Tls.ValueBool()))
	err := conn.SetAuthentication(connection.NewBasicAuth(data.Username.ValueString(), data.Password.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Authentication configuration failed", fmt.Sprintf("Authentication configuration failed: %v", err))
	}

	// Create a client
	client := arangodb.NewClient(conn)

	resp.ResourceData = &client
}

func (p *ArangoProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDatabaseResource,
		NewUserPermissionResource,
		NewUserResource,
	}
}

func (p *ArangoProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *ArangoProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ArangoProvider{
			version: version,
		}
	}
}

func jsonHttpConnectionConfig(endpoint connection.Endpoint, tlsEnabled bool) connection.HttpConfiguration {
	var tlsConfig *tls.Config = nil
	if tlsEnabled {
		tlsConfig = &tls.Config{InsecureSkipVerify: true}
	}

	return connection.HttpConfiguration{
		Endpoint:    endpoint,
		ContentType: connection.ApplicationJSON,
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 90 * time.Second,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
}
