package provider

import (
	"context"
	"fmt"
	"terraform-provider-daw/internal/gcp"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = &notebookResource{}
	_ resource.ResourceWithConfigure = &notebookResource{}
)

type notebookResource struct {
	client *gcp.NotebookClient
}

type notebookResourceModel struct {
	Name        types.String `tfsdk:"name"`
	DisplayName types.String `tfsdk:"display_name"`
	Description types.String `tfsdk:"description"`
}

func NewNotebookResource() resource.Resource {
	return &notebookResource{}
}

// Create implements resource.Resource.
func (n *notebookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	var plan notebookResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	notebook := gcp.NotebookRuntimeTemplate{
		DisplayName: plan.DisplayName.ValueStringPointer(),
		Description: plan.Description.ValueStringPointer(),
	}

	new_notebook, err := n.client.CreateNotebook(&notebook)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating template",
			"Could not create template, unexpected error: "+err.Error(),
		)
		return
	}

	plan.Name = types.StringPointerValue(new_notebook.Name)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete implements resource.Resource.
func (n *notebookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

	var state notebookResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := n.client.DeleteNotebookRuntimeTemplate(state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting template",
			"Could not delete template, unexpected error: "+err.Error(),
		)
		return
	}
}

// Metadata implements resource.Resource.
func (n *notebookResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notebook"
}

// Read implements resource.Resource.
func (n *notebookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

	var state notebookResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	notebook, err := n.client.GetNotebook(state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading GCP Notebooks",
			"Could not read Notebook with name "+state.Name.ValueString()+": "+err.Error(),
		)
		return
	}

	// Overwrite with refreshed state
	state.Name = types.StringPointerValue(notebook.Name)
	state.DisplayName = types.StringPointerValue(notebook.DisplayName)
	state.Description = types.StringPointerValue(notebook.Description)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Schema implements resource.Resource.
func (n *notebookResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"display_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

// Update implements resource.Resource.
func (n *notebookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	var plan notebookResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	notebook := gcp.NotebookRuntimeTemplate{
		DisplayName: plan.DisplayName.ValueStringPointer(),
		Description: plan.Description.ValueStringPointer(),
	}

	new_notebook, err := n.client.UpdateNotebook(&notebook)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating template",
			"Could not create template, unexpected error: "+err.Error(),
		)
		return
	}

	plan.Name = types.StringPointerValue(new_notebook.Name)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (n *notebookResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*gcp.NotebookClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *gcp.NotebookClient, got: %T", req.ProviderData),
		)

		return
	}

	n.client = client
}
