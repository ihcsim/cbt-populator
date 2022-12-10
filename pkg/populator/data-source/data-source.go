package datasource

import "io"

type DataSource interface {
	FetchCBTs(params *Params, w io.Writer) error
}

type Params struct {
	FromSnapshotHandle string
	ToSnapshotHandle   string
	MaxBytes           uint64
}
