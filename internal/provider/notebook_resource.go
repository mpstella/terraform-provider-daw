package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/mpstella/terraform-provider-daw/internal/gcp"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource              = &notebookResource{}
	_ resource.ResourceWithConfigure = &notebookResource{}
)

// just making alias to not get confused
type notebookResource gcpNotebookClient

func (n *notebookResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {

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

func (n notebookResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {

	var data notebookModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// check that these values don't conflict
	if data.IdleShutdownConfig.IdleShutdownDisabled.ValueBool() && !data.IdleShutdownConfig.IdleTimeout.IsNull() {
		resp.Diagnostics.AddAttributeWarning(
			path.Root("idle_shutdown_disabled"),
			"idle_shutdown_disabled can't be set to True and have a value in idle_timeout",
			"Expected idle_timeout to be nil if idle_shutdown_disabled is set to True.",
		)
	}

	// check the idleTimeout values are correctly applied
	if !data.IdleShutdownConfig.IdleTimeout.IsNull() {
		val := data.IdleShutdownConfig.IdleTimeout.ValueString()
		if !strings.HasSuffix(val, "s") {
			resp.Diagnostics.AddAttributeWarning(
				path.Root("idle_timeout"),
				"idle_timeout must end in 's'",
				"Expected idle_timeout end in 's' as it is defined in seconds",
			)
		}
		val = strings.TrimSuffix(val, "s")
		num, err := strconv.Atoi(val)
		if err != nil {
			resp.Diagnostics.AddAttributeWarning(
				path.Root("idle_timeout"),
				"idle_timeout must end in 's' and be a valid integer",
				"Expected idle_timeout end in 's' as it is defined in seconds as an integer",
			)
		} else {
			if num < 600 || num > 86400 {
				// [600, 86400]
				resp.Diagnostics.AddAttributeWarning(
					path.Root("idle_timeout"),
					"idle_timeout must end in 's' and be a valid integer between 600 and 86400",
					"Expected idle_timeout to be between 600 and 86400",
				)
			}
		}
	}

	// if enable_internet_access is false then both network and subdomain need to be set
	if !data.NetworkSpec.EnableInternetAccess.ValueBool() {
		if data.NetworkSpec.Network.IsNull() {
			resp.Diagnostics.AddAttributeWarning(
				path.Root("network"),
				"network can't be nil if enable_internet_access is false",
				"Expected network to be configured",
			)
		}
		if data.NetworkSpec.Subnetwork.IsNull() {
			resp.Diagnostics.AddAttributeWarning(
				path.Root("subnetwork"),
				"subnetwork can't be nil if enable_internet_access is false",
				"Expected subnetwork to be configured",
			)
		}
	}

	// check machine type accelerator settings
	if data.MachineSpec.AcceleratorType.IsNull() && !data.MachineSpec.AcceleratorCount.IsNull() {
		resp.Diagnostics.AddAttributeWarning(
			path.Root("accelerator_count"),
			"accelerator_count must be nil if accelerator_type is nil",
			"Expected accelerator_count to not be configured",
		)
	}
}

func NewNotebookResource() resource.Resource {
	return &notebookResource{}
}

// Create implements resource.Resource.
func (n *notebookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	tflog.Debug(ctx, "********* In Create(notebook_resource) *********")

	var plan notebookModel
	diags := req.Plan.Get(ctx, &plan)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

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

	// accelerator_type spec can be nil or set depending on machine type
	if !plan.MachineSpec.AcceleratorType.IsNull() {

		notebook.MachineSpec.AcceleratorType = plan.MachineSpec.AcceleratorType.ValueStringPointer()

		// accelerator count only make sense if acceleratory_type is set
		if plan.MachineSpec.AcceleratorCount.ValueInt64() > 0 {
			notebook.MachineSpec.AcceleratorCount = plan.MachineSpec.AcceleratorCount.ValueInt64Pointer()
		}
	}

	if !plan.KmsKeyName.IsNull() {
		notebook.EncryptionSpec = &gcp.EncryptionSpec{
			KmsKeyName: plan.KmsKeyName.ValueStringPointer(),
		}
	}

	if !plan.Labels.IsNull() {

		tflog.Debug(ctx, "********* In Create(plan.Labels IsNotNull()) *********")

		labels := make(map[string]string)
		plan.Labels.ElementsAs(ctx, &labels, false)
		notebook.Labels = &labels
	} else {
		tflog.Debug(ctx, "********* In Create(plan.Labels IsNull()) *********")
	}

	nas, _ := notebook.AsString()
	tflog.Debug(ctx, nas)

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

	tflog.Debug(ctx, "********* In Delete(notebook_resource) *********")

	var state notebookModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// going to ignore deletes as this only occurs when resource has already been deleted
	n.client.DeleteNotebookRuntimeTemplate(state.Name.ValueString())
}

// Metadata implements resource.Resource.
func (n *notebookResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notebook"
}

// Read implements resource.Resource.
func (n *notebookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

	tflog.Debug(ctx, "********* In Read(notebook_resource) *********")

	var state notebookModel

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
	state = notebookModel{
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

	if notebook.IsDefault == nil {
		state.IsDefault = types.BoolValue(false)
	}
	if notebook.IdleShutdownConfig.IdleShutdownDisabled == nil {
		state.IdleShutdownConfig.IdleShutdownDisabled = types.BoolValue(false)
	}
	if notebook.ShieldedVmConfig.EnableSecureBoot == nil {
		state.EnableSecureBoot = types.BoolValue(false)
	}

	if notebook.EncryptionSpec == nil {
		state.KmsKeyName = types.StringNull()
	} else {
		state.KmsKeyName = types.StringPointerValue(notebook.EncryptionSpec.KmsKeyName)
	}

	if notebook.Labels == nil {
		state.Labels = types.MapNull(types.StringType)
	} else {
		var diags diag.Diagnostics
		state.Labels, diags = basetypes.NewMapValueFrom(ctx, types.StringType, *notebook.Labels)
		resp.Diagnostics.Append(diags...)
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

	tflog.Debug(ctx, "********* In Schema(notebook_resource) *********")

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
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
					boolplanmodifier.RequiresReplaceIfConfigured(),
				},
				Default: booldefault.StaticBool(false),
			},
			"enable_secure_boot": schema.BoolAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
					boolplanmodifier.RequiresReplaceIfConfigured(),
				},
				Default: booldefault.StaticBool(false),
			},
			"kms_key_name": schema.StringAttribute{
				Optional: true,
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
				Computed: true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{
					"enable_internet_access": schema.BoolAttribute{
						Required: true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
							boolplanmodifier.RequiresReplaceIfConfigured(),
						},
					},
					"network": schema.StringAttribute{
						Required: true,
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
				Optional: true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{
					"idle_timeout": schema.StringAttribute{
						Optional: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
							stringplanmodifier.RequiresReplaceIfConfigured(),
						},
					},
					"idle_shutdown_disabled": schema.BoolAttribute{
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
							boolplanmodifier.RequiresReplaceIfConfigured(),
						},
						Default: booldefault.StaticBool(false),
					},
				},
			},
			"labels": schema.MapAttribute{
				Description: "A set of key/value label pairs to assign to the resource.",
				Optional:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.UseStateForUnknown(),
					mapplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
		},
	}
}

// Update implements resource.Resource.
func (n *notebookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	tflog.Debug(ctx, "********* In Update(notebook_resource) *********")

	var plan notebookModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	notebook := gcp.NotebookRuntimeTemplate{
		Name: plan.Name.ValueStringPointer(),
		EncryptionSpec: &gcp.EncryptionSpec{
			KmsKeyName: plan.KmsKeyName.ValueStringPointer(),
		},
	}

	err := n.client.UpdateNotebook(&notebook)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating template",
			"Could update template, unexpected error: "+err.Error(),
		)
		return
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
