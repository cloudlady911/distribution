package schema2

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/docker/distribution"
)

func TestManifest(t *testing.T) {
	manifest := Manifest{
		Versioned: SchemaVersion,
		MediaType: MediaTypeManifest,
		Config: distribution.Descriptor{
			Digest:    "sha256:1a9ec845ee94c202b2d5da74a24f0ed2058318bfa9879fa541efaecba272e86b",
			Size:      985,
			MediaType: MediaTypeConfig,
		},
		Layers: []distribution.Descriptor{
			{
				Digest:    "sha256:62d8908bee94c202b2d35224a221aaa2058318bfa9879fa541efaecba272331b",
				Size:      153263,
				MediaType: MediaTypeLayer,
			},
		},
	}

	deserialized, err := FromStruct(manifest)
	if err != nil {
		t.Fatalf("error creating DeserializedManifest: %v", err)
	}

	mediaType, canonical, err := deserialized.Payload()

	if mediaType != MediaTypeManifest {
		t.Fatalf("unexpected media type: %s", mediaType)
	}

	// Check that the Canonical field is the same as json.MarshalIndent
	// with these parameters.
	p, err := json.MarshalIndent(&manifest, "", "   ")
	if err != nil {
		t.Fatalf("error marshaling manifest: %v", err)
	}
	if !bytes.Equal(p, canonical) {
		t.Fatalf("manifest bytes not equal: %q != %q", string(canonical), string(p))
	}

	var unmarshalled DeserializedManifest
	if err := json.Unmarshal(deserialized.Canonical, &unmarshalled); err != nil {
		t.Fatalf("error unmarshaling manifest: %v", err)
	}

	if !reflect.DeepEqual(&unmarshalled, deserialized) {
		t.Fatalf("manifests are different after unmarshaling: %v != %v", unmarshalled, *deserialized)
	}

	target := deserialized.Target()
	if target.Digest != "sha256:1a9ec845ee94c202b2d5da74a24f0ed2058318bfa9879fa541efaecba272e86b" {
		t.Fatalf("unexpected digest in target: %s", target.Digest.String())
	}
	if target.MediaType != MediaTypeConfig {
		t.Fatalf("unexpected media type in target: %s", target.MediaType)
	}
	if target.Size != 985 {
		t.Fatalf("unexpected size in target: %d", target.Size)
	}

	references := deserialized.References()
	if len(references) != 1 {
		t.Fatalf("unexpected number of references: %d", len(references))
	}
	if references[0].Digest != "sha256:62d8908bee94c202b2d35224a221aaa2058318bfa9879fa541efaecba272331b" {
		t.Fatalf("unexpected digest in reference: %s", references[0].Digest.String())
	}
	if references[0].MediaType != MediaTypeLayer {
		t.Fatalf("unexpected media type in reference: %s", references[0].MediaType)
	}
	if references[0].Size != 153263 {
		t.Fatalf("unexpected size in reference: %d", references[0].Size)
	}
}
