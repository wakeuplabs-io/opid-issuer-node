package jsonschema

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	core "github.com/iden3/go-iden3-core/v2"
	jsonSuite "github.com/iden3/go-schema-processor/v2/json"
	"github.com/iden3/go-schema-processor/v2/merklize"
	"github.com/iden3/go-schema-processor/v2/processor"
	"github.com/mitchellh/mapstructure"
	"github.com/piprate/json-gold/ld"

	"github.com/wakeup-labs/issuer-node/internal/common"
	"github.com/wakeup-labs/issuer-node/internal/loader"
)

const (
	fakeUserDID   = "did:opid:optimism:sepolia:2qDDDKmo436EZGCBAvkqZjADYoNRJszkG7UymZeCHQ"
	fakeIssuerDID = "did:opid:optimism:sepolia:2qH7XAwYQzCp9VfhpNgeLtK2iCehDDrfMWUCEg5ig5"
)

// Attributes is a list of Attribute entities
type Attributes []Attribute

// SchemaAttrs converts jsonschema.Attributes into []string]
func (a Attributes) SchemaAttrs() []string {
	out := make([]string, len(a))
	for i, attr := range a {
		out[i] = attr.String()
	}
	return out
}

// Attribute represents a json schema attribute
type Attribute struct {
	ID         string
	Title      string
	Type       string
	Format     string
	Properties map[string]any `json:"-"`
}

func (a Attribute) String() string {
	if a.Title != "" {
		return fmt.Sprintf("%s(%s)", a.Title, a.ID)
	}
	return a.ID
}

// JSONSchema provides some methods to load a schema and do some inspections over it.
type JSONSchema struct {
	content map[string]any
}

// Load loads the json file
func Load(ctx context.Context, jsonSchemaURL string, loader loader.DocumentLoader) (*JSONSchema, error) {
	pr := processor.InitProcessorOptions(
		&processor.Processor{},
		processor.WithValidator(jsonSuite.Validator{}),
		processor.WithParser(jsonSuite.Parser{}),
		processor.WithDocumentLoader(loader))
	raw, err := pr.Load(ctx, jsonSchemaURL)
	if err != nil {
		return nil, err
	}

	schema := &JSONSchema{content: make(map[string]any)}
	if err := json.Unmarshal(raw, &schema.content); err != nil {
		return nil, err
	}
	return schema, nil
}

// Attributes returns a list with the attributes in properties.credentialSubject.properties
func (s *JSONSchema) Attributes() (Attributes, error) {
	var props map[string]any
	var ok bool
	props, ok = s.content["properties"].(map[string]any)
	if !ok {
		return nil, errors.New("missing properties field")
	}
	credSubject, ok := props["credentialSubject"].(map[string]any)
	if !ok {
		return nil, errors.New("missing properties.credentialSubject field")
	}
	props, ok = credSubject["properties"].(map[string]any)
	if !ok {
		return nil, errors.New("missing properties.credentialSubject.properties field")
	}
	attrs, err := processProperties(props)
	if err != nil {
		return nil, err
	}
	return attrs, nil
}

// AttributeByID returns the attribute with this id or an error if not found
func (s *JSONSchema) AttributeByID(id string) (*Attribute, error) {
	attrs, err := s.Attributes()
	if err != nil {
		return nil, err
	}
	i := findIndexForSchemaAttribute(attrs, id)
	if i < 0 {
		return nil, fmt.Errorf("schema attribute <%s> not found", id)
	}
	return &attrs[i], nil
}

// Bytes returns the json representation of the schema
func (s *JSONSchema) Bytes() ([]byte, error) {
	return json.Marshal(s.content)
}

// BytesNoErr returns the json representation of the schema ignoring errors, which should never happen
func (s *JSONSchema) BytesNoErr() []byte {
	b, _ := s.Bytes()
	return b
}

// JSONLdContext returns the value of $metadata.uris.jsonLdContext
func (s *JSONSchema) JSONLdContext() (string, error) {
	var metadata map[string]any
	var ok bool
	metadata, ok = s.content["$metadata"].(map[string]any)
	if !ok {
		return "", errors.New("missing $metadata field")
	}
	uris, ok := metadata["uris"].(map[string]any)
	if !ok {
		return "", errors.New("missing $metadata.uris field")
	}
	jsonLdContext, ok := uris["jsonLdContext"].(string)
	if !ok {
		return "", errors.New("missing $metadata.uris.jsonLdContext field")
	}
	return jsonLdContext, nil
}

