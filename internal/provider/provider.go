// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"

	"github.com/mpstella/terraform-provider-daw/internal/gcp"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure notebookProvider satisfies various provider interfaces.
var _ provider.Provider = &notebookProvider{}

// notebookProvider defines the provider implementation.
type notebookProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

type notebookProviderModel struct {
	Project  types.String `tfsdk:"project"`
	Location types.String `tfsdk:"location"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &notebookProvider{
			version: version,
		}
	}
}

func (p *notebookProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "daw"
	resp.Version = p.version
}

func (p *notebookProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"project": schema.StringAttribute{
				Optional: true,
			},
			"location": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func (p *notebookProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config notebookProviderModel

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.Project.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("project"),
			"Unknown Project",
			"The provider cannot create the client as there is an unknown configuration value for project",
		)
	}

	if config.Location.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("location"),
			"Unknown Location",
			"The provider cannot create the client as there is an unknown configuration value for location",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// provide some common defaults
	project := os.Getenv("CLOUDSDK_CORE_PROJECT")
	location := os.Getenv("CLOUDSDK_COMPUTE_REGION")

	if !config.Project.IsNull() {
		project = config.Project.ValueString()
	}

	if !config.Location.IsNull() {
		location = config.Location.ValueString()
	}

	if project == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("project"),
			"Missing project",
			"The provider cannot create the client as there is a missing or empty value for the GCP Project",
		)
	}

	if location == "" {
		location = "australia-southeast1"
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "gcp_project", project)
	ctx = tflog.SetField(ctx, "gcp_location", location)
	tflog.Debug(ctx, "Creating GCP client")

	client, err := gcp.NewNotebookClient(project, location)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create GCP Client",
			"An unexpected error occurred when creating the GCP client:"+err.Error(),
		)
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configured GCP client", map[string]any{"success": true})
}

func (p *notebookProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewNotebookResource,
	}
}

func (p *notebookProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewNotebookDataSource,
	}
}
