package openapi

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type Parser struct {
	content []byte
}

func NewParser(content []byte) *Parser {
	return &Parser{
		content: content,
	}
}

func (p *Parser) Parse() (*OpenAPISpec, error) {
	// Parse the YAML/JSON into a map first
	var raw map[string]interface{}
	if err := yaml.Unmarshal(p.content, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse OpenAPI spec: %w", err)
	}

	spec := &OpenAPISpec{
		Paths: make(map[string]*PathItem),
	}

	// Parse openapi version
	if openapi, ok := raw["openapi"].(string); ok {
		spec.OpenAPI = openapi
	}

	// Parse info
	if info, ok := raw["info"].(map[string]interface{}); ok {
		spec.Info = p.parseInfo(info)
	}

	// Parse servers
	if servers, ok := raw["servers"].([]interface{}); ok {
		spec.Servers = p.parseServers(servers)
	}

	// Parse paths
	if paths, ok := raw["paths"].(map[string]interface{}); ok {
		for path, pathItem := range paths {
			if pathItemMap, ok := pathItem.(map[string]interface{}); ok {
				spec.Paths[path] = p.parsePathItem(pathItemMap)
			}
		}
	}

	// Parse components
	if components, ok := raw["components"].(map[string]interface{}); ok {
		spec.Components = p.parseComponents(components)
	}

	// Parse tags
	if tags, ok := raw["tags"].([]interface{}); ok {
		spec.Tags = p.parseTags(tags)
	}

	return spec, nil
}

func (p *Parser) parseInfo(info map[string]interface{}) *Info {
	i := &Info{}

	if title, ok := info["title"].(string); ok {
		i.Title = title
	}
	if version, ok := info["version"].(string); ok {
		i.Version = version
	}
	if description, ok := info["description"].(string); ok {
		i.Description = description
	}
	if contact, ok := info["contact"].(map[string]interface{}); ok {
		i.Contact = p.parseContact(contact)
	}

	return i
}

func (p *Parser) parseContact(contact map[string]interface{}) *Contact {
	c := &Contact{}

	if name, ok := contact["name"].(string); ok {
		c.Name = name
	}
	if email, ok := contact["email"].(string); ok {
		c.Email = email
	}
	if url, ok := contact["url"].(string); ok {
		c.URL = url
	}

	return c
}

func (p *Parser) parseServers(servers []interface{}) []*Server {
	var result []*Server

	for _, server := range servers {
		if serverMap, ok := server.(map[string]interface{}); ok {
			s := &Server{}
			if url, ok := serverMap["url"].(string); ok {
				s.URL = url
			}
			if description, ok := serverMap["description"].(string); ok {
				s.Description = description
			}
			result = append(result, s)
		}
	}

	return result
}

func (p *Parser) parsePathItem(pathItem map[string]interface{}) *PathItem {
	pi := &PathItem{}

	if ref, ok := pathItem["$ref"].(string); ok {
		pi.Ref = ref
	}
	if summary, ok := pathItem["summary"].(string); ok {
		pi.Summary = summary
	}
	if description, ok := pathItem["description"].(string); ok {
		pi.Description = description
	}

	// Parse operations
	if get, ok := pathItem["get"].(map[string]interface{}); ok {
		pi.Get = p.parseOperation(get)
	}
	if post, ok := pathItem["post"].(map[string]interface{}); ok {
		pi.Post = p.parseOperation(post)
	}
	if put, ok := pathItem["put"].(map[string]interface{}); ok {
		pi.Put = p.parseOperation(put)
	}
	if delete, ok := pathItem["delete"].(map[string]interface{}); ok {
		pi.Delete = p.parseOperation(delete)
	}
	if patch, ok := pathItem["patch"].(map[string]interface{}); ok {
		pi.Patch = p.parseOperation(patch)
	}

	return pi
}

