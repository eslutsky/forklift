package base

import (
	api "github.com/konveyor/forklift-controller/pkg/apis/forklift/v1beta1"
	planapi "github.com/konveyor/forklift-controller/pkg/apis/forklift/v1beta1/plan"
	"github.com/konveyor/forklift-controller/pkg/apis/forklift/v1beta1/ref"
	plancontext "github.com/konveyor/forklift-controller/pkg/controller/plan/context"
	core "k8s.io/api/core/v1"
	cnv "kubevirt.io/client-go/api/v1"
	cdi "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"
)

// Annotations
const (
	// Used on DataVolume, contains disk source -- e.g. backing file in
	// VMware or disk ID in oVirt.
	AnnDiskSource = "forklift.konveyor.io/disk-source"
)

// Adapter API.
// Constructs provider-specific implementations
// of the Builder, Client, and Validator.
type Adapter interface {
	// Construct builder.
	Builder(ctx *plancontext.Context) (Builder, error)
	// Construct VM client.
	Client(ctx *plancontext.Context) (Client, error)
	// Construct validator.
	Validator(plan *api.Plan) (Validator, error)
	// Construct DestinationClient.
	DestinationClient(ctx *plancontext.Context) (DestinationClient, error)
}

// Builder API.
// Builds/updates objects as needed with provider
// specific constructs.
type Builder interface {
	// Build secret.
	Secret(vmRef ref.Ref, in, object *core.Secret) error
	// Build DataVolume config map.
	ConfigMap(vmRef ref.Ref, secret *core.Secret, object *core.ConfigMap) error
	// Build the Kubevirt VirtualMachine spec.
	VirtualMachine(vmRef ref.Ref, object *cnv.VirtualMachineSpec, persistentVolumeClaims []core.PersistentVolumeClaim) error
	// Build DataVolumes.
	DataVolumes(vmRef ref.Ref, secret *core.Secret, configMap *core.ConfigMap, dvTemplate *cdi.DataVolume) (dvs []cdi.DataVolume, err error)
	// Build tasks.
	Tasks(vmRef ref.Ref) ([]*planapi.Task, error)
	// Build template labels.
	TemplateLabels(vmRef ref.Ref) (labels map[string]string, err error)
	// Return a stable identifier for a DataVolume.
	ResolveDataVolumeIdentifier(dv *cdi.DataVolume) string
	// Return a stable identifier for a PersistentDataVolume
	ResolvePersistentVolumeClaimIdentifier(pvc *core.PersistentVolumeClaim) string
	// Conversion Pod environment
	PodEnvironment(vmRef ref.Ref, sourceSecret *core.Secret) (env []core.EnvVar, err error)

	// Create PersistentVolumeClaim with a DataSourceRef
	PersistentVolumeClaimWithSourceRef(da interface{}, storageName *string, populatorName string, accessModes []core.PersistentVolumeAccessMode, volumeMode *core.PersistentVolumeMode) *core.PersistentVolumeClaim
	// Add custom steps before creating PVC/DataVolume
	PreTransferActions(c Client, vmRef ref.Ref) (ready bool, err error)
}

// Client API.
// Performs provider-specific actions on the source VM.
type Client interface {
	// Power on the source VM.
	PowerOn(vmRef ref.Ref) error
	// Power off the source VM.
	PowerOff(vmRef ref.Ref) error
	// Return the source VM's power state.
	PowerState(vmRef ref.Ref) (string, error)
	// Return whether the source VM is powered off.
	PoweredOff(vmRef ref.Ref) (bool, error)
	// Create a snapshot of the source VM.
	CreateSnapshot(vmRef ref.Ref) (string, error)
	// Remove all warm migration snapshots.
	RemoveSnapshots(vmRef ref.Ref, precopies []planapi.Precopy) error
	// Check if a snapshot is ready to transfer.
	CheckSnapshotReady(vmRef ref.Ref, snapshot string) (bool, error)
	// Set DataVolume checkpoints.
	SetCheckpoints(vmRef ref.Ref, precopies []planapi.Precopy, datavolumes []cdi.DataVolume, final bool) (err error)
	// Close connections to the provider API.
	Close()
	// Finalize migrations
	Finalize(vms []*planapi.VMStatus, planName string)
}

// Validator API.
// Performs provider-specific validation.
type Validator interface {
	// Validate that a VM's disk backing storage has been mapped.
	StorageMapped(vmRef ref.Ref) (bool, error)
	// Validate that a VM's networks have been mapped.
	NetworksMapped(vmRef ref.Ref) (bool, error)
	// Validate that a VM's Host isn't in maintenance mode.
	MaintenanceMode(vmRef ref.Ref) (bool, error)
	// Validate whether warm migration is supported from this provider type.
	WarmMigration() bool
	// Validate that no more than one of a VM's networks is mapped to the pod network.
	PodNetwork(vmRef ref.Ref) (bool, error)
}

// DestinationClient API.
// Performs provider-specific actions on the Destination cluster
type DestinationClient interface {
	// Deletes Populator Data Source
	DeletePopulatorDataSource(vm *planapi.VMStatus) error
	// Set the VolumePopulator CustomResource Ownership.
	SetPopulatorCrOwnership() error
}
