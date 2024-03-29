package main

import (
	"flag"
	"fmt"
	"path/filepath"
	"time"

	cbt "github.com/ihcsim/cbt-populator/pkg/apis/cbt.storage.k8s.io"
	"github.com/ihcsim/cbt-populator/pkg/apis/cbt.storage.k8s.io/v1alpha1"

	populatormachinery "github.com/kubernetes-csi/lib-volume-populator/populator-machinery"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
)

var (
	kubeconfig, k8sURL                                                          string
	listenAddr, metricsPath                                                     string
	populatorImage, populatorNamespace, populatorDevicePath, populatorMountPath string

	prefix = "cbt.storage.k8s.io"
)

func main() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. For out-of-cluster development only.")
	flag.StringVar(&k8sURL, "k8s-url", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. For out-of-cluster development only e.g., with `kubectl proxy`.")
	flag.StringVar(&listenAddr, "listen-addr", ":8080", "The TCP network address for metrics and leader election health check. Empty string means the server is disabled.")
	flag.StringVar(&metricsPath, "metrics-path", "/metrics", "The HTTP path where prometheus metrics will be exposed. Default is `/metrics`.")
	flag.StringVar(&populatorImage, "populator-image", "debian:11-slim", "Name and tag of the populator container image")
	flag.StringVar(&populatorNamespace, "populator-namespace", "cbt-populator", "Namespace of the populator pod")
	flag.StringVar(&populatorDevicePath, "populator-device-path", "/dev/sdh", "Device path to use in the populator pod (for block PVC)")
	flag.StringVar(&populatorMountPath, "populator-mount-path", "/data", "Mount path to use in the populator pod (for file PVC)")

	flag.Parse()

	klog.Infof("listen address=%s, metrics path=%s, namespace=%s", listenAddr, metricsPath, populatorNamespace)

	populatormachinery.RunController(
		k8sURL,
		kubeconfig,
		populatorImage,
		listenAddr,
		metricsPath,
		populatorNamespace,
		prefix,
		v1alpha1.Kind(cbt.ChangedBlockRangeKind),
		v1alpha1.VersionResource(cbt.ChangedBlockRangeResource),
		populatorMountPath,
		populatorDevicePath,
		populatorArgs)
}

func populatorArgs(rawBlock bool, u *unstructured.Unstructured) ([]string, error) {
	var (
		obj  v1alpha1.ChangedBlockRange
		args []string
	)

	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.UnstructuredContent(), &obj); err != nil {
		return nil, err
	}

	args = append(args, "--obj-name", obj.GetName())
	args = append(args, "--obj-namespace", obj.GetNamespace())

	if rawBlock {
		return append(args, "--filename", populatorDevicePath), nil
	}

	filename := filepath.Join(populatorMountPath, fmt.Sprintf("cbt-%d", time.Now().Unix()))
	return append(args, "--filename", filename), nil
}
