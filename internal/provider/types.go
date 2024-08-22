package provider

import (
	"terraform-provider-daw/internal/gcp"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

type gcpNotebookClient struct {
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
