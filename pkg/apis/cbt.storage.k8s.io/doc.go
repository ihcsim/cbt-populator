// +k8s:deepcopy-gen=package

// Package cbt defines the high-level scheme information of all the CBT
// resources.
package cbt

const (
	//GroupName is the name of this group.
	GroupName = "cbt.storage.k8s.io"

	// ChangedBlockRangeKind defines the kind of the ChangedBlockRange resource.
	ChangedBlockRangeKind = "ChangedBlockRange"

	// ChangedBlockRangeResource defines the resource of the ChangedBlockRange resource.
	ChangedBlockRangeResource = "changedblockranges"
)
