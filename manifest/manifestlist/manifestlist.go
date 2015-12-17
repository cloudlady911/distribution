package manifestlist

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/docker/distribution"
	"github.com/docker/distribution/digest"
	"github.com/docker/distribution/manifest"
)

// MediaTypeManifestList specifies the mediaType for manifest lists.
const MediaTypeManifestList = "application/vnd.docker.distribution.manifest.list.v2+json"

// SchemaVersion provides a pre-initialized version structure for this
// packages version of the manifest.
var SchemaVersion = manifest.Versioned{
	SchemaVersion: 2,
}

func init() {
	manifestListFunc := func(b []byte) (distribution.Manifest, distribution.Descriptor, error) {
		m := new(DeserializedManifestList)
		err := m.UnmarshalJSON(b)
		if err != nil {
			return nil, distribution.Descriptor{}, err
		}

		dgst := digest.FromBytes(b)
		return m, distribution.Descriptor{Digest: dgst, Size: int64(len(b)), MediaType: MediaTypeManifestList}, err
	}
	err := distribution.RegisterManifestSchema(MediaTypeManifestList, manifestListFunc)
	if err != nil {
		panic(fmt.Sprintf("Unable to register manifest: %s", err))
	}
}

// PlatformSpec specifies a platform where a particular image manifest is
// applicable.
type PlatformSpec struct {
	Architecture string   `json:"architecture"`
	OS           string   `json:"os"`
	Variant      string   `json:"variant,omitempty"`
	Features     []string `json:"features,omitempty"`
}

// A ManifestDescriptor references a platform-specific manifest.
type ManifestDescriptor struct {
	distribution.Descriptor
	PlatformSpec
}

// ManifestList references manifests for various platforms.
type ManifestList struct {
	manifest.Versioned

	// MediaType is the media type of this document. It should always
	// be set to MediaTypeManifestList.
	MediaType string `json:"mediaType"`

	// Config references the image configuration as a blob.
	Manifests []ManifestDescriptor `json:"manifests"`
}

// References returnes the distribution descriptors for the referenced image
// manifests.
func (m ManifestList) References() []distribution.Descriptor {
	dependencies := make([]distribution.Descriptor, len(m.Manifests))
	for i := range m.Manifests {
		dependencies[i] = m.Manifests[i].Descriptor
	}

	return dependencies
}

// DeserializedManifestList wraps ManifestList with a copy of the original
// JSON.
type DeserializedManifestList struct {
	ManifestList

	// Canonical is the canonical byte representation of the Manifest.
	Canonical []byte `json:"-"`
}

// FromDescriptors takes a slice of descriptors, and returns a
// DeserializedManifestList which contains the resulting manifest list
// and its JSON representation.
func FromDescriptors(descriptors []ManifestDescriptor) (*DeserializedManifestList, error) {
	m := ManifestList{
		Versioned: SchemaVersion,
		MediaType: MediaTypeManifestList,
	}

	m.Manifests = make([]ManifestDescriptor, len(descriptors), len(descriptors))
	copy(m.Manifests, descriptors)

	deserialized := DeserializedManifestList{
		ManifestList: m,
	}

	var err error
	deserialized.Canonical, err = json.MarshalIndent(&m, "", "   ")
	return &deserialized, err
}

// UnmarshalJSON populates a new ManifestList struct from JSON data.
func (m *DeserializedManifestList) UnmarshalJSON(b []byte) error {
	m.Canonical = make([]byte, len(b), len(b))
	// store manifest list in Canonical
	copy(m.Canonical, b)

	// Unmarshal canonical JSON into ManifestList object
	var manifestList ManifestList
	if err := json.Unmarshal(m.Canonical, &manifestList); err != nil {
		return err
	}

	m.ManifestList = manifestList

	return nil
}

// MarshalJSON returns the contents of Canonical. If Canonical is empty,
// marshals the inner contents.
func (m *DeserializedManifestList) MarshalJSON() ([]byte, error) {
	if len(m.Canonical) > 0 {
		return m.Canonical, nil
	}

	return nil, errors.New("JSON representation not initialized in DeserializedManifestList")
}

// Payload returns the raw content of the manifest list. The contents can be
// used to calculate the content identifier.
func (m DeserializedManifestList) Payload() (string, []byte, error) {
	return m.MediaType, m.Canonical, nil
}
