package cbt

import "io"

// Provider exposes the FetchCBTs API which serves an abstract interface betewedwe
// between CBT consumer and the CBT providers.
type Provider interface {
	// FetchCBTs sends a CBT request, containing input parameters defined by
	// params to the CBT provider. The CBT response from the provider is written
	// to the writer w.
	FetchCBTs(params *Params, w io.WriteCloser) error
}

// Params is a collection of parameters used to construct to CBT request to be
// sent to the provider.
type Params struct {
	// FromSnaphotHandle is the 'from' snapshot handle to be used for the CBT
	// computation.
	FromSnapshotHandle string

	// ToSnaphotHandle is the 'to' snapshot handle to be used for the CBT
	// computation.
	ToSnapshotHandle string

	// MaxBytes defines the maximum amount of data to be returned to the CBT
	// consumer.
	MaxBytes uint64
}
