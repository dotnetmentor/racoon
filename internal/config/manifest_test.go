package config_test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dotnetmentor/racoon/internal/config"
	"gopkg.in/yaml.v2"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Manifest", func() {
	Context("NewManifest", func() {
		testdataDir := "./../../testdata"

		When("creating a new manifest from an existing file", func() {
			It("produces an error when file path is not found", func() {
				_, err := config.NewManifest([]string{filepath.Join(testdataDir, "notfound.yaml")})

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to find manifest file paths=[../../testdata/notfound.yaml]"))
			})
		})

		When("creating a new manifest from an existing file", func() {
			var m config.Manifest
			var err error

			BeforeEach(func() {
				m, err = config.NewManifest([]string{filepath.Join(testdataDir, "racoon.yaml")})
			})

			It("parses file to manifest without error", func() {
				Expect(err).To(Not(HaveOccurred()))
				Expect(m.Name).To(Equal("racoon-e2e-tests"))
			})

			It("merges layers from base layer", func() {
				Expect(m.Layers).To(HaveLen(5))
			})
		})

		When("parsing single file manifest", func() {
			It("parses layers", func() {
				tmpfile, err := os.CreateTemp("", "racoon-*.yaml")
				defer os.Remove(tmpfile.Name())
				if err != nil {
					Fail(fmt.Sprintf("failed to create temp file for test, %v", err))
				}

				expected := config.Manifest{
					Layers: []config.LayerConfig{
						{
							Name: "layer1",
						},
					},
				}

				b, _ := yaml.Marshal(expected)
				if _, err = tmpfile.Write(b); err != nil {
					Fail(fmt.Sprintf("failed to write to temp file %s", tmpfile.Name()))
				}

				m, err := config.NewManifest([]string{tmpfile.Name()})

				Expect(err).To(Not(HaveOccurred()))
				Expect(m.Layers).To(HaveLen(1))
				Expect(m.Layers[0].Name).To(Equal("layer1"))
			})
		})

		When("parsing manifest with base file", func() {
			It("can parse layers from base only", func() {
				base := config.Manifest{
					ExtendsConfig: config.ExtendsConfig{},
					Layers: []config.LayerConfig{
						{
							Name: "layer1",
						},
					},
				}
				bf, err := NewTempManifestFile(base, "base-*.yaml")
				defer os.Remove(bf.Name())
				if err != nil {
					Fail(fmt.Sprintf("failed to create temp file for test, %v", err))
				}

				manifest := config.Manifest{
					ExtendsConfig: config.ExtendsConfig{
						Extends: bf.Name(),
					},
				}
				mf, err := NewTempManifestFile(manifest, "racoon-*.yaml")
				defer os.Remove(mf.Name())
				if err != nil {
					Fail(fmt.Sprintf("failed to create temp file for test, %v", err))
				}

				m, err := config.NewManifest([]string{mf.Name()})

				Expect(err).To(Not(HaveOccurred()))
				Expect(m.Layers).To(HaveLen(1))
				Expect(m.Layers[0].Name).To(Equal("layer1"))
			})

			It("can parse layers from both base and extending manifest", func() {
				base := config.Manifest{
					ExtendsConfig: config.ExtendsConfig{},
					Layers: []config.LayerConfig{
						{
							Name: "layer1",
						},
					},
				}
				bf, err := NewTempManifestFile(base, "base-*.yaml")
				defer os.Remove(bf.Name())
				if err != nil {
					Fail(fmt.Sprintf("failed to create temp file for test, %v", err))
				}

				manifest := config.Manifest{
					ExtendsConfig: config.ExtendsConfig{
						Extends: bf.Name(),
					},
					Layers: []config.LayerConfig{
						{
							Name: "layer2",
						},
					},
				}
				mf, err := NewTempManifestFile(manifest, "racoon-*.yaml")
				defer os.Remove(mf.Name())
				if err != nil {
					Fail(fmt.Sprintf("failed to create temp file for test, %v", err))
				}

				m, err := config.NewManifest([]string{mf.Name()})

				Expect(err).To(Not(HaveOccurred()))
				Expect(m.Layers).To(HaveLen(2))
				Expect(m.Layers[0].Name).To(Equal("layer1"))
				Expect(m.Layers[1].Name).To(Equal("layer2"))
			})

			It("can parse layers from both base and extending manifest", func() {
				base := config.Manifest{
					ExtendsConfig: config.ExtendsConfig{},
				}
				bf, err := NewTempManifestFile(base, "base-*.yaml")
				defer os.Remove(bf.Name())
				if err != nil {
					Fail(fmt.Sprintf("failed to create temp file for test, %v", err))
				}

				manifest := config.Manifest{
					ExtendsConfig: config.ExtendsConfig{
						Extends: bf.Name(),
					},
					Layers: []config.LayerConfig{
						{
							Name: "layer2",
						},
					},
				}
				mf, err := NewTempManifestFile(manifest, "racoon-*.yaml")
				defer os.Remove(mf.Name())
				if err != nil {
					Fail(fmt.Sprintf("failed to create temp file for test, %v", err))
				}

				m, err := config.NewManifest([]string{mf.Name()})

				Expect(err).To(Not(HaveOccurred()))
				Expect(m.Layers).To(HaveLen(1))
				Expect(m.Layers[0].Name).To(Equal("layer2"))
			})

			It("produces error when one layer exists in both manifests", func() {
				base := config.Manifest{
					ExtendsConfig: config.ExtendsConfig{},
					Layers: []config.LayerConfig{
						{
							Name: "layer3",
						},
					},
				}
				bf, err := NewTempManifestFile(base, "base-*.yaml")
				defer os.Remove(bf.Name())
				if err != nil {
					Fail(fmt.Sprintf("failed to create temp file for test, %v", err))
				}

				manifest := config.Manifest{
					ExtendsConfig: config.ExtendsConfig{
						Extends: bf.Name(),
					},
					Layers: []config.LayerConfig{
						{
							Name: "layer3",
						},
					},
				}
				mf, err := NewTempManifestFile(manifest, "racoon-*.yaml")
				defer os.Remove(mf.Name())
				if err != nil {
					Fail(fmt.Sprintf("failed to create temp file for test, %v", err))
				}

				_, err = config.NewManifest([]string{mf.Name()})

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("duplicate layer, layer3 defined multiple times"))
			})
		})
	})
})

func NewTempManifestFile(manifest config.Manifest, filepattern string) (file *os.File, err error) {
	if filepattern == "" {
		filepattern = "racoon-*.yaml"
	}

	tmpfile, err := os.CreateTemp("", "racoon-*.yaml")
	if err != nil {
		defer os.Remove(tmpfile.Name())
		return nil, err
	}

	b, _ := yaml.Marshal(manifest)
	if _, err = tmpfile.Write(b); err != nil {
		return nil, err
	}

	return tmpfile, nil
}
