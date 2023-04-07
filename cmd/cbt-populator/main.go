package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"time"

	cbtclient "github.com/ihcsim/cbt-populator/pkg/generated/cbt.storage.k8s.io/clientset/versioned"
	"github.com/ihcsim/cbt-populator/pkg/populator"
	snapshotclient "github.com/kubernetes-csi/external-snapshotter/client/v6/clientset/versioned"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

func main() {
	var (
		kubeconfig   = flag.String("kubeconfig", "", "Path to a kubeconfig. For out-of-cluster development only.")
		k8sURL       = flag.String("k8s-url", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. For out-of-cluster development only e.g., with `kubectl proxy`.")
		objName      = flag.String("obj-name", "", "Name of the ChangedBlockRange object to use for data population")
		objNamespace = flag.String("obj-namespace", "", "Namespace of the ChangedBlockRange object")
		filename     = flag.String("filename", "", "Path to the file on the volume where the CBT entries will be stored")
	)
	flag.Parse()

	if err := validateFlags(kubeconfig, k8sURL, filename, objName, objNamespace); err != nil {
		klog.Error(fmt.Errorf("flags validation failed: %w", err))
		os.Exit(1)
	}

	klog.Infof("working on ChangedBlockRange '%s/%s'", *objName, *objNamespace)

	u, err := user.Current()
	if err != nil {
		klog.Error(err)
		os.Exit(1)
	}
	klog.Infof("process user info: %+v", u)

	if err := os.MkdirAll(filepath.Dir(*filename), 0777); err != nil {
		klog.Error(err)
		os.Exit(1)
	}
	klog.Infof("created data folder %s", filepath.Dir(*filename))

	file, err := os.Create(*filename)
	if err != nil {
		klog.Error(err)
		os.Exit(1)
	}
	defer file.Close()

	klog.Infof("writing CBT entries to %s", file.Name())

	cbtClient, snapshotClient, err := initCSIClientSets(*kubeconfig, *k8sURL)
	if err != nil {
		klog.Error(err)
		os.Exit(1)
	}

	var (
		p    = populator.New(cbtClient, snapshotClient)
		ctx  = context.Background()
		done = make(chan struct{})
	)

	go func() {
		defer func() {
			if err := file.Sync(); err != nil {
				klog.Error(err)
			}

			if err := p.R.Close(); err != nil {
				klog.Error(err)
			}

			done <- struct{}{}
		}()

		nbr, err := io.Copy(file, p.R)
		if err != nil {
			klog.Error(err)
			return
		}
		klog.Infof("CBT data received (bytes): %d", nbr)
	}()

	if err := p.Run(ctx, *objName, *objNamespace); err != nil {
		klog.Error(err)
		return
	}
	klog.Info("waiting for data I/O to complete")
	<-done

	time.Sleep(time.Minute * 2)
}

func validateFlags(kubeconfig, k8sURL, filename, objName, objNamespace *string) error {
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
