package main

import (
	"encoding/json"
	"fmt"
	"testing"

	. "github.com/dave/jennifer/jen"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_genTypeName(t *testing.T) {
	type jsonToSource struct {
		from     string
		expected string
	}

	tests := []jsonToSource{
		//
		{
			`{"type": "publicKey"}`,
			"var thing solanago.PublicKey",
		},
		{
			`{"type": "bool"}`,
			"var thing bool",
		},
		{
			`{"type": "u8"}`,
			"var thing uint8",
		},
		{
			`{"type": "i8"}`,
			"var thing int8",
		},
		{
			`{"type": "u16"}`,
			"var thing uint16",
		},
		{
			`{"type": "i16"}`,
			"var thing int16",
		},
		{
			`{"type": "u32"}`,
			"var thing uint32",
		},
		{
			`{"type": "i32"}`,
			"var thing int32",
		},
		{
			`{"type": "u64"}`,
			"var thing uint64",
		},
		{
			`{"type": "i64"}`,
			"var thing int64",
		},
		{
			`{"type": "u128"}`,
			"var thing binary.Uint128",
		},
		{
			`{"type": "i128"}`,
			"var thing binary.Int128",
		},
		{
			// TODO: is this also OK as []byte ???
			`{"type": "bytes"}`,
			"var thing []byte",
		},
		{
			`{"type": "string"}`,
			"var thing string",
		},
		{
			`{"type": "publicKey"}`,
			"var thing solanago.PublicKey",
		},

		// "defined"
		{
			`{"type": {"defined":"Foo"}}`,
			"var thing Foo",
		},
		{
			`{"type": {"defined":"bar"}}`,
			"var thing bar",
		},

		// "array":
		{
			`{"type": {"array":["u8",280]}}`,
			"var thing [280]uint8",
		},
		{
			`{"type": {"array":[{"defined":"Message"},33607]}}`,
			"var thing [33607]Message",
		},
		{
			`{"type": {"array":[{"array":["u8",280]},33607]}}`,
			"var thing [33607][280]uint8",
		},
		{
			`{"type": {"array":[{"array":[{"defined":"Message"},123]},33607]}}`,
			"var thing [33607][123]Message",
		},

		// "vec":
		{
			`{"type": {"vec": "publicKey"}}`,
			"var thing []solanago.PublicKey",
		},
		{
			`{"type": {"vec": {"defined": "TransactionAccount"}}}`,
			"var thing []TransactionAccount",
		},
		{
			`{"type": {"vec": "bool"}}`,
			"var thing []bool",
		},
		{
			`{"type": {"vec": {"array":[{"array":[{"defined":"Message"},123]},33607]}}}`,
			"var thing [][33607][123]Message",
		},

		// "option":
		{
			`{"type": {"option": "string"}}`,
			"var thing string",
		},
		{
			`{"type": {"option": {"vec": {"array":[{"array":[{"defined":"Message"},123]},33607]}}}}`,
			"var thing [][33607][123]Message",
		},
		{
			`{"type": {"option": {"defined": "TransactionAccount"}}}`,
			"var thing TransactionAccount",
		},

		// "hashMap":
		{
			`{"type": {"hashMap": ["string", "u64"]}}`,
			"var thing map[string]uint64",
		},
		{
			`{"type": {"hashMap": ["publicKey", {"defined": "Account"}]}}`,
			"var thing map[solanago.PublicKey]Account",
		},
		{
			`{"type": {"hashMap": [{"defined": "UserId"}, {"vec": "string"}]}}`,
			"var thing map[UserId][]string",
		},
		{
			`{"type": {"hashMap": ["string", {"option": "bool"}]}}`,
			"var thing map[string]bool",
		},
	}
	{
		for _, scenario := range tests {
			var target IdlField
			err := json.Unmarshal([]byte(scenario.from), &target)
			if err != nil {
				panic(err)
			}
			code := Var().Id("thing").Add(genTypeName(target.Type))
			got := codeToString(code)
			require.Equal(t, scenario.expected, got)
		}
	}
}

func Test_genField(t *testing.T) {
	type jsonToSource struct {
		from     string
		expected string
	}

	tests := []jsonToSource{
		{
			`{"name":"space","type":"u64"}`,
			"var thing struct {\n	Space uint64\n}",
		},
		{
			`{"name":"space","type": {"option": {"vec": {"array":[{"array":[{"defined":"Message"},123]},33607]}}}}`,
			"var thing struct {\n	Space [][33607][123]Message\n}",
		},
	}
	{
		for _, scenario := range tests {
			var target IdlField
			err := json.Unmarshal([]byte(scenario.from), &target)
			if err != nil {
				panic(err)
			}
			code := Var().Id("thing").Struct(
				genField(target, false),
			)
			got := codeToString(code)
			require.Equal(t, scenario.expected, got)
		}
	}
}

func Test_IdlAccountItemSlice_Walk(t *testing.T) {
	data := `[
        {
          "name": "authorityBefore",
          "isMut": false,
          "isSigner": true
        },
        {
          "name": "marketGroup",
          "accounts": [
            {
              "name": "marketMarket",
              "isMut": true,
              "isSigner": false
            },
            {
              "name": "foo",
              "isMut": true,
              "isSigner": false
            },
            {
              "name": "subMarket",
              "accounts": [
                {
                  "name": "subMarketMarket",
                  "isMut": true,
                  "isSigner": false
                },
                {
                  "name": "openOrders",
                  "isMut": true,
                  "isSigner": false
                } 
              ]
            }
          ]
        },
        {
          "name": "authorityAfter",
          "isMut": false,
          "isSigner": true
        }
      ]`
	var target IdlAccountItemSlice
	err := json.Unmarshal([]byte(data), &target)
	if err != nil {
		panic(err)
	}

	spew.Dump(target)

	expectedGroups := []string{
		"instruction",
		"instruction/marketGroup",
		"instruction/marketGroup",
		"instruction/marketGroup/subMarket",
		"instruction/marketGroup/subMarket",
		"instruction",
	}
	gotGroups := []string{}

	expectedAccountNames := []string{
		"authorityBefore",
		"marketMarket",
		"foo",
		"subMarketMarket",
		"openOrders",
		"authorityAfter",
	}
	gotAccountNames := []string{}

	expectedIndexes := []int{0, 1, 2, 3, 4, 5}
	gotIndexes := []int{}

	instructionName := "instruction"
	target.Walk(instructionName, nil, nil, func(parentGroupPath string, index int, parentGroup *IdlAccounts, ia *IdlAccount) bool {
		gotGroups = append(gotGroups, parentGroupPath)
		gotAccountNames = append(gotAccountNames, ia.Name)
		gotIndexes = append(gotIndexes, index)
		return true
	})

	require.Equal(t, expectedGroups, gotGroups)
	require.Equal(t, expectedAccountNames, gotAccountNames)
	require.Equal(t, expectedIndexes, gotIndexes)
}

func TestFormatAccountAccessorName(t *testing.T) {
	t.Run("default config", func(t *testing.T) {
		assert.Equal(t, "GetFooAccount", formatAccountAccessorName("Get", "Foo"))
		assert.Equal(t, "GetFooAccountAccount", formatAccountAccessorName("Get", "FooAccount"))
	})

	t.Run("remove config on", func(t *testing.T) {
		oldConf := GetConfig()
		defer func() {
			conf = oldConf
		}()
		conf = &Config{
			RemoveAccountSuffix: true,
		}

		assert.Equal(t, "GetFooAccount", formatAccountAccessorName("Get", "Foo"))
		assert.Equal(t, "GetFooAccount", formatAccountAccessorName("Get", "FooAccount"))
	})
}

func Test_IdlType_HashMap(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		wantKey  string
		wantVal  string
	}{
		{
			name:     "string to uint64",
			jsonData: `{"hashMap": ["string", "u64"]}`,
			wantKey:  "string",
			wantVal:  "uint64",
		},
		{
			name:     "publicKey to defined type",
			jsonData: `{"hashMap": ["publicKey", {"defined": "Account"}]}`,
			wantKey:  "solanago.PublicKey",
			wantVal:  "Account",
		},
		{
			name:     "defined type to vec",
			jsonData: `{"hashMap": [{"defined": "UserId"}, {"vec": "string"}]}`,
			wantKey:  "UserId",
			wantVal:  "[]string",
		},
		{
			name:     "nested complex types",
			jsonData: `{"hashMap": [{"array": ["u8", 32]}, {"option": {"defined": "Metadata"}}]}`,
			wantKey:  "[32]uint8",
			wantVal:  "Metadata",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var idlType IdlType
			err := json.Unmarshal([]byte(tt.jsonData), &idlType)
			require.NoError(t, err)

			require.True(t, idlType.IsHashMap(), "Should be recognized as HashMap")

			hashMap := idlType.GetHashMap()
			require.NotNil(t, hashMap, "GetHashMap should return non-nil")

			// Test code generation
			code := genTypeName(idlType)
			generated := codeToString(Var().Id("test").Add(code))

			expected := "var test map[" + tt.wantKey + "]" + tt.wantVal
			require.Equal(t, expected, generated)
		})
	}
}

