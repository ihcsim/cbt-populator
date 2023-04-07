package populator

import (
	"context"
	"fmt"
	"io"

	"github.com/ihcsim/cbt-populator/pkg/apis/cbt.storage.k8s.io/v1alpha1"
	"github.com/ihcsim/cbt-populator/pkg/cbt"
	"github.com/ihcsim/cbt-populator/pkg/cbt/provider/mock"
	cbtclient "github.com/ihcsim/cbt-populator/pkg/generated/cbt.storage.k8s.io/clientset/versioned"
	snapshotclient "github.com/kubernetes-csi/external-snapshotter/client/v6/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

// CBTPopulator can populate a volume with CBT entries associated with a
// ChangedBlockRange object, identified by the ObjName and ObjNamespace
// fields.
type CBTPopulator struct {
	W              io.WriteCloser
	R              io.ReadCloser
	cbtClient      cbtclient.Interface
	snapshotClient snapshotclient.Interface
	provider       cbt.Provider
}

// New returns a new instance of CBTPopulator.
func New(
	cbtClient cbtclient.Interface,
	snapshotClient snapshotclient.Interface) *CBTPopulator {

	r, w := io.Pipe()
	return &CBTPopulator{
		W:              w,
		R:              r,
		cbtClient:      cbtClient,
		snapshotClient: snapshotClient,
		provider:       &mock.Mock{},
	}
}

// Run starts the CBT data population.
func (p *CBTPopulator) Run(ctx context.Context, objName, objNamespace string) error {
	opt := metav1.GetOptions{}
	obj, err := p.cbtClient.CbtV1alpha1().ChangedBlockRanges(objNamespace).Get(ctx, objName, opt)
	if err != nil {
		return err
	}
	klog.Infof("found changed block range: %s/%s", obj.GetNamespace(), obj.GetName())

	snapshotHandles, err := p.findSnapshotHandles(ctx, obj)
	if err != nil {
		return err
	}

	params := &cbt.Params{
		FromSnapshotHandle: snapshotHandles[0],
		ToSnapshotHandle:   snapshotHandles[1],
	}
	return p.provider.FetchCBTs(params, p.W)
}

func (p *CBTPopulator) findSnapshotHandles(ctx context.Context, obj *v1alpha1.ChangedBlockRange) ([]string, error) {
	var (
		namespace              = obj.GetNamespace()
		fromVolumeSnapshotName = obj.Spec.FromVolumeSnapshotName
		toVolumeSnapshotName   = obj.Spec.ToVolumeSnapshotName
		opt                    = metav1.GetOptions{}
		snapshotHandles        = []string{}
	)

	for _, snapshotName := range []string{fromVolumeSnapshotName, toVolumeSnapshotName} {
		klog.Infof("trying to find volume snapshot %s/%s", namespace, snapshotName)
		volumeSnapshot, err := p.snapshotClient.SnapshotV1().VolumeSnapshots(namespace).Get(ctx, snapshotName, opt)
		if err != nil {
			return nil, err
		}

		if volumeSnapshot.Status == nil || volumeSnapshot.Status.ReadyToUse == nil || !(*volumeSnapshot.Status.ReadyToUse) {
			return nil, fmt.Errorf("volume snapshot not ready: %s/%s", namespace, volumeSnapshot.GetName())
		}

		volumeContentName := *volumeSnapshot.Status.BoundVolumeSnapshotContentName
		volumeContent, err := p.snapshotClient.SnapshotV1().VolumeSnapshotContents().Get(ctx, volumeContentName, opt)
		if err != nil {
			return nil, err
		}

		if volumeContent.Status == nil {
			return nil, fmt.Errorf("volume snapshot content not ready: %s", volumeContent.GetName())
		}

		snapshotHandle := volumeContent.Status.SnapshotHandle
		if snapshotHandle == nil {
			return nil, fmt.Errorf("missing snapshot handle in volume snapshot content: %s", volumeContent.GetName())
		}
		klog.Infof("found volume snapshot %s/%s with snapshot handle %s", namespace, volumeSnapshot.GetName(), *snapshotHandle)

		snapshotHandles = append(snapshotHandles, *snapshotHandle)
	}

	return snapshotHandles, nil
}
