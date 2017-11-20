package graphql

import (
	"fmt"
	"strings"
	"github.com/graphql-go/graphql"
	"github.com/dpineda64/graphql-gateway/services"
)

type Helper struct {}

// TODO: improve go - graphql types
var typesMapping = map[string]graphql.Type{
	"string": graphql.String,
	"int32": graphql.Int,
	"bool": graphql.Boolean,
}

var Services = new(services.ServiceHelper)

func (h *Helper) BuildSchema() graphql.Schema {
	Services.FindServices()

	fields, _ := h.analyzeServices(Services.Services)

	rootQuery := graphql.ObjectConfig{
		Name: "RootQuery",
		Fields: fields,
	}

	schemaConfig := graphql.SchemaConfig{Query: graphql.NewObject(rootQuery)}

	schema, err := graphql.NewSchema(schemaConfig)

	if err != nil {
		fmt.Printf("\n SCHEMA BUILD ERROR: %s", err)
	}

	return schema
}
// Analyze Services
func (h *Helper) analyzeServices(Services map[string]services.Service) (graphql.Fields, string)  {
	var fields = graphql.Fields{}
	var name string
	for _, service := range Services {
		for _, endpoint := range service.ParsedEndpoints{
			name = strings.Replace(service.Name, ".", "_", -1)
			field, fieldName := h.BuildObject(service.Name, endpoint)
			fields[fieldName] = field
		}
	}

	return fields, name
}

// Build Graphql Object

func (h *Helper) BuildObject(serviceName string, endpoint services.Endpoint) (*graphql.Field, string){
	var fieldMap = make(graphql.Fields)
	var name = strings.Replace(endpoint.Name, ".", "_", -1)
	// Start endpoint object
	endpointType := graphql.ObjectConfig{
		Name: name,
		Fields: graphql.Fields{},
	}
	// Start endpoint field
	var endpointField = &graphql.Field{
		Description: name,
		Args:graphql.FieldConfigArgument{},
	}

	// Build request Params
	for _, requestArg := range endpoint.RequestFields{
		endpointField.Args[requestArg.Name] = &graphql.ArgumentConfig{
			Type: getFieldType(requestArg.Type),
		}
	}

	// Build endpointType fields
	for _, responseArg := range endpoint.ResponseFields{
		fieldMap[responseArg.Name] = &graphql.Field{
			Name: responseArg.Name,
			Type:getFieldType(responseArg.Type),
		}
	}

	// Set endpoint fields
	endpointType.Fields = fieldMap
	// Set endpoint field type
	endpointField.Type = graphql.NewObject(endpointType)
	// Set endpoint field resolver
	// Call to service endpoint and returns an interface equal to endpointType.Fields
	endpointField.Resolve = func(p graphql.ResolveParams) (interface{}, error) {
		msg, err := Services.Communicate(serviceName, endpoint.Name, p.Args)
		if err != nil {
			return "", err
		}
		return *msg , nil
	}

	return endpointField, name
}


// Find field type

func getFieldType(fieldType string) graphql.Type {
	j := typesMapping[fmt.Sprintf(fieldType)]
	if j == nil {
		return nil
	}
	return j
}