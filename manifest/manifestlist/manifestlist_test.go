package manifestlist

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/docker/distribution"
)

func TestManifestList(t *testing.T) {
	manifestDescriptors := []ManifestDescriptor{
		{
			Descriptor: distribution.Descriptor{
				Digest:    "sha256:1a9ec845ee94c202b2d5da74a24f0ed2058318bfa9879fa541efaecba272e86b",
				Size:      985,
				MediaType: "application/vnd.docker.distribution.manifest.v2+json",
			},
			PlatformSpec: PlatformSpec{
				Architecture: "amd64",
				OS:           "linux",
				Features:     []string{"sse4"},
			},
		},
		{
			Descriptor: distribution.Descriptor{
				Digest:    "sha256:6346340964309634683409684360934680934608934608934608934068934608",
				Size:      2392,
				MediaType: "application/vnd.docker.distribution.manifest.v2+json",
			},
			PlatformSpec: PlatformSpec{
				Architecture: "sun4m",
				OS:           "sunos",
			},
		},
	}

	deserialized, err := FromDescriptors(manifestDescriptors)
	if err != nil {
		t.Fatalf("error creating DeserializedManifestList: %v", err)
	}

	mediaType, canonical, err := deserialized.Payload()

	if mediaType != MediaTypeManifestList {
		t.Fatalf("unexpected media type: %s", mediaType)
	}

	// Check that the Canonical field is the same as json.MarshalIndent
	// with these parameters.
	p, err := json.MarshalIndent(&deserialized.ManifestList, "", "   ")
	if err != nil {
		t.Fatalf("error marshaling manifest list: %v", err)
	}
	if !bytes.Equal(p, canonical) {
		t.Fatalf("manifest bytes not equal: %q != %q", string(canonical), string(p))
	}

	var unmarshalled DeserializedManifestList
	if err := json.Unmarshal(deserialized.Canonical, &unmarshalled); err != nil {
		t.Fatalf("error unmarshaling manifest: %v", err)
	}

	if !reflect.DeepEqual(&unmarshalled, deserialized) {
		t.Fatalf("manifests are different after unmarshaling: %v != %v", unmarshalled, *deserialized)
	}

	references := deserialized.References()
	if len(references) != 2 {
		t.Fatalf("unexpected number of references: %d", len(references))
	}
	for i := range references {
		if !reflect.DeepEqual(references[i], manifestDescriptors[i].Descriptor) {
			t.Fatalf("unexpected value %d returned by References: %v", i, references[i])
		}
	}
}
