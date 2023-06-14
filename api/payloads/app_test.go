package payloads_test

import (
	"net/url"

	"code.cloudfoundry.org/korifi/api/handlers"
	"code.cloudfoundry.org/korifi/api/payloads"
	"code.cloudfoundry.org/korifi/tools"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gstruct"
)

var _ = Describe("AppList", func() {
	Describe("DecodeFromURLValues", func() {
		appList := payloads.AppList{}
		err := appList.DecodeFromURLValues(url.Values{
			"names":       []string{"name"},
			"guids":       []string{"guid"},
			"space_guids": []string{"space_guid"},
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(appList).To(Equal(payloads.AppList{
			Names:      "name",
			GUIDs:      "guid",
			SpaceGuids: "space_guid",
		}))
	})
})

var _ = Describe("App payload validation", func() {
	var (
		decoderValidator *handlers.DecoderValidator
		validatorErr     error
	)

	BeforeEach(func() {
		var err error
		decoderValidator, err = handlers.NewDefaultDecoderValidator()
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("AppCreate", func() {
		var (
			payload        payloads.AppCreate
			decodedPayload *payloads.AppCreate
		)

		BeforeEach(func() {
			payload = payloads.AppCreate{
				Name: "my-app",
				Relationships: payloads.AppRelationships{
					Space: payloads.Relationship{
						Data: &payloads.RelationshipData{
							GUID: "app-guid",
						},
					},
				},
			}

			decodedPayload = new(payloads.AppCreate)
		})

		JustBeforeEach(func() {
			validatorErr = decoderValidator.DecodeAndValidateJSONPayload(createRequest(payload), decodedPayload)
		})

		It("succeeds", func() {
			Expect(validatorErr).NotTo(HaveOccurred())
			Expect(decodedPayload).To(gstruct.PointTo(Equal(payload)))
		})

		When("name is not set", func() {
			BeforeEach(func() {
				payload.Name = ""
			})

			It("returns an error", func() {
				expectUnprocessableEntityError(validatorErr, "name cannot be blank")
			})
		})

		When("lifecycle is invalid", func() {
			BeforeEach(func() {
				payload.Lifecycle = &payloads.Lifecycle{}
			})

			It("returns an unprocessable entity error", func() {
				expectUnprocessableEntityError(validatorErr, "lifecycle.type cannot be blank")
			})
		})

		When("relationships are not set", func() {
			BeforeEach(func() {
				payload.Relationships = payloads.AppRelationships{}
			})

			It("returns an unprocessable entity error", func() {
				expectUnprocessableEntityError(validatorErr, "relationships cannot be blank")
			})
		})

		When("metadata is invalid", func() {
			BeforeEach(func() {
				payload.Metadata = payloads.Metadata{
					Labels: map[string]string{
						"foo.cloudfoundry.org/bar": "jim",
					},
				}
			})

			It("returns an appropriate error", func() {
				expectUnprocessableEntityError(validatorErr, "label/annotation key cannot use the cloudfoundry.org domain")
			})
		})
	})

	Describe("AppPatch", func() {
		var (
			payload        payloads.AppPatch
			decodedPayload *payloads.AppPatch
		)

		BeforeEach(func() {
			payload = payloads.AppPatch{
				Metadata: payloads.MetadataPatch{
					Labels: map[string]*string{
						"foo": tools.PtrTo("bar"),
					},
					Annotations: map[string]*string{
						"example.org/jim": tools.PtrTo("hello"),
					},
				},
			}

			decodedPayload = new(payloads.AppPatch)
		})

		JustBeforeEach(func() {
			validatorErr = decoderValidator.DecodeAndValidateJSONPayload(createRequest(payload), decodedPayload)
		})

		It("succeeds", func() {
			Expect(validatorErr).NotTo(HaveOccurred())
			Expect(decodedPayload).To(gstruct.PointTo(Equal(payload)))
		})

		When("metadata is invalid", func() {
			BeforeEach(func() {
				payload.Metadata = payloads.MetadataPatch{
					Labels: map[string]*string{
						"foo.cloudfoundry.org/bar": tools.PtrTo("jim"),
					},
				}
			})

			It("returns an appropriate error", func() {
				expectUnprocessableEntityError(validatorErr, `Labels and annotations cannot begin with "cloudfoundry.org" or its subdomains`)
			})
		})
	})

	Describe("AppSetCurrentDroplet", func() {
		var (
			payload        payloads.AppSetCurrentDroplet
			decodedPayload *payloads.AppSetCurrentDroplet
		)

		BeforeEach(func() {
			payload = payloads.AppSetCurrentDroplet{
				Relationship: payloads.Relationship{
					Data: &payloads.RelationshipData{
						GUID: "the-guid",
					},
				},
			}

			decodedPayload = new(payloads.AppSetCurrentDroplet)
		})

		JustBeforeEach(func() {
			validatorErr = decoderValidator.DecodeAndValidateJSONPayload(createRequest(payload), decodedPayload)
		})

		It("succeeds", func() {
			Expect(validatorErr).NotTo(HaveOccurred())
			Expect(decodedPayload).To(gstruct.PointTo(Equal(payload)))
		})

		When("relationship is invalid", func() {
			BeforeEach(func() {
				payload.Relationship = payloads.Relationship{}
			})

			It("returns an appropriate error", func() {
				expectUnprocessableEntityError(validatorErr, "Relationship cannot be blank")
			})
		})
	})

	Describe("AppPatchEnvVars", func() {
		var (
			payload        payloads.AppPatchEnvVars
			decodedPayload *payloads.AppPatchEnvVars
		)

		BeforeEach(func() {
			payload = payloads.AppPatchEnvVars{
				Var: map[string]interface{}{
					"foo": "bar",
				},
			}

			decodedPayload = new(payloads.AppPatchEnvVars)
		})

		JustBeforeEach(func() {
			validatorErr = decoderValidator.DecodeAndValidateJSONPayload(createRequest(payload), decodedPayload)
		})

		It("succeeds", func() {
			Expect(validatorErr).NotTo(HaveOccurred())
			Expect(decodedPayload).To(gstruct.PointTo(Equal(payload)))
		})

		When("it contains a 'PORT' key", func() {
			BeforeEach(func() {
				payload.Var["PORT"] = "2222"
			})

			It("returns an appropriate error", func() {
				expectUnprocessableEntityError(validatorErr, "value PORT is not allowed")
			})
		})

		When("it contains a key with prefix 'VCAP_'", func() {
			BeforeEach(func() {
				payload.Var["VCAP_foo"] = "bar"
			})

			It("returns an appropriate error", func() {
				expectUnprocessableEntityError(validatorErr, "prefix VCAP_ is not allowed")
			})
		})

		When("it contains a key with prefix 'VMC_'", func() {
			BeforeEach(func() {
				payload.Var["VMC_foo"] = "bar"
			})

			It("returns an appropriate error", func() {
				expectUnprocessableEntityError(validatorErr, "prefix VMC_ is not allowed")
			})
		})
	})
})
