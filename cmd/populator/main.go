package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/ihcsim/cbt-populator/pkg/apis/cbt/v1alpha1"
	cbtclient "github.com/ihcsim/cbt-populator/pkg/generated/cbt/clientset/versioned"
	"github.com/ihcsim/cbt-populator/pkg/populator"
	snapshotclient "github.com/kubernetes-csi/external-snapshotter/client/v6/clientset/versioned"

	populatormachinery "github.com/kubernetes-csi/lib-volume-populator/populator-machinery"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

const (
	modeController = "controller"
	modePopulator  = "populator"
)

var (
	kubeconfig, k8sURL                                      string
	listenAddr, metricsPath                                 string
	populatorImage, populatorDevicePath, populatorMountPath string

	controllerNamespace = "default"
	prefix              = "cbt.csi.k8s.io"
)

func main() {
	stop := make(chan struct{})
	mode := modeController
	switch mode {
	case modeController:
		klog.Info("starting controller...")
		runAsController()
	case modePopulator:
		klog.Info("starting populator...")
		runAsPopulator(stop)
	default:
		klog.Fatalf("process terminated due to unsupported mode: %s", mode)
	}
}

func runAsController() {
	flag.StringVar(&kubeconfig, "kubeconfig", "~/.kube/config", "Path to a kubeconfig. For out-of-cluster development only.")
	flag.StringVar(&k8sURL, "k8s-url", "https://0.0.0.0:6443", "The address of the Kubernetes API server. Overrides any value in kubeconfig. For out-of-cluster development only e.g., with `kubectl proxy`.")
	flag.StringVar(&listenAddr, "listen-addr", ":8080", "The TCP network address for metrics and leader election health check. Empty string means the server is disabled.")
	flag.StringVar(&metricsPath, "metrics-path", "/metrics", "The HTTP path where prometheus metrics will be exposed. Default is `/metrics`.")
	flag.StringVar(&populatorImage, "populator-image", "debian:11-slim", "Name and tag of the populator container image")
	flag.StringVar(&populatorDevicePath, "populator-device-path", "/dev/sdh", "Device path to use in the populator pod (for block PVC)")
	flag.StringVar(&populatorMountPath, "populator-mount-path", "/mnt/cbt", "Mount path to use in the populator pod (for file PVC)")
	flag.Parse()

	if ns := os.Getenv("CONTROLLER_NAMESPACE"); ns != "" {
		controllerNamespace = ns
	}
	klog.Infof("listen address=%s, metrics path=%s, namespace=%s", listenAddr, metricsPath, controllerNamespace)

	populatormachinery.RunController(
		k8sURL,
		kubeconfig,
		populatorImage,
		listenAddr,
		metricsPath,
		controllerNamespace,
		prefix,
		v1alpha1.Kind(v1alpha1.KindVolumeSnapshotDelta),
		v1alpha1.VersionResource(v1alpha1.ResourceVolumeSnapshotDelta),
		populatorMountPath,
		populatorDevicePath,
		populatorArgs)
}

func runAsPopulator(stop <-chan struct{}) error {
	var (
		objName      = flag.String("obj-name", "", "Name of the VolumeSnapshotDelta object to use for data population")
		objNamespace = flag.String("obj-namespace", "", "Namespace of the VolumeSnapshotDelta object")
		filename     = flag.String("filename", "", "Path to the file on the volume where the CBT entries will be stored")
	)
	flag.Parse()

	if filename == nil || *filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}
	if objName == nil || *objName == "" {
		return fmt.Errorf("object name cannot be empty")
	}
	if objNamespace == nil || *objNamespace == "" {
		return fmt.Errorf("object namespace cannot be empty")
	}
	klog.Infof("obj name=%s, obj namespace=%s", objName, objNamespace)

	resync := time.Minute * 20
	cbtClient, snapshotClient, err := initCSIClientSets()
	populator := populator.New(cbtClient, snapshotClient, resync, stop)
	go func() {
		if err := populator.Run(*objName, *objNamespace); err != nil {
			klog.Error(err)
			return
		}
	}()

	file, err := os.Create(*filename)
	if err != nil {
		return err
	}
	klog.Infof("writing CBT entries to %s", file.Name)

	nbr, err := io.Copy(file, populator.R)
	if err != nil {
		return err
	}

	klog.Info("number of bytes read: %d", nbr)
	return nil
}

func populatorArgs(rawBlock bool, u *unstructured.Unstructured) ([]string, error) {
	var (
		obj  v1alpha1.VolumeSnapshotDelta
		args []string
	)

	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.UnstructuredContent(), &obj); err != nil {
		return nil, err
	}

	args = append(args, "--obj-name=", obj.GetName())
	args = append(args, "--obj-namespace=", obj.GetNamespace())

	if rawBlock {
		args = append(args, "--filename="+populatorDevicePath)
	} else {
		filename := filepath.Join(populatorMountPath, fmt.Sprintf("cbt-%d", time.Now().Unix()))
		args = append(args, "--filename="+filename)
	}

	return args, nil
}

func initCSIClientSets() (cbtclient.Interface, snapshotclient.Interface, error) {
	restConfig, err := clientcmd.BuildConfigFromFlags(k8sURL, kubeconfig)
	if err != nil {
		return nil, nil, err
	}

	cbtClient, err := cbtclient.NewForConfig(restConfig)
	if err != nil {
		return nil, nil, err
	}

	snapshotClient, err := snapshotclient.NewForConfig(restConfig)
	if err != nil {
		return nil, nil, err
	}

	return cbtClient, snapshotClient, nil
}
