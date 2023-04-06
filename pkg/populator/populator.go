package populator

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/ihcsim/cbt-populator/pkg/apis/cbt.storage.k8s.io/v1alpha1"
	"github.com/ihcsim/cbt-populator/pkg/csi/mock"
	cbtclient "github.com/ihcsim/cbt-populator/pkg/generated/cbt.storage.k8s.io/clientset/versioned"
	cbtinformers "github.com/ihcsim/cbt-populator/pkg/generated/cbt.storage.k8s.io/informers/externalversions"
	datasource "github.com/ihcsim/cbt-populator/pkg/populator/data-source"
	snapshotclient "github.com/kubernetes-csi/external-snapshotter/client/v6/clientset/versioned"
	snapshotinformers "github.com/kubernetes-csi/external-snapshotter/client/v6/informers/externalversions"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

// CBTPopulator can populate a volume with CBT entries associated with a
// ChangedBlockRange object, identified by the ObjName and ObjNamespace
// fields.
type CBTPopulator struct {
	W                       io.Writer
	R                       io.Reader
	cbtInformerFactory      cbtinformers.SharedInformerFactory
	snapshotInformerFactory snapshotinformers.SharedInformerFactory
	dataSource              datasource.DataSource
}

// New returns a new instance of CBTPopulator.
func New(
	cbtClient cbtclient.Interface,
	snapshotClient snapshotclient.Interface,
	resync time.Duration,
	stop <-chan struct{}) *CBTPopulator {

	r, w := io.Pipe()
	cbtInformerFactory, snapshotInformerFactory := initInformerFactories(cbtClient, snapshotClient, resync, stop)
	return &CBTPopulator{
		W:                       w,
		R:                       r,
		cbtInformerFactory:      cbtInformerFactory,
		snapshotInformerFactory: snapshotInformerFactory,
		dataSource:              &mock.Mock{},
	}
}

// Run starts the CBT data population.
func (p *CBTPopulator) Run(objName, objNamespace string) error {
	obj, err := p.cbtInformerFactory.Cbt().V1alpha1().ChangedBlockRanges().Lister().ChangedBlockRanges(objNamespace).Get(objName)
	if err != nil {
		return err
	}
	klog.Infof("found ChangedBlockRange: %s/%s", obj.GetNamespace(), obj.GetName())

	from, err := p.snapshotInformerFactory.Snapshot().V1().VolumeSnapshots().Lister().VolumeSnapshots(objNamespace).Get(
		obj.Spec.FromVolumeSnapshotName)
	if err != nil {
		return err
	}
	if from.Status == nil || from.Status.ReadyToUse == nil || !(*from.Status.ReadyToUse) {
		return fmt.Errorf("VolumeSnapshot not ready: %s/%s", from.GetNamespace(), from.GetName())
	}

	to, err := p.snapshotInformerFactory.Snapshot().V1().VolumeSnapshots().Lister().VolumeSnapshots(objNamespace).Get(
		obj.Spec.ToVolumeSnapshotName)
	if err != nil {
		return err
	}
	if to.Status == nil || to.Status.ReadyToUse == nil || !(*to.Status.ReadyToUse) {
		return fmt.Errorf("VolumeSnapshot not ready: %s/%s", to.GetNamespace(), to.GetName())
	}

	fromVolumeContent, err := p.snapshotInformerFactory.Snapshot().V1().VolumeSnapshotContents().Lister().Get(
		*from.Status.BoundVolumeSnapshotContentName)
	if err != nil {
		return err
	}
	if fromVolumeContent.Status == nil {
		return fmt.Errorf("VolumeSnapshotContent not ready: %s", fromVolumeContent.GetName())
	}

	fromSnapshotHandle := fromVolumeContent.Status.SnapshotHandle
	if fromSnapshotHandle == nil {
		return fmt.Errorf("missing snapshot handle in VolumeSnapshotContent: %s", fromVolumeContent.GetName())
	}

	toVolumeContent, err := p.snapshotInformerFactory.Snapshot().V1().VolumeSnapshotContents().Lister().Get(
		*to.Status.BoundVolumeSnapshotContentName)
	if err != nil {
		return err
	}
	if toVolumeContent.Status == nil {
		return fmt.Errorf("VolumeSnapshotContent not ready: %s", toVolumeContent.GetName())
	}

	toSnapshotHandle := toVolumeContent.Status.SnapshotHandle
	if toSnapshotHandle == nil {
		return fmt.Errorf("missing snapshot handle in VolumeSnapshotContent: %s", toVolumeContent.GetName())
	}

	params := &datasource.Params{
		FromSnapshotHandle: *fromSnapshotHandle,
		ToSnapshotHandle:   *toSnapshotHandle,
	}
	return p.dataSource.FetchCBTs(params, p.W)
}

func initInformerFactories(
	cbtClient cbtclient.Interface,
	snapshotClient snapshotclient.Interface,
	resync time.Duration,
	stop <-chan struct{}) (
	cbtinformers.SharedInformerFactory,
	snapshotinformers.SharedInformerFactory,
) {

	var (
		cif = cbtinformers.NewSharedInformerFactoryWithOptions(cbtClient, resync)
		sif = snapshotinformers.NewSharedInformerFactoryWithOptions(snapshotClient, resync)
	)

	cif.Cbt().V1alpha1().ChangedBlockRanges().Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(new interface{}) {
				obj := new.(*v1alpha1.ChangedBlockRange)
				klog.Infof("adding ChangedBlockRange: %s", obj.GetName())
			},
			UpdateFunc: func(old, new interface{}) {
				obj := new.(*v1alpha1.ChangedBlockRange)
				klog.Infof("update ChangedBlockRange: %s", obj.GetName())
			},
			DeleteFunc: func(old interface{}) {
				obj := old.(*v1alpha1.ChangedBlockRange)
				klog.Infof("deleting ChangedBlockRange: %s", obj.GetName())
			},
		})

	var wg *sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		cif.Start(stop)
		cif.WaitForCacheSync(stop)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		sif.Start(stop)
		sif.WaitForCacheSync(stop)
	}()
	wg.Wait()

	klog.Info("all informers ready")
	return cif, sif
}
