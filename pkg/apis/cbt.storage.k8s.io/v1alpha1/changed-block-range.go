package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +genclient
// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ChangedBlockRange represents the range of changed blocks between two block
// volume snapshots.
type ChangedBlockRange struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec of the ChangedBlockRange resource.
	Spec *ChangedBlockRangeSpec `json:"spec,omitempty"`

	// Status of the ChangedBlockRange resource.
	Status *ChangedBlockRangeStatus `json:"status,omitempty"`
}

// ChangedBlockRangeSpec defines the desired state of ChangedBlockRange.
type ChangedBlockRangeSpec struct {
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

// ChangedBlockRangeStatus defines the observed state of ChangedBlockRange.
type ChangedBlockRangeStatus struct {
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
	// ChangedBlockRangeStatePending defines the 'pending' state.
	ChangedBlockRangeStatePending = "pending"

	// ChangedBlockRangeStateReady defines the 'ready' state.
	ChangedBlockRangeStateReady = "ready"

	// ChangedBlockRangeStateFailed defines the 'failed' state.
	ChangedBlockRangeStateFailed = "failed"
)

// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ChangedBlockRangeList represents a list of ChangedBlockRange resources.
type ChangedBlockRangeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []ChangedBlockRange `json:"items"`
}