func (p *Parser) parseOperation(operation map[string]interface{}) *Operation {
	op := &Operation{
		Responses: make(map[string]*Response),
	}

	if operationID, ok := operation["operationId"].(string); ok {
		op.OperationID = operationID
	}
	if summary, ok := operation["summary"].(string); ok {
		op.Summary = summary
	}
	if description, ok := operation["description"].(string); ok {
		op.Description = description
	}

	// Parse tags
	if tags, ok := operation["tags"].([]interface{}); ok {
		for _, tag := range tags {
			if tagStr, ok := tag.(string); ok {
				op.Tags = append(op.Tags, tagStr)
			}
		}
	}

	// Parse parameters
	if parameters, ok := operation["parameters"].([]interface{}); ok {
		op.Parameters = p.parseParameters(parameters)
	}

	// Parse requestBody
	if requestBody, ok := operation["requestBody"].(map[string]interface{}); ok {
		op.RequestBody = p.parseRequestBody(requestBody)
	}

	// Parse responses
	if responses, ok := operation["responses"].(map[string]interface{}); ok {
		for statusCode, response := range responses {
			if responseMap, ok := response.(map[string]interface{}); ok {
				op.Responses[statusCode] = p.parseResponse(responseMap)
			}
		}
	}

	return op
}

func (p *Parser) parseParameters(parameters []interface{}) []*Parameter {
	var result []*Parameter

	for _, param := range parameters {
		if paramMap, ok := param.(map[string]interface{}); ok {
			par := &Parameter{}

			if name, ok := paramMap["name"].(string); ok {
				par.Name = name
			}
			if in, ok := paramMap["in"].(string); ok {
				par.In = in
			}
			if description, ok := paramMap["description"].(string); ok {
				par.Description = description
			}
			if required, ok := paramMap["required"].(bool); ok {
				par.Required = required
			}
			if schema, ok := paramMap["schema"].(map[string]interface{}); ok {
				par.Schema = p.parseSchema(schema)
			}
			if example := paramMap["example"]; example != nil {
				par.Example = example
			}

			result = append(result, par)
		}
	}

	return result
}

func (p *Parser) parseRequestBody(requestBody map[string]interface{}) *RequestBody {
	rb := &RequestBody{
		Content: make(map[string]*MediaType),
	}

	if description, ok := requestBody["description"].(string); ok {
		rb.Description = description
	}
	if required, ok := requestBody["required"].(bool); ok {
		rb.Required = required
	}
	if content, ok := requestBody["content"].(map[string]interface{}); ok {
		for mediaType, mediaTypeObj := range content {
			if mediaTypeMap, ok := mediaTypeObj.(map[string]interface{}); ok {
				rb.Content[mediaType] = p.parseMediaType(mediaTypeMap)
			}
		}
	}

	return rb
}

func (p *Parser) parseResponse(response map[string]interface{}) *Response {
	r := &Response{
		Content: make(map[string]*MediaType),
		Headers: make(map[string]*Header),
	}

	if description, ok := response["description"].(string); ok {
		r.Description = description
	}
	if content, ok := response["content"].(map[string]interface{}); ok {
		for mediaType, mediaTypeObj := range content {
			if mediaTypeMap, ok := mediaTypeObj.(map[string]interface{}); ok {
				r.Content[mediaType] = p.parseMediaType(mediaTypeMap)
			}
		}
	}

	return r
}

func (p *Parser) parseMediaType(mediaType map[string]interface{}) *MediaType {
	mt := &MediaType{}

	if schema, ok := mediaType["schema"].(map[string]interface{}); ok {
		mt.Schema = p.parseSchema(schema)
	}
	if example := mediaType["example"]; example != nil {
		mt.Example = example
	}

	return mt
}

