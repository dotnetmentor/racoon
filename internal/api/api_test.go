package api_test

import (
	"math/rand"
	"testing"

	"github.com/dotnetmentor/racoon/internal/api"
	"github.com/dotnetmentor/racoon/internal/config"
	"github.com/sirupsen/logrus"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestBooks(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Api Suite")
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandomString(n int) *string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	s := string(b)
	return &s
}

var _ = Describe("Properties", func() {
	Describe("NewProperty", func() {
		var expectedRuleConfig config.RuleConfig
		var expectedFormattingConfigArg []config.FormattingConfig

		var property api.Property
		var properties api.PropertyList
		var isNew bool

		When("property is first defined", func() {
			BeforeEach(func() {
				api.SetLogger(logrus.New())

				expectedRuleConfig = config.RuleConfig{
					Validation: config.ValidationRuleConfig{
						AllowEmpty: true,
					},
				}
				expectedFormattingConfigArg = []config.FormattingConfig{
					{
						Replace: RandomString(8),
						Source: &config.ValueSourceConfig{
							Literal: RandomString(8),
						},
					},
				}

				properties = api.PropertyList{}
				property, isNew = api.NewProperty(properties, "Property1", "description", "layer-1", true, expectedRuleConfig, expectedFormattingConfigArg)
				properties = api.PropertyList{property}
			})

			It("returns true for isNew", func() {
				Expect(isNew).To(BeTrue())
			})

			It("has description set from arguments", func() {
				Expect(property.Description).To(Equal("description"))
			})

			It("has source set from arguments", func() {
				Expect(property.Source()).To(Equal("layer-1"))
			})

			It("marks property as sensitive", func() {
				Expect(property.Sensitive()).To(BeTrue())
			})

			It("has rule config set from arguments", func() {
				Expect(property.Rules()).To(Equal(expectedRuleConfig))
			})

			It("has formatting config set from arguments", func() {
				Expect(property.Formatting()).To(Equal(expectedFormattingConfigArg))
			})

			It("has an empty values list", func() {
				Expect(property.Values()).To(Equal(api.ValueList{}))
			})
		})

		When("property is previously defined", func() {
			BeforeEach(func() {
				api.SetLogger(logrus.New())

				expectedRuleConfig = config.RuleConfig{
					Validation: config.ValidationRuleConfig{
						AllowEmpty: false,
					},
					Override: config.OverrideRuleConfig{
						AllowImplicit: true,
						AllowExplicit: true,
					},
				}
				expectedFormattingConfigArg = []config.FormattingConfig{
					{
						Replace: RandomString(8),
						Source: &config.ValueSourceConfig{
							Literal: RandomString(8),
						},
					},
				}

				properties = api.PropertyList{}
				p, _ := api.NewProperty(properties, "Property1", "description", "layer-1", true, expectedRuleConfig, []config.FormattingConfig{})
				properties = api.PropertyList{p}

				property, isNew = api.NewProperty(
					properties,
					"Property1",
					"new description",
					"layer-2",
					false,
					config.RuleConfig{
						Validation: config.ValidationRuleConfig{
							AllowEmpty: true,
						},
					},
					expectedFormattingConfigArg,
				)
			})

			It("returns false for isNew", func() {
				Expect(isNew).To(BeFalse())
			})

			It("has description copied from existing property", func() {
				Expect(property.Description).To(Equal("description"))
			})

			It("has source set from arguments", func() {
				Expect(property.Source()).To(Equal("layer-2"))
			})

			It("makes sure new property is sensitive when existing property is sensitive", func() {
				Expect(properties[0].Sensitive()).To(BeTrue())
				Expect(property.Sensitive()).To(BeTrue())
			})

			It("has rules copied from existing property", func() {
				Expect(property.Rules()).To(Equal(expectedRuleConfig))
			})

			It("has formatting config set from arguments", func() {
				Expect(property.Formatting()).To(Equal(expectedFormattingConfigArg))
			})

			It("has an empty values list", func() {
				Expect(property.Values()).To(Equal(api.ValueList{}))
			})
		})
	})
})
