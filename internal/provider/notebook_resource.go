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
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ resource.Resource              = &notebookResource{}
	_ resource.ResourceWithConfigure = &notebookResource{}
)

type notebookResource struct {
	client *gcp.NotebookClient
}

type notebookResourceModel struct {
	Name                   types.String                         `tfsdk:"name"`
	DisplayName            types.String                         `tfsdk:"display_name"`
	Description            types.String                         `tfsdk:"description"`
	IsDefault              types.Bool                           `tfsdk:"is_default"`
	EnableSecureBoot       types.Bool                           `tfsdk:"enable_secure_boot"`
	MachineSpec            notebookMachineSpecModel             `tfsdk:"machine_spec"`
	DataPersistentDiskSpec *notebookDataPersistentDiskSpecModel `tfsdk:"data_persistent_disk_spec"`
	NetworkSpec            notebookNetworkSpecModel             `tfsdk:"network_spec"`
	IdleShutdownConfig     *notebookIdleShutdownConfigModel     `tfsdk:"idle_shutdown_config"`
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
		MachineSpec: &gcp.MachineSpec{
			MachineType: plan.MachineSpec.MachineType.ValueStringPointer(),
		},
		NetworkSpec: &gcp.NetworkSpec{
			EnableInternetAccess: plan.NetworkSpec.EnableInternetAccess.ValueBoolPointer(),
		},
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
	state = notebookResourceModel{
		Name:        types.StringPointerValue(notebook.Name),
		DisplayName: types.StringPointerValue(notebook.DisplayName),
		Description: types.StringPointerValue(notebook.Description),
		IsDefault:   types.BoolPointerValue(notebook.IsDefault),
	}

	if notebook.ShieldedVmConfig != nil {
		state.EnableSecureBoot = types.BoolPointerValue(notebook.ShieldedVmConfig.EnableSecureBoot)
	}

	var ac64 int64
	if notebook.MachineSpec.AcceleratorCount != nil {
		ac64 = int64(*notebook.MachineSpec.AcceleratorCount)
	}

	state.MachineSpec = notebookMachineSpecModel{
		MachineType:      types.StringPointerValue(notebook.MachineSpec.MachineType),
		AcceleratorType:  types.StringPointerValue(notebook.MachineSpec.AcceleratorType),
		AcceleratorCount: basetypes.NewInt64Value(ac64),
		TpuTopofmty:      types.StringPointerValue(notebook.MachineSpec.TpuTopofmty),
	}

	// add in the Persisitent Disk Specification
	if notebook.DataPersistentDiskSpec != nil {
		state.DataPersistentDiskSpec = &notebookDataPersistentDiskSpecModel{
			DiskType:   types.StringPointerValue(notebook.DataPersistentDiskSpec.DiskType),
			DiskSizeGb: types.StringPointerValue(notebook.DataPersistentDiskSpec.DiskSizeGb),
		}
	}

	// add in the Network Specification
	if notebook.NetworkSpec != nil {
		state.NetworkSpec = notebookNetworkSpecModel{
			EnableInternetAccess: types.BoolPointerValue(notebook.NetworkSpec.EnableInternetAccess),
			Network:              types.StringPointerValue(notebook.NetworkSpec.Network),
			Subnetwork:           types.StringPointerValue(notebook.NetworkSpec.Subnetwork),
		}
	}

	// add in Idle Shutdown Configuration

	if notebook.IdleShutdownConfig != nil {
		state.IdleShutdownConfig = &notebookIdleShutdownConfigModel{
			IdleTimeout:          types.StringPointerValue(notebook.IdleShutdownConfig.IdleTimeout),
			IdleShutdownDisabled: types.BoolPointerValue(notebook.IdleShutdownConfig.IdleShutdownDisabled),
		}
	}

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
			"name": schema.StringAttribute{
				Computed: true,
			},
			"display_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Optional: true,
			},
			"is_default": schema.BoolAttribute{
				Optional: true,
			},
			"enable_secure_boot": schema.BoolAttribute{
				Optional: true,
			},
			"machine_spec": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"machine_type": schema.StringAttribute{
						Required: true,
					},
					"accelerator_type": schema.StringAttribute{
						Optional: true,
					},
					"accelerator_count": schema.Int64Attribute{
						Optional: true,
					},
					"tpu_topofmty": schema.StringAttribute{
						Optional: true,
					},
				},
			},
			"data_persistent_disk_spec": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"disk_type": schema.StringAttribute{
						Optional: true,
					},
					"disk_size_gb": schema.StringAttribute{
						Optional: true,
					},
				},
			},
			"network_spec": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"enable_internet_access": schema.BoolAttribute{
						Required: true,
					},
					"network": schema.StringAttribute{
						Optional: true,
					},
					"subnetwork": schema.StringAttribute{
						Optional: true,
					},
				},
			},
			"idle_shutdown_config": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"idle_timeout": schema.StringAttribute{
						Optional: true,
					},
					"idle_shutdown_disabled": schema.BoolAttribute{
						Optional: true,
					},
				},
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
