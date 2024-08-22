package gcp

// https://cloud.google.com/vertex-ai/docs/reference/rest/v1beta1/projects.locations.notebookRuntimeTemplates

type MachineSpec struct {
	MachineType      *string `json:"machineType,omitempty" yaml:"machineType,omitempty"`
	AcceleratorType  *string `json:"acceleratorType,omitempty" yaml:"acceleratorType,omitempty"`
	AcceleratorCount *int64  `json:"acceleratorCount,omitempty" yaml:"acceleratorCount,omitempty"`
}

type DataPersistentDiskSpec struct {
	DiskType   *string `json:"diskType,omitempty" yaml:"diskType,omitempty"`
	DiskSizeGb *string `json:"diskSizeGb,omitempty" yaml:"diskSizeGb,omitempty"`
}

type NetworkSpec struct {
	EnableInternetAccess *bool   `json:"enableInternetAccess,omitempty" yaml:"enableInternetAccess,omitempty"`
	Network              *string `json:"network,omitempty" yaml:"network,omitempty"`
	Subnetwork           *string `json:"subnetwork,omitempty" yaml:"subnetwork,omitempty"`
}

type IdleShutdownConfig struct {
	IdleTimeout          *string `json:"idleTimeout,omitempty" yaml:"idleTimeout,omitempty"`
	IdleShutdownDisabled *bool   `json:"idleShutdownDisabled,omitempty" yaml:"idleShutdownDisabled,omitempty"`
}

type EucConfig struct {
	EucDisabled      *bool `json:"eucDisabled,omitempty" yaml:"eucDisabled,omitempty"`
	BypassActasCheck *bool `json:"bypassActasCheck,omitempty" yaml:"bypassActasCheck,omitempty"`
}

type ShieldedVmConfig struct {
	EnableSecureBoot *bool `json:"enableSecureBoot,omitempty" yaml:"enableSecureBoot,omitempty"`
}

type EncryptionSpec struct {
	KmsKeyName *string `json:"kmsKeyName,omitempty" yaml:"kmsKeyName,omitempty"`
}

type NotebookRuntimeTemplate struct {
	Name                   *string                 `json:"name,omitempty" yaml:"name,omitempty"`
	DisplayName            *string                 `json:"displayName" yaml:"displayName"`
	Description            *string                 `json:"description,omitempty" yaml:"description,omitempty"`
	IsDefault              *bool                   `json:"isDefault,omitempty" yaml:"isDefault,omitempty"`
	MachineSpec            *MachineSpec            `json:"machineSpec,omitempty" yaml:"machineSpec,omitempty"`
	DataPersistentDiskSpec *DataPersistentDiskSpec `json:"dataPersistentDiskSpec,omitempty" yaml:"dataPersistentDiskSpec,omitempty"`
	NetworkSpec            *NetworkSpec            `json:"networkSpec,omitempty" yaml:"networkSpec,omitempty"`
	ServiceAccount         *string                 `json:"serviceAccount,omitempty" yaml:"serviceAccount,omitempty"`
	Etag                   *string                 `json:"etag,omitempty" yaml:"etag,omitempty"`
	Labels                 *map[string]string      `json:"labels,omitempty" yaml:"labels,omitempty"`
	IdleShutdownConfig     *IdleShutdownConfig     `json:"idleShutdownConfig,omitempty" yaml:"idleShutdownConfig,omitempty"`
	EucConfig              *EucConfig              `json:"eucConfig,omitempty" yaml:"eucConfig,omitempty"`
	CreateTime             *string                 `json:"createTime,omitempty" yaml:"createTime,omitempty"`
	UpdateTime             *string                 `json:"updateTime,omitempty" yaml:"updateTime,omitempty"`
	NotebookRuntimeType    *string                 `json:"notebookRuntimeType,omitempty" yaml:"notebookRuntimeType,omitempty"`
	ShieldedVmConfig       *ShieldedVmConfig       `json:"shieldedVmConfig,omitempty" yaml:"shieldedVmConfig,omitempty"`
	EncryptionSpec         *EncryptionSpec         `json:"encryptionSpec,omitempty"`
}

// this get's returned when we perform a GET
type ListNotebookRuntimeTemplatesResult struct {
	NotebookRuntimeTemplates []NotebookRuntimeTemplate `json:"notebookRuntimeTemplates"`
}