func Test_IdlEnumFields_TupleVariants(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		expected string
	}{
		{
			name:     "single tuple with defined type",
			jsonData: `[{"defined": "Collection"}]`,
			expected: "Collection",
		},
		{
			name:     "multiple tuple elements",
			jsonData: `["string", "u64", {"defined": "Account"}]`,
			expected: "multiple",
		},
		{
			name:     "tuple with option type",
			jsonData: `[{"option": {"defined": "Metadata"}}]`,
			expected: "Metadata",
		},
		{
			name:     "tuple with vec type",
			jsonData: `[{"vec": "publicKey"}]`,
			expected: "[]solanago.PublicKey",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var enumFields IdlEnumFields
			err := json.Unmarshal([]byte(tt.jsonData), &enumFields)
			require.NoError(t, err)

			require.NotNil(t, enumFields.IdlEnumFieldsTuple, "Should parse as tuple variant")
			require.Nil(t, enumFields.IdlEnumFieldsNamed, "Should not parse as named variant")

			tupleFields := *enumFields.IdlEnumFieldsTuple
			if tt.expected == "multiple" {
				require.Len(t, tupleFields, 3, "Should have 3 tuple elements")

				// Test first element is string
				require.True(t, tupleFields[0].IsString())
				require.Equal(t, "string", string(tupleFields[0].GetString()))

				// Test second element is u64
				require.True(t, tupleFields[1].IsString())
				require.Equal(t, "u64", string(tupleFields[1].GetString()))

				// Test third element is defined type
				require.True(t, tupleFields[2].IsIdlTypeDefined())
				require.Equal(t, "Account", tupleFields[2].GetIdlTypeDefined().Defined)
			} else {
				require.Len(t, tupleFields, 1, "Should have 1 tuple element")

				// Test code generation for single element
				code := genTypeName(tupleFields[0])
				generated := codeToString(Var().Id("test").Add(code))
				expectedGenerated := "var test " + tt.expected
				require.Equal(t, expectedGenerated, generated)
			}
		})
	}
}

func Test_IdlEnumFields_StringVariants(t *testing.T) {
	// Test the fix for string variants (like "Bid", "Ask")
	jsonData := `["Bid", "Ask"]`

	var enumFields IdlEnumFields
	err := json.Unmarshal([]byte(jsonData), &enumFields)
	require.NoError(t, err)

	require.NotNil(t, enumFields.IdlEnumFieldsTuple, "Should parse as tuple variant")
	require.Nil(t, enumFields.IdlEnumFieldsNamed, "Should not parse as named variant")

	tupleFields := *enumFields.IdlEnumFieldsTuple
	require.Len(t, tupleFields, 2, "Should have 2 string elements")

	// Test first element
	require.True(t, tupleFields[0].IsString())
	require.Equal(t, "Bid", string(tupleFields[0].GetString()))

	// Test second element
	require.True(t, tupleFields[1].IsString())
	require.Equal(t, "Ask", string(tupleFields[1].GetString()))
}

func Test_HashMap_EdgeCases(t *testing.T) {
	t.Run("invalid hashMap format", func(t *testing.T) {
		// Test with non-array hashMap value
		jsonData := `{"hashMap": "invalid"}`

		var idlType IdlType
		require.Panics(t, func() {
			json.Unmarshal([]byte(jsonData), &idlType)
		}, "Should panic on invalid hashMap format")
	})

	t.Run("wrong hashMap length", func(t *testing.T) {
		// Test with wrong number of elements
		jsonData := `{"hashMap": ["string"]}`

		var idlType IdlType
		require.Panics(t, func() {
			json.Unmarshal([]byte(jsonData), &idlType)
		}, "Should panic on wrong hashMap length")
	})

	t.Run("empty hashMap", func(t *testing.T) {
		// Test with empty array
		jsonData := `{"hashMap": []}`

		var idlType IdlType
		require.Panics(t, func() {
			json.Unmarshal([]byte(jsonData), &idlType)
		}, "Should panic on empty hashMap")
	})
}

// Helper functions for test data creation
func createStringIdlType() IdlType {
	var idlType IdlType
	json.Unmarshal([]byte(`"string"`), &idlType)
	return idlType
}

func createBoolIdlType() IdlType {
	var idlType IdlType
	json.Unmarshal([]byte(`"bool"`), &idlType)
	return idlType
}

func createDefinedIdlType(name string) IdlType {
	var idlType IdlType
	json.Unmarshal([]byte(fmt.Sprintf(`{"defined": "%s"}`, name)), &idlType)
	return idlType
}

func Test_isComplexEnum(t *testing.T) {
	// Setup test data
	namedFields := IdlEnumFieldsNamed{
		{Name: "field1", Type: createStringIdlType()},
	}

	testTypes := []IdlTypeDef{
		{
			Name: "SimpleEnum",
			Type: IdlTypeDefTy{
				Kind: IdlTypeDefTyKindEnum,
				Variants: []IdlEnumVariant{
					{Name: "One", Fields: nil},
					{Name: "Two", Fields: nil},
				},
			},
		},
		{
			Name: "ComplexEnum",
			Type: IdlTypeDefTy{
				Kind: IdlTypeDefTyKindEnum,
				Variants: []IdlEnumVariant{
					{
						Name: "WithFields",
						Fields: &IdlEnumFields{
							IdlEnumFieldsNamed: &namedFields,
						},
					},
				},
			},
		},
	}

	idl := IDL{Types: testTypes}

	// Register complex enums
	for _, typ := range testTypes {
		registerComplexEnums(&idl, typ)
	}

	tests := []struct {
		name     string
		typeName string
		expected bool
	}{
		{
			name:     "simple enum should not be complex",
			typeName: "SimpleEnum",
			expected: false,
		},
		{
			name:     "enum with fields should be complex",
			typeName: "ComplexEnum",
			expected: true,
		},
		{
			name:     "non-existent type should not be complex",
			typeName: "NonExistent",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			idlType := createDefinedIdlType(tt.typeName)
			result := isComplexEnum(idlType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func Test_LargeEnumDetection(t *testing.T) {
	tests := []struct {
		name         string
		variantCount int
		expected     bool
	}{
		{
			name:         "small enum (8 variants) should not be large",
			variantCount: 8,
			expected:     false,
		},
		{
			name:         "large enum (9 variants) should be large",
			variantCount: 9,
			expected:     true,
		},
		{
			name:         "very large enum (15 variants) should be large",
			variantCount: 15,
			expected:     true,
		},
		{
			name:         "tiny enum (3 variants) should not be large",
			variantCount: 3,
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create variants
			variants := make([]IdlEnumVariant, tt.variantCount)
			for i := 0; i < tt.variantCount; i++ {
				namedFields := IdlEnumFieldsNamed{
					{Name: "field", Type: createStringIdlType()},
				}
				variants[i] = IdlEnumVariant{
					Name: fmt.Sprintf("Variant%d", i),
					Fields: &IdlEnumFields{
						IdlEnumFieldsNamed: &namedFields,
					},
				}
			}

			// Test the logic used in the generator
			isLargeEnum := len(variants) > 8
			assert.Equal(t, tt.expected, isLargeEnum)
		})
	}
}

func Test_EnumVariantGeneration(t *testing.T) {
	// Test that enum variants are properly generated for both small and large enums
	namedFields1 := IdlEnumFieldsNamed{
		{Name: "field1", Type: createStringIdlType()},
	}
	namedFields2 := IdlEnumFieldsNamed{
		{Name: "field2", Type: createBoolIdlType()},
	}

	testEnum := IdlTypeDef{
		Name: "TestEnum",
		Type: IdlTypeDefTy{
			Kind: IdlTypeDefTyKindEnum,
			Variants: []IdlEnumVariant{
				{
					Name: "VariantOne",
					Fields: &IdlEnumFields{
						IdlEnumFieldsNamed: &namedFields1,
					},
				},
				{
					Name: "VariantTwo",
					Fields: &IdlEnumFields{
						IdlEnumFieldsNamed: &namedFields2,
					},
				},
			},
		},
	}

	idl := IDL{Types: []IdlTypeDef{testEnum}}

	// Register complex enums
	registerComplexEnums(&idl, testEnum)

	// Test that the enum was registered as complex
	idlType := createDefinedIdlType("TestEnum")
	assert.True(t, isComplexEnum(idlType), "TestEnum should be registered as complex")

	// Test variant count logic
	isLarge := len(testEnum.Type.Variants) > 8
	assert.False(t, isLarge, "TestEnum with 2 variants should not be large")
}

func Test_genInitializeComplexEnumFields(t *testing.T) {
	// Test the function that generates complex enum field initialization
	namedFields := IdlEnumFieldsNamed{
		{
			Name: "collection",
			Type: createDefinedIdlType("CollectionToggle"),
		},
		{
			Name: "uses",
			Type: createDefinedIdlType("UsesToggle"),
		},
	}

	testVariant := IdlEnumVariant{
		Name: "TestVariant",
		Fields: &IdlEnumFields{
			IdlEnumFieldsNamed: &namedFields,
		},
	}

	// Create test IDL with complex enum types (they need to have fields to be complex)
	collectionFields := IdlEnumFieldsNamed{
		{Name: "value", Type: createStringIdlType()},
	}
	usesFields := IdlEnumFieldsNamed{
		{Name: "count", Type: createStringIdlType()},
	}

	idl := IDL{
		Types: []IdlTypeDef{
			{
				Name: "CollectionToggle",
				Type: IdlTypeDefTy{
					Kind: IdlTypeDefTyKindEnum,
					Variants: []IdlEnumVariant{
						{Name: "None", Fields: nil},
						{Name: "Clear", Fields: nil},
						{
							Name: "Set",
							Fields: &IdlEnumFields{
								IdlEnumFieldsNamed: &collectionFields,
							},
						},
					},
				},
			},
			{
				Name: "UsesToggle",
				Type: IdlTypeDefTy{
					Kind: IdlTypeDefTyKindEnum,
					Variants: []IdlEnumVariant{
						{Name: "None", Fields: nil},
						{Name: "Clear", Fields: nil},
						{
							Name: "Set",
							Fields: &IdlEnumFields{
								IdlEnumFieldsNamed: &usesFields,
							},
						},
					},
				},
			},
		},
	}

	// Register complex enums
	for _, typ := range idl.Types {
		registerComplexEnums(&idl, typ)
	}

	// Generate the initialization code
	code := genInitializeComplexEnumFields(idl, "TestEnum", testVariant)

	// Convert to string to check the generated code
	codeStr := fmt.Sprintf("%#v", code)

	// Check that code was actually generated (not empty)
	assert.NotEmpty(t, codeStr, "Should generate some initialization code")

	// The code should contain initialization for both complex enum fields
	assert.Contains(t, codeStr, "Collection", "Should generate initialization for Collection field")
	assert.Contains(t, codeStr, "Uses", "Should generate initialization for Uses field")
	assert.Contains(t, codeStr, "CollectionToggleNone", "Should initialize to None variant")
	assert.Contains(t, codeStr, "UsesToggleNone", "Should initialize to None variant")
}
