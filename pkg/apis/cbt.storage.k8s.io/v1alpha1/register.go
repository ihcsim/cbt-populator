package v1alpha1

import (
	cbt "github.com/ihcsim/cbt-populator/pkg/apis/cbt.storage.k8s.io"
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

	// SchemeBuilder handles all scheme-related functions.
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)

	// AddToScheme is used to add new scheme to SchemeBuilder.
	AddToScheme = SchemeBuilder.AddToScheme
)

// Adds the list of known types to the given scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&ChangedBlockRange{},
		&ChangedBlockRangeList{},
	)

	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}

// Resource takes an unqualified resource and returns a Group qualified GroupResource.
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

// Kind converts the given 'kind' into the CBT GroupKind object.
func Kind(kind string) schema.GroupKind {
	return schema.GroupKind{
		Group: cbt.GroupName,
		Kind:  kind,
	}
}

// VersionResource converts the given 'resource' into the CBT GroupVersionResource.
func VersionResource(resource string) schema.GroupVersionResource {
	return SchemeGroupVersion.WithResource(resource)
}
