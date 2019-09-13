package jddf_test

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	jddf "github.com/jddf/jddf-go"
	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
)

func strptr(s string) *string {
	return &s
}

func TestVerifyAndForm(t *testing.T) {
	type testCase struct {
		in   string
		out  jddf.Schema
		err  error
		form jddf.Form
	}

	testCases := []testCase{
		{
			`{}`,
			jddf.Schema{},
			nil,
			jddf.FormEmpty,
		},
		{
			`{"ref":""}`,
			jddf.Schema{Ref: strptr("")},
			jddf.ErrNoSuchDefinition(""),
			jddf.FormRef,
		},
		{
			`{"definitions":{"":{}},"ref":""}`,
			jddf.Schema{
				Definitions: map[string]jddf.Schema{"": jddf.Schema{}},
				Ref:         strptr(""),
			},
			nil,
			jddf.FormRef,
		},
		{
			`{"definitions":{"":{}},"ref":"","type":"boolean"}`,
			jddf.Schema{
				Definitions: map[string]jddf.Schema{"": jddf.Schema{}},
				Ref:         strptr(""),
				Type:        jddf.TypeBoolean,
			},
			jddf.ErrInvalidForm,
			jddf.FormRef,
		},
		{
			`{"type":"boolean"}`,
			jddf.Schema{Type: jddf.TypeBoolean},
			nil,
			jddf.FormType,
		},
		{
			`{"type":"nonsense"}`,
			jddf.Schema{Type: "nonsense"},
			jddf.ErrInvalidType("nonsense"),
			jddf.FormType,
		},
		{
			`{"type":"boolean","enum":["a"]}`,
			jddf.Schema{Type: jddf.TypeBoolean, Enum: []string{"a"}},
			jddf.ErrInvalidForm,
			jddf.FormType,
		},
		{
			`{"enum":[]}`,
			jddf.Schema{Enum: []string{}},
			jddf.ErrEmptyEnum,
			jddf.FormEnum,
		},
		{
			`{"enum":["a","a"]}`,
			jddf.Schema{Enum: []string{"a", "a"}},
			jddf.ErrRepeatedEnumValue("a"),
			jddf.FormEnum,
		},
		{
			`{"enum":["a","b","c"]}`,
			jddf.Schema{Enum: []string{"a", "b", "c"}},
			nil,
			jddf.FormEnum,
		},
		{
			`{"enum":["a"],"properties":{}}`,
			jddf.Schema{Enum: []string{"a"}, RequiredProperties: map[string]jddf.Schema{}},
			jddf.ErrInvalidForm,
			jddf.FormEnum,
		},
		{
			`{"enum":["a"],"elements":{}}`,
			jddf.Schema{Enum: []string{"a"}, Elements: &jddf.Schema{}},
			jddf.ErrInvalidForm,
			jddf.FormEnum,
		},
		{
			`{"elements":{"ref":""}}`,
			jddf.Schema{Elements: &jddf.Schema{Ref: strptr("")}},
			jddf.ErrNoSuchDefinition(""),
			jddf.FormElements,
		},
		{
			`{"elements":{}}`,
			jddf.Schema{Elements: &jddf.Schema{}},
			nil,
			jddf.FormElements,
		},
		{
			`{"elements":{},"properties":{}}`,
			jddf.Schema{Elements: &jddf.Schema{}, RequiredProperties: map[string]jddf.Schema{}},
			jddf.ErrInvalidForm,
			jddf.FormElements,
		},
		{
			`{"elements":{},"optionalProperties":{}}`,
			jddf.Schema{Elements: &jddf.Schema{}, OptionalProperties: map[string]jddf.Schema{}},
			jddf.ErrInvalidForm,
			jddf.FormElements,
		},
		{
			`{"properties":{"a":{}},"optionalProperties":{"a":{}}}`,
			jddf.Schema{
				RequiredProperties: map[string]jddf.Schema{"a": jddf.Schema{}},
				OptionalProperties: map[string]jddf.Schema{"a": jddf.Schema{}},
			},
			jddf.ErrRepeatedProperty("a"),
			jddf.FormProperties,
		},
		{
			`{"properties":{"a":{"ref":""}}}`,
			jddf.Schema{
				RequiredProperties: map[string]jddf.Schema{"a": jddf.Schema{Ref: strptr("")}},
			},
			jddf.ErrNoSuchDefinition(""),
			jddf.FormProperties,
		},
		{
			`{"optionalProperties":{"a":{"ref":""}}}`,
			jddf.Schema{
				OptionalProperties: map[string]jddf.Schema{"a": jddf.Schema{Ref: strptr("")}},
			},
			jddf.ErrNoSuchDefinition(""),
			jddf.FormProperties,
		},
		{
			`{"properties":{"a":{}},"optionalProperties":{"b":{}}}`,
			jddf.Schema{
				RequiredProperties: map[string]jddf.Schema{"a": jddf.Schema{}},
				OptionalProperties: map[string]jddf.Schema{"b": jddf.Schema{}},
			},
			nil,
			jddf.FormProperties,
		},
		{
			`{"properties":{},"values":{}}`,
			jddf.Schema{
				RequiredProperties: map[string]jddf.Schema{},
				Values:             &jddf.Schema{},
			},
			jddf.ErrInvalidForm,
			jddf.FormProperties,
		},
		{
			`{"values":{"ref":""}}`,
			jddf.Schema{Values: &jddf.Schema{Ref: strptr("")}},
			jddf.ErrNoSuchDefinition(""),
			jddf.FormValues,
		},
		{
			`{"values":{}}`,
			jddf.Schema{Values: &jddf.Schema{}},
			nil,
			jddf.FormValues,
		},
		{
			`{"values":{},"discriminator":{"mapping":{}}}`,
			jddf.Schema{
				Values:        &jddf.Schema{},
				Discriminator: jddf.Discriminator{Mapping: map[string]jddf.Schema{}},
			},
			jddf.ErrInvalidForm,
			jddf.FormValues,
		},
		{
			`{"discriminator":{"mapping":{"":{}}}}`,
			jddf.Schema{
				Discriminator: jddf.Discriminator{Mapping: map[string]jddf.Schema{
					"": jddf.Schema{},
				}},
			},
			jddf.ErrNonPropertiesMapping,
			jddf.FormDiscriminator,
		},
		{
			`{"discriminator":{"tag":"a","mapping":{"":{"properties":{"a":{}}}}}}`,
			jddf.Schema{
				Discriminator: jddf.Discriminator{
					Tag: "a",
					Mapping: map[string]jddf.Schema{
						"": jddf.Schema{
							RequiredProperties: map[string]jddf.Schema{
								"a": jddf.Schema{},
							},
						},
					},
				},
			},
			jddf.ErrRepeatedTagInProperties("a"),
			jddf.FormDiscriminator,
		},
		{
			`{"discriminator":{"tag":"a","mapping":{"":{"optionalProperties":{"a":{}}}}}}`,
			jddf.Schema{
				Discriminator: jddf.Discriminator{
					Tag: "a",
					Mapping: map[string]jddf.Schema{
						"": jddf.Schema{
							OptionalProperties: map[string]jddf.Schema{
								"a": jddf.Schema{},
							},
						},
					},
				},
			},
			jddf.ErrRepeatedTagInProperties("a"),
			jddf.FormDiscriminator,
		},
		{
			`{"discriminator":{"tag":"a","mapping":{"":{"properties":{"b":{}}}}}}`,
			jddf.Schema{
				Discriminator: jddf.Discriminator{
					Tag: "a",
					Mapping: map[string]jddf.Schema{
						"": jddf.Schema{
							RequiredProperties: map[string]jddf.Schema{
								"b": jddf.Schema{},
							},
						},
					},
				},
			},
			nil,
			jddf.FormDiscriminator,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.in, func(t *testing.T) {
			var out jddf.Schema
			assert.NoError(t, json.Unmarshal([]byte(tt.in), &out))

			assert.Equal(t, tt.out, out)
			assert.Equal(t, tt.err, out.Verify())
			assert.Equal(t, tt.form, out.Form())
		})
	}
}

func TestSpecInvalidSchemas(t *testing.T) {
	type testCase struct {
		Name   string      `json:"name"`
		Schema interface{} `json:"schema"`
	}

	data, err := ioutil.ReadFile("spec/tests/invalid-schemas.json")
	assert.NoError(t, err)

	var testCases []testCase
	assert.NoError(t, json.Unmarshal(data, &testCases))

	for _, tt := range testCases {
		// We skip this test case because though it can be handled (by making Tag a
		// *string), the result is a less useful library with an otherwise useless
		// allocation.
		if tt.Name == "discriminator has no tag" {
			continue
		}

		// The mapstructure package just treats "null" as being a no-op to parse
		// from, leaving the schema as its zero value. This is not really an
		// interesting bug, nor one that can readily be worked around.
		if tt.Name == "null schema" {
			continue
		}

		// The mapstructure package has a similar issue with empty arrays.
		if tt.Name == "enum empty array" {
			continue
		}

		t.Run(tt.Name, func(t *testing.T) {
			var schema jddf.Schema

			decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
				TagName: "json",
				Result:  &schema,
			})

			assert.NoError(t, err)

			// An error while decoding is acceptable.
			if err := decoder.Decode(tt.Schema); err != nil {
				return
			}

			// If no error occurred while decoding, then verify that the resulting
			// schema is considered invalid.
			assert.Error(t, schema.Verify())
		})
	}
}
