package v1alpha1

import (
	"github.com/ihcsim/cbt-populator/pkg/apis/cbt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{
		Group:   cbt.GroupName,
		Version: "v1alpha1",
	}

	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)

	AddToScheme = SchemeBuilder.AddToScheme
)

// Adds the list of known types to the given scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&VolumeSnapshotDelta{},
		&VolumeSnapshotDeltaList{},
	)

	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}

// Resource takes an unqualified resource and returns a Group qualified GroupResource
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

// func Kind(kind string) schema.GroupKind {
// return schema.GroupKind{
// Group: "cbt",
// Kind:  kind,
// }
// }

// func VersionResource(resource string) schema.GroupVersionResource {
// return SchemeGroupVersion.WithResource(resource)
// }