func (p *Parser) parseSchema(schema map[string]interface{}) *Schema {
	s := &Schema{
		Properties: make(map[string]*Schema),
	}

	if ref, ok := schema["$ref"].(string); ok {
		s.Ref = ref
	}
	if typ, ok := schema["type"].(string); ok {
		s.Type = typ
	}
	if format, ok := schema["format"].(string); ok {
		s.Format = format
	}
	if title, ok := schema["title"].(string); ok {
		s.Title = title
	}
	if description, ok := schema["description"].(string); ok {
		s.Description = description
	}
	if nullable, ok := schema["nullable"].(bool); ok {
		s.Nullable = nullable
	}
	if readOnly, ok := schema["readOnly"].(bool); ok {
		s.ReadOnly = readOnly
	}
	if writeOnly, ok := schema["writeOnly"].(bool); ok {
		s.WriteOnly = writeOnly
	}

	// Parse properties
	if properties, ok := schema["properties"].(map[string]interface{}); ok {
		for propName, propSchema := range properties {
			if propSchemaMap, ok := propSchema.(map[string]interface{}); ok {
				s.Properties[propName] = p.parseSchema(propSchemaMap)
			}
		}
	}

	// Parse required
	if required, ok := schema["required"].([]interface{}); ok {
		for _, req := range required {
			if reqStr, ok := req.(string); ok {
				s.Required = append(s.Required, reqStr)
			}
		}
	}

	// Parse items (for arrays)
	if items, ok := schema["items"].(map[string]interface{}); ok {
		s.Items = p.parseSchema(items)
	}

	// Parse enum
	if enum, ok := schema["enum"].([]interface{}); ok {
		s.Enum = enum
	}

	// Parse default
	if defaultVal := schema["default"]; defaultVal != nil {
		s.Default = defaultVal
	}

	// Parse example
	if example := schema["example"]; example != nil {
		s.Example = example
	}

	// Parse numeric constraints
	if minimum, ok := schema["minimum"].(float64); ok {
		s.Minimum = &minimum
	} else if minimum, ok := schema["minimum"].(int); ok {
		min := float64(minimum)
		s.Minimum = &min
	}

	if maximum, ok := schema["maximum"].(float64); ok {
		s.Maximum = &maximum
	} else if maximum, ok := schema["maximum"].(int); ok {
		max := float64(maximum)
		s.Maximum = &max
	}

	// Parse oneOf
	if oneOf, ok := schema["oneOf"].([]interface{}); ok {
		for _, item := range oneOf {
			if itemMap, ok := item.(map[string]interface{}); ok {
				s.OneOf = append(s.OneOf, p.parseSchema(itemMap))
			}
		}
	}

	// Parse anyOf
	if anyOf, ok := schema["anyOf"].([]interface{}); ok {
		for _, item := range anyOf {
			if itemMap, ok := item.(map[string]interface{}); ok {
				s.AnyOf = append(s.AnyOf, p.parseSchema(itemMap))
			}
		}
	}

	// Parse allOf
	if allOf, ok := schema["allOf"].([]interface{}); ok {
		for _, item := range allOf {
			if itemMap, ok := item.(map[string]interface{}); ok {
				s.AllOf = append(s.AllOf, p.parseSchema(itemMap))
			}
		}
	}

	// Parse additionalProperties
	if addlProps := schema["additionalProperties"]; addlProps != nil {
		s.AdditionalProperties = addlProps
	}

	// Parse string constraints
	if pattern, ok := schema["pattern"].(string); ok {
		s.Pattern = pattern
	}
	if minLength, ok := schema["minLength"].(int); ok {
		s.MinLength = &minLength
	}
	if maxLength, ok := schema["maxLength"].(int); ok {
		s.MaxLength = &maxLength
	}

	return s
}

func (p *Parser) parseComponents(components map[string]interface{}) *Components {
	c := &Components{
		Schemas:         make(map[string]*Schema),
		Responses:       make(map[string]*Response),
		Parameters:      make(map[string]*Parameter),
		SecuritySchemes: make(map[string]*SecurityScheme),
	}

	// Parse schemas
	if schemas, ok := components["schemas"].(map[string]interface{}); ok {
		for name, schema := range schemas {
			if schemaMap, ok := schema.(map[string]interface{}); ok {
				c.Schemas[name] = p.parseSchema(schemaMap)
			}
		}
	}

	// Parse responses
	if responses, ok := components["responses"].(map[string]interface{}); ok {
		for name, response := range responses {
			if responseMap, ok := response.(map[string]interface{}); ok {
				c.Responses[name] = p.parseResponse(responseMap)
			}
		}
	}

	return c
}

func (p *Parser) parseTags(tags []interface{}) []*Tag {
	var result []*Tag

	for _, tag := range tags {
		if tagMap, ok := tag.(map[string]interface{}); ok {
			t := &Tag{}
			if name, ok := tagMap["name"].(string); ok {
				t.Name = name
			}
			if description, ok := tagMap["description"].(string); ok {
				t.Description = description
			}
			result = append(result, t)
		}
	}

	return result
}