// SchemaHash calculates the hash of a schemaType
func (s *JSONSchema) SchemaHash(schemaType string) (core.SchemaHash, error) {
	jsonLdContext, err := s.JSONLdContext()
	if err != nil {
		return core.SchemaHash{}, err
	}
	id := jsonLdContext + "#" + schemaType
	return common.CreateSchemaHash([]byte(id)), nil
}

// ValidateCredentialSubject validates that the given credential subject matches the given schema
func ValidateCredentialSubject(ctx context.Context, loader loader.DocumentLoader, schemaURL string, schemaType string, cSubject map[string]interface{}) error {
	schema, err := Load(ctx, schemaURL, loader)
	if err != nil {
		return err
	}

	schemaContext, err := schema.JSONLdContext()
	if err != nil {
		return err
	}

	dummyVC, err := createDummyVC(cSubject, schemaType, schemaContext)
	if err != nil {
		return err
	}

	err = validateDummyVCAgainstSchema(dummyVC, schema)
	if err != nil {
		return err
	}

	return validateDummyVCEntries(dummyVC, loader)
}

func validateDummyVCAgainstSchema(dummyVC map[string]interface{}, schema *JSONSchema) error {
	schemaBytes, err := json.Marshal(schema.content)
	if err != nil {
		return err
	}

	dummyVCBytes, err := json.Marshal(dummyVC)
	if err != nil {
		return err
	}

	validator := jsonSuite.Validator{}

	return validator.ValidateData(dummyVCBytes, schemaBytes)
}

func createDummyVC(cSubject map[string]interface{}, schemaType string, schemaContext string) (map[string]interface{}, error) {
	cSubject["id"] = fakeUserDID
	cSubject["type"] = schemaType

	vc := map[string]interface{}{
		"@context": []interface{}{"https://www.w3.org/2018/credentials/v1", "https://schema.iden3.io/core/jsonld/iden3proofs.jsonld", schemaContext},
		"credentialSchema": map[string]interface{}{
			"id":   "https://www.w3.org/2018/credentials/v1",
			"type": "JsonSchemaValidator2018",
		},
		"credentialStatus": map[string]interface{}{
			"id":   "testStatus",
			"type": "someType",
		},
		"credentialSubject": cSubject,
		"id":                "https://someURL/someschemaContext.jsonld",
		"issuanceDate":      "2023-04-27T23:45:29.498555+02:00",
		"issuer":            fakeIssuerDID,
		"type": []interface{}{
			"VerifiableCredential", schemaType,
		},
	}

	mBytes, err := json.Marshal(vc) // this marshal and unmarshal needs to be done for the validateDummyVCEntries function
	if err != nil {
		return nil, err
	}

	var resp map[string]interface{}
	err = json.Unmarshal(mBytes, &resp)

	return resp, err
}

func validateDummyVCEntries(vc map[string]interface{}, loader loader.DocumentLoader) error {
	proc := ld.NewJsonLdProcessor()
	options := ld.NewJsonLdOptions("")
	options.Algorithm = ld.AlgorithmURDNA2015
	options.SafeMode = true
	options.DocumentLoader = loader

	normDoc, err := proc.Normalize(vc, options)
	if err != nil {
		return err
	}

	dataset, ok := normDoc.(*ld.RDFDataset)
	if !ok {
		return errors.New("[assertion] expected *ld.RDFDataset type")
	}

	_, err = merklize.EntriesFromRDF(dataset)

	return err
}

func findIndexForSchemaAttribute(attributes Attributes, name string) int {
	for i, attribute := range attributes {
		if attribute.ID == name {
			return i
		}
	}
	return -1
}

func processProperties(props map[string]any) ([]Attribute, error) {
	attrs := make([]Attribute, 0, len(props))
	for id, prop := range props {
		attr := Attribute{}
		if err := mapstructure.Decode(prop, &attr); err != nil {
			return nil, fmt.Errorf("parsing attribute <%s>: %w", prop, err)
		}
		attr.ID = id
		if len(attr.Properties) > 0 {
			attrs = append(attrs, Attribute{
				ID:   id,
				Type: "object",
			})
			attrs1, err := processProperties(attr.Properties)
			if err != nil {
				return nil, err
			}
			attrs = append(attrs, attrs1...)
		} else {
			attrs = append(attrs, attr)
		}
	}
	return attrs, nil
}
