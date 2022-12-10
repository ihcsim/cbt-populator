package mock

import (
	"io"

	datasource "github.com/ihcsim/cbt-populator/pkg/populator/data-source"
)

type Mock struct{}

func (m *Mock) FetchCBTs(params *datasource.Params, w io.Writer) error {
	return nil
}
