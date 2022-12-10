package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +genclient
// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VolumeSnapshotDelta represents a request to find the block-level deltas
// between two volume snapshots.
type VolumeSnapshotDelta struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec of the VolumeSnapshotDelta resource.
	Spec *VolumeSnapshotDeltaSpec `json:"spec,omitempty"`

	// Status of the VolumeSnapshotDelta resource.
	Status *VolumeSnapshotDeltaStatus `json:"status,omitempty"`
}

// VolumeSnapshotDeltaSpec defines the desired state of VolumeSnapshotDelta.
type VolumeSnapshotDeltaSpec struct {
	// The name of the base volume snapshot to use for comparison.
	// If not specified, return all changed blocks.
	// +optional
	FromVolumeSnapshotName string `json:"fromVolumeSnapshotName,omitempty"`

	// The name of the target volume snapshot to use for comparison.
	// Required.
	ToVolumeSnapshotName string `json:"toVolumeSnapshotName"`

	// The number of bytes of CBT entries return should not exceed this limit.
	// +optional
	MaxSizeInBytes uint64 `json:"maxSizeInbytes"`
}

// VolumeSnapshotDeltaStatus defines the observed state of VolumeSnapshotDelta.
type VolumeSnapshotDeltaStatus struct {
	// The number of entries found.
	// Required.
	EntryCount uint64 `json:"entryCount"`

	// The block size, which is usually some constant returned by the provider.
	// Required.
	BlockSize uint64 `json:"blockSize"`

	// The number of bytes written by the volume populator to the CBT persistent
	// persistent volume.
	NumBytesWritten uint64 `json:"numBytesWritten"`

	// The last time the request status is updated.
	// LastTransitionTime metav1.Time `json:"lastTransitionTime"`

	// Human readable messages such as warnings and errors from processing the
	// request.
	// +optional
	Message string `json:"message,omitempty"`

	// The latest state of the request: "pending", "ready", or "failed".
	// Required.
	State string `json:"state,omitempty"`
}

const (
	VolumeSnapshotDeltaStatePending = "pending"
	VolumeSnapshotDeltaStateReady   = "ready"
	VolumeSnapshotDeltaStateFailed  = "failed"
)

// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VolumeSnapshotDeltaList
type VolumeSnapshotDeltaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []VolumeSnapshotDelta `json:"items"`
}
