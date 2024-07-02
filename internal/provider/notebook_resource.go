package provider

import (
	"context"
	"fmt"
	"terraform-provider-daw/internal/gcp"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource              = &notebookResource{}
	_ resource.ResourceWithConfigure = &notebookResource{}
)

type notebookResource struct {
	client *gcp.NotebookClient
}

type notebookResourceModel struct {
	Name                   types.String                        `tfsdk:"name"`
	DisplayName            types.String                        `tfsdk:"display_name"`
	Description            types.String                        `tfsdk:"description"`
	IsDefault              types.Bool                          `tfsdk:"is_default"`
	EnableSecureBoot       types.Bool                          `tfsdk:"enable_secure_boot"`
	MachineSpec            notebookMachineSpecModel            `tfsdk:"machine_spec"`
	DataPersistentDiskSpec notebookDataPersistentDiskSpecModel `tfsdk:"data_persistent_disk_spec"`
	NetworkSpec            notebookNetworkSpecModel            `tfsdk:"network_spec"`
	IdleShutdownConfig     notebookIdleShutdownConfigModel     `tfsdk:"idle_shutdown_config"`
}

func NewNotebookResource() resource.Resource {
	return &notebookResource{}
}

// Create implements resource.Resource.
func (n *notebookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	tflog.Debug(ctx, "********* In Create *********")

	var plan notebookResourceModel
	diags := req.Plan.Get(ctx, &plan)

	tflog.Debug(ctx, "********* 0 *********")

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		tflog.Debug(ctx, "********* 0xxxE *********")
		return
	}

	tflog.Debug(ctx, "********* 1 *********")

	notebook := gcp.NotebookRuntimeTemplate{
		DisplayName: plan.DisplayName.ValueStringPointer(),
		Description: plan.Description.ValueStringPointer(),
		IsDefault:   plan.IsDefault.ValueBoolPointer(),
		MachineSpec: &gcp.MachineSpec{
			MachineType: plan.MachineSpec.MachineType.ValueStringPointer(),
		},
		NetworkSpec: &gcp.NetworkSpec{
			EnableInternetAccess: plan.NetworkSpec.EnableInternetAccess.ValueBoolPointer(),
			Network:              plan.NetworkSpec.Network.ValueStringPointer(),
			Subnetwork:           plan.NetworkSpec.Subnetwork.ValueStringPointer(),
		},
		ShieldedVmConfig: &gcp.ShieldedVmConfig{
			EnableSecureBoot: plan.EnableSecureBoot.ValueBoolPointer(),
		},
		DataPersistentDiskSpec: &gcp.DataPersistentDiskSpec{
			DiskType:   plan.DataPersistentDiskSpec.DiskType.ValueStringPointer(),
			DiskSizeGb: plan.DataPersistentDiskSpec.DiskSizeGb.ValueStringPointer(),
		},
		IdleShutdownConfig: &gcp.IdleShutdownConfig{
			IdleTimeout:          plan.IdleShutdownConfig.IdleTimeout.ValueStringPointer(),
			IdleShutdownDisabled: plan.IdleShutdownConfig.IdleShutdownDisabled.ValueBoolPointer(),
		},
	}

	tflog.Debug(ctx, fmt.Sprintf("IdleShutdownDisabled: '%+v' (%+v)",
		notebook.IdleShutdownConfig.IdleShutdownDisabled,
		*notebook.IdleShutdownConfig.IdleShutdownDisabled))

	// accelerator_type spec can be nil or set depending on machine type
	if plan.MachineSpec.AcceleratorType.ValueString() != "" {

		tflog.Debug(ctx, fmt.Sprintf("AcceleratorType is not nil but '%+v'", plan.MachineSpec.AcceleratorType.ValueStringPointer()))
		notebook.MachineSpec.AcceleratorType = plan.MachineSpec.AcceleratorType.ValueStringPointer()

		// accelerator count only make sense if acceleratory_type is set
		if plan.MachineSpec.AcceleratorCount.ValueInt64() > 0 {
			notebook.MachineSpec.AcceleratorCount = plan.MachineSpec.AcceleratorCount.ValueInt64Pointer()
		}
	}

	tflog.Debug(ctx, fmt.Sprintf("notebook: %+v", notebook))

	new_notebook, err := n.client.CreateNotebook(&notebook)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating template",
			"Could not create template, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "********* 3 *********")

	plan.Name = types.StringPointerValue(new_notebook.Name)

	tflog.Info(ctx, fmt.Sprintf("New template id: %s", plan.Name))

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete implements resource.Resource.
func (n *notebookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

	tflog.Debug(ctx, "********* In Delete *********")

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

	tflog.Debug(ctx, "********* In Read *********")

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
		DataPersistentDiskSpec: notebookDataPersistentDiskSpecModel{
			DiskType:   types.StringPointerValue(notebook.DataPersistentDiskSpec.DiskType),
			DiskSizeGb: types.StringPointerValue(notebook.DataPersistentDiskSpec.DiskSizeGb),
		},
		MachineSpec: notebookMachineSpecModel{
			MachineType:      types.StringPointerValue(notebook.MachineSpec.MachineType),
			AcceleratorType:  types.StringPointerValue(notebook.MachineSpec.AcceleratorType),
			AcceleratorCount: types.Int64PointerValue(notebook.MachineSpec.AcceleratorCount),
		},
		IdleShutdownConfig: notebookIdleShutdownConfigModel{
			IdleTimeout:          types.StringPointerValue(notebook.IdleShutdownConfig.IdleTimeout),
			IdleShutdownDisabled: types.BoolPointerValue(notebook.IdleShutdownConfig.IdleShutdownDisabled),
		},
		NetworkSpec: notebookNetworkSpecModel{
			EnableInternetAccess: types.BoolPointerValue(notebook.NetworkSpec.EnableInternetAccess),
			Network:              types.StringPointerValue(notebook.NetworkSpec.Network),
			Subnetwork:           types.StringPointerValue(notebook.NetworkSpec.Subnetwork),
		},
	}

	if notebook.ShieldedVmConfig != nil {
		state.EnableSecureBoot = types.BoolPointerValue(notebook.ShieldedVmConfig.EnableSecureBoot)
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Schema implements resource.Resource.
func (n *notebookResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {

	tflog.Debug(ctx, "********* In Schema *********")

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"display_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
				MarkdownDescription: "The Display name of the template",
			},
			"description": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"is_default": schema.BoolAttribute{
				// Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
					boolplanmodifier.RequiresReplaceIfConfigured(),
				},
				Default: booldefault.StaticBool(false),
			},
			"enable_secure_boot": schema.BoolAttribute{
				// Required: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
					boolplanmodifier.RequiresReplaceIfConfigured(),
				},
				Default: booldefault.StaticBool(false),
			},
			"machine_spec": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"machine_type": schema.StringAttribute{
						Required: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
							stringplanmodifier.RequiresReplaceIfConfigured(),
						},
					},
					"accelerator_type": schema.StringAttribute{
						Optional: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
							stringplanmodifier.RequiresReplaceIfConfigured(),
						},
					},
					"accelerator_count": schema.Int64Attribute{
						Optional: true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
							int64planmodifier.RequiresReplaceIfConfigured(),
						},
					},
				},
			},
			"data_persistent_disk_spec": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"disk_type": schema.StringAttribute{
						Required: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
							stringplanmodifier.RequiresReplaceIfConfigured(),
						},
					},
					"disk_size_gb": schema.StringAttribute{
						Required: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
							stringplanmodifier.RequiresReplaceIfConfigured(),
						},
					},
				},
			},
			"network_spec": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"enable_internet_access": schema.BoolAttribute{
						Required: true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
							boolplanmodifier.RequiresReplaceIfConfigured(),
						},
					},
					"network": schema.StringAttribute{
						Optional: true,
						// Computed: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
							stringplanmodifier.RequiresReplaceIfConfigured(),
						},
					},
					"subnetwork": schema.StringAttribute{
						Optional: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
							stringplanmodifier.RequiresReplaceIfConfigured(),
						},
					},
				},
			},
			"idle_shutdown_config": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"idle_timeout": schema.StringAttribute{
						Optional: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
							stringplanmodifier.RequiresReplaceIfConfigured(),
						},
					},
					"idle_shutdown_disabled": schema.BoolAttribute{
						Required: true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
							boolplanmodifier.RequiresReplaceIfConfigured(),
						},
					},
				},
			},
		},
	}
}

// Update implements resource.Resource.
func (n *notebookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	tflog.Debug(ctx, "********* In Update *********")

	var plan notebookResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	notebook := gcp.NotebookRuntimeTemplate{
		Name:        plan.Name.ValueStringPointer(),
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
