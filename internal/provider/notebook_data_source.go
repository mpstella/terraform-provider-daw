package provider

import (
	"context"
	"fmt"
	"terraform-provider-daw/internal/gcp"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &notebookDataSource{}
	_ datasource.DataSourceWithConfigure = &notebookDataSource{}
)

type notebookDataSource struct {
	client *gcp.NotebookClient
}

type notebookModel struct {
	Name                   types.String                        `tfsdk:"name"`
	DisplayName            types.String                        `tfsdk:"display_name"`
	Description            types.String                        `tfsdk:"description"`
	IsDefault              types.Bool                          `tfsdk:"is_default"`
	EnableSecureBoot       types.Bool                          `tfsdk:"enable_secure_boot"`
	KmsKeyName             types.String                        `tfsdk:"kms_key_name"`
	MachineSpec            notebookMachineSpecModel            `tfsdk:"machine_spec"`
	DataPersistentDiskSpec notebookDataPersistentDiskSpecModel `tfsdk:"data_persistent_disk_spec"`
	NetworkSpec            notebookNetworkSpecModel            `tfsdk:"network_spec"`
	IdleShutdownConfig     notebookIdleShutdownConfigModel     `tfsdk:"idle_shutdown_config"`
}

type notebookMachineSpecModel struct {
	MachineType      types.String `tfsdk:"machine_type"`
	AcceleratorType  types.String `tfsdk:"accelerator_type"`
	AcceleratorCount types.Int64  `tfsdk:"accelerator_count"`
}

type notebookDataPersistentDiskSpecModel struct {
	DiskType   types.String `tfsdk:"disk_type"`
	DiskSizeGb types.String `tfsdk:"disk_size_gb"`
}

type notebookNetworkSpecModel struct {
	EnableInternetAccess types.Bool   `tfsdk:"enable_internet_access"`
	Network              types.String `tfsdk:"network"`
	Subnetwork           types.String `tfsdk:"subnetwork"`
}

type notebookIdleShutdownConfigModel struct {
	IdleTimeout          types.String `tfsdk:"idle_timeout"`
	IdleShutdownDisabled types.Bool   `tfsdk:"idle_shutdown_disabled"`
}

type notebookDataSourceModel struct {
	Notebooks []notebookModel `tfsdk:"notebooks"`
}

// Configure implements datasource.DataSourceWithConfigure.
func (n *notebookDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {

	tflog.Debug(ctx, "********* In Configure (notebook_data_source) *********")

	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*gcp.NotebookClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *gcp.NotebookClient, got %T", req.ProviderData),
		)
		return
	}
	n.client = client
}

func NewNotebookDataSource() datasource.DataSource {
	return &notebookDataSource{}
}

// Metadata implements datasource.DataSource.
func (n *notebookDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notebook"
}

// Read implements datasource.DataSource.
func (n *notebookDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	tflog.Debug(ctx, "********* In Read (notebook_data_source) *********")

	var state notebookDataSourceModel

	notebooks, err := n.client.GetNotebooks()

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read GCP Notebooks",
			err.Error(),
		)
		return
	}

	for _, notebook := range notebooks.NotebookRuntimeTemplates {

		n, _ := notebook.AsString()

		tflog.Debug(ctx, "********* notebook *********", map[string]interface{}{"notebook": n})

		notebookState := notebookModel{
			Name:        types.StringPointerValue(notebook.Name),
			DisplayName: types.StringPointerValue(notebook.DisplayName),
			Description: types.StringPointerValue(notebook.Description),
			IsDefault:   types.BoolPointerValue(notebook.IsDefault),
			DataPersistentDiskSpec: notebookDataPersistentDiskSpecModel{
				DiskType:   types.StringPointerValue(notebook.DataPersistentDiskSpec.DiskType),
				DiskSizeGb: types.StringPointerValue(notebook.DataPersistentDiskSpec.DiskSizeGb),
			},

			NetworkSpec: notebookNetworkSpecModel{
				EnableInternetAccess: types.BoolPointerValue(notebook.NetworkSpec.EnableInternetAccess),
				Network:              types.StringPointerValue(notebook.NetworkSpec.Network),
				Subnetwork:           types.StringPointerValue(notebook.NetworkSpec.Subnetwork),
			},

			IdleShutdownConfig: notebookIdleShutdownConfigModel{
				IdleShutdownDisabled: types.BoolPointerValue(notebook.IdleShutdownConfig.IdleShutdownDisabled),
				IdleTimeout:          types.StringPointerValue(notebook.IdleShutdownConfig.IdleTimeout),
			},

			MachineSpec: notebookMachineSpecModel{
				MachineType:      types.StringPointerValue(notebook.MachineSpec.MachineType),
				AcceleratorType:  types.StringPointerValue(notebook.MachineSpec.AcceleratorType),
				AcceleratorCount: types.Int64PointerValue(notebook.MachineSpec.AcceleratorCount),
			},
		}

		if notebook.ShieldedVmConfig != nil {
			notebookState.EnableSecureBoot = types.BoolPointerValue(notebook.ShieldedVmConfig.EnableSecureBoot)
		}

		// might be nil still, set to false
		if notebookState.EnableSecureBoot.IsNull() {
			notebookState.EnableSecureBoot = types.BoolValue(false)
		}

		if notebook.EncryptionSpec != nil {
			notebookState.KmsKeyName = types.StringPointerValue(notebook.EncryptionSpec.KmsKeyName)
		}

		if notebook.NetworkSpec.EnableInternetAccess == nil {
			notebookState.NetworkSpec.EnableInternetAccess = types.BoolValue(false)
		}

		if notebook.IsDefault == nil {
			notebookState.IsDefault = types.BoolValue(false)
		}

		if notebook.IdleShutdownConfig.IdleShutdownDisabled == nil {
			notebookState.IdleShutdownConfig.IdleShutdownDisabled = types.BoolValue(false)
		}

		state.Notebooks = append(state.Notebooks, notebookState)
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Schema implements datasource.DataSource.
func (n *notebookDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {

	tflog.Debug(ctx, "********* In Schema (notebook_data_source) *********")

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"notebooks": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed: true,
						},
						"display_name": schema.StringAttribute{
							Computed: true,
						},
						"description": schema.StringAttribute{
							Computed: true,
						},
						"is_default": schema.BoolAttribute{
							Computed: true,
						},
						"enable_secure_boot": schema.BoolAttribute{
							Computed: true,
						},
						"kms_key_name": schema.StringAttribute{
							Optional: true,
							Computed: true,
						},
						"machine_spec": schema.SingleNestedAttribute{
							// Computed: true,
							Optional: true,
							Attributes: map[string]schema.Attribute{
								"machine_type": schema.StringAttribute{
									// Computed: true,
									Required: true,
								},
								"accelerator_type": schema.StringAttribute{
									Optional: true,
									// Computed: true,
								},
								"accelerator_count": schema.Int64Attribute{
									Optional: true,
									// Computed: true,
								},
							},
						},
						"data_persistent_disk_spec": schema.SingleNestedAttribute{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"disk_type": schema.StringAttribute{
									Computed: true,
								},
								"disk_size_gb": schema.StringAttribute{
									Computed: true,
								},
							},
						},
						"network_spec": schema.SingleNestedAttribute{
							Required: true,
							Attributes: map[string]schema.Attribute{
								"enable_internet_access": schema.BoolAttribute{
									Computed: true,
								},
								"network": schema.StringAttribute{
									Required: true,
								},
								"subnetwork": schema.StringAttribute{
									Computed: true,
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
				},
			},
		},
	}
}
