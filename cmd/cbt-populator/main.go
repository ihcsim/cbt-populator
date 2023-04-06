package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	cbtclient "github.com/ihcsim/cbt-populator/pkg/generated/cbt.storage.k8s.io/clientset/versioned"
	"github.com/ihcsim/cbt-populator/pkg/populator"
	snapshotclient "github.com/kubernetes-csi/external-snapshotter/client/v6/clientset/versioned"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

func main() {
	klog.Info("starting populator...")

	var (
		stop         = make(chan struct{})
		kubeconfig   = flag.String("kubeconfig", "~/.kube/config", "Path to a kubeconfig. For out-of-cluster development only.")
		k8sURL       = flag.String("k8s-url", "https://0.0.0.0:6443", "The address of the Kubernetes API server. Overrides any value in kubeconfig. For out-of-cluster development only e.g., with `kubectl proxy`.")
		objName      = flag.String("obj-name", "", "Name of the ChangedBlockRange object to use for data population")
		objNamespace = flag.String("obj-namespace", "", "Namespace of the ChangedBlockRange object")
		filename     = flag.String("filename", "", "Path to the file on the volume where the CBT entries will be stored")
	)
	flag.Parse()

	if err := validateFlags(kubeconfig, k8sURL, filename, objName, objNamespace); err != nil {
		klog.Error(fmt.Errorf("flags validation failed: %w", err))
		os.Exit(1)
	}

	klog.Infof("obj name=%s, obj namespace=%s", objName, objNamespace)

	resync := time.Minute * 20
	cbtClient, snapshotClient, err := initCSIClientSets(*kubeconfig, *k8sURL)
	populator := populator.New(cbtClient, snapshotClient, resync, stop)
	go func() {
		if err := populator.Run(*objName, *objNamespace); err != nil {
			klog.Error(err)
			return
		}
	}()

	file, err := os.Create(*filename)
	if err != nil {
		klog.Error(err)
		os.Exit(1)
	}
	klog.Infof("writing CBT entries to %s", file.Name)

	nbr, err := io.Copy(file, populator.R)
	if err != nil {
		klog.Error(err)
		os.Exit(1)
	}

	klog.Info("number of bytes read: %d", nbr)
}

func validateFlags(kubeconfig, k8sURL, filename, objName, objNamespace *string) error {
	if kubeconfig == nil || *kubeconfig == "" {
		return fmt.Errorf("kubeconfig cannot be empty")
	}

	if k8sURL == nil || *k8sURL == "" {
		return fmt.Errorf("k8sURL cannot be empty")
	}

	if filename == nil || *filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}

	if objName == nil || *objName == "" {
		return fmt.Errorf("object name cannot be empty")
	}

	if objNamespace == nil || *objNamespace == "" {
		return fmt.Errorf("object namespace cannot be empty")
	}

	return nil
}

func initCSIClientSets(kubeconfig, k8sURL string) (cbtclient.Interface, snapshotclient.Interface, error) {
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
