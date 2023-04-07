package mock

import (
	"io"

	"github.com/ihcsim/cbt-populator/pkg/cbt"
	"k8s.io/klog/v2"
)

type Mock struct{}

func (m *Mock) FetchCBTs(params *cbt.Params, w io.WriteCloser) error {
	defer w.Close()

	// source: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/readsnapshots.html#list-different-blocks
	data := []byte(`{
  "ChangedBlocks": [
    {
      "BlockIndex": 0,
      "FirstBlockToken": "AAABAVahm9SO60Dyi0ORySzn2ZjGjW/KN3uygGlS0QOYWesbzBbDnX2dGpmC",
      "SecondBlockToken": "AAABAf8o0o6UFi1rDbSZGIRaCEdDyBu9TlvtCQxxoKV8qrUPQP7vcM6iWGSr"
    },
    {
      "BlockIndex": 6000,
      "FirstBlockToken": "AAABAbYSiZvJ0/R9tz8suI8dSzecLjN4kkazK8inFXVintPkdaVFLfCMQsKe",
      "SecondBlockToken": "AAABAZnqTdzFmKRpsaMAsDxviVqEI/3jJzI2crq2eFDCgHmyNf777elD9oVR"
    },
    {
      "BlockIndex": 6001,
      "FirstBlockToken": "AAABASBpSJ2UAD3PLxJnCt6zun4/T4sU25Bnb8jB5Q6FRXHFqAIAqE04hJoR"
    },
    {
      "BlockIndex": 6002,
      "FirstBlockToken": "AAABASqX4/NWjvNceoyMUljcRd0DnwbSwNnes1UkoP62CrQXvn47BY5435aw"
    },
    {
      "BlockIndex": 6003,
      "FirstBlockToken": "AAABASmJ0O5JxAOce25rF4P1sdRtyIDsX12tFEDunnePYUKOf4PBROuICb2A"
    },
  ],
  "ExpiryTime": 1576308931.973,
  "VolumeSize": 32212254720,
  "BlockSize": 524288,
  "NextToken": "AAADARqElNng/sV98CYk/bJDCXeLJmLJHnNSkHvLzVaO0zsPH/QM3Bi3zF//O6Mdi/BbJarBnp8h"
  }`)

	nbr, err := w.Write(data)
	klog.Infof("CBT data written (bytes): %d", nbr)
	return err
}
