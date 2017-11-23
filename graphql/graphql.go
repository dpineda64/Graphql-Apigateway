package graphql

import (
	"fmt"
	"strings"
	"github.com/graphql-go/graphql"
	"github.com/dpineda64/graphql-gateway/services"
	"github.com/micro/go-micro/errors"
)

type Helper struct {
	fields []*graphql.Field
}

// TODO: improve go - graphql types
var typesMapping = map[string]graphql.Type{
	"string": graphql.String,
	"int32": graphql.Int,
	"bool": graphql.Boolean,
}

var Services = new(services.ServiceHelper)

func (h *Helper) BuildSchema() graphql.Schema {
	Services.FindServices()

	query, mutation := h.analyzeServices(Services.Services)

	rootQuery := graphql.ObjectConfig{
		Name: "RootQuery",
		Fields: query,
	}

	rootMutation := graphql.ObjectConfig{
		Name: "RootMutation",
		Fields: mutation,
	}

	schemaConfig := graphql.SchemaConfig{Query: graphql.NewObject(rootQuery), Mutation: graphql.NewObject(rootMutation)}

	schema, err := graphql.NewSchema(schemaConfig)

	if err != nil {
		fmt.Printf("Error: %e", errors.New("SERVICES_NOT_FOUND", err.Error(), 404))
	}

	return schema
}
// Analyze Services
func (h *Helper) analyzeServices(Services map[string]services.Service) (query, mutation graphql.Fields)  {
	var queries = graphql.Fields{}
	var mutations = graphql.Fields{}

	for _, service := range Services {
		for _, endpoint := range service.ParsedEndpoints{
			field, fieldName := h.BuildObject(service.Name, endpoint)
			if endpoint.Type == "mutation" {
				mutations[fieldName] = field
			}
			queries[fieldName] = field
		}
	}

	return queries, mutations
}

// Build Graphql Object

func (h *Helper) BuildObject(serviceName string, endpoint services.Endpoint) (*graphql.Field, string){
	var name = strings.Replace(endpoint.Name, ".", "_", -1)
	endpointType := graphql.ObjectConfig{}
	// Start endpoint field
	var endpointField = &graphql.Field{
		Description: name,
		Args:graphql.FieldConfigArgument{},
	}

	if len(endpoint.ResponseFields) > 1 {
		var fieldMap = make(graphql.Fields)
		endpointType.Name = name
		endpointType.Fields = graphql.Fields{}
		// Build endpointType fields
		for _, responseArg := range endpoint.ResponseFields{
			var field = &graphql.Field{
				Name: responseArg.Name,
				Type:getFieldType(responseArg),
			}

			fieldMap[responseArg.Name] = field
			h.fields = append(h.fields, field)
		}

		// Set endpoint fields
		endpointType.Fields = fieldMap
		// Set endpoint field type
		endpointField.Type = graphql.NewObject(endpointType)
	} else {
		endpointField.Type = getFieldType(endpoint.ResponseFields[0])
	}

	// Start endpoint object

	// Build request Params
	for _, requestArg := range endpoint.RequestFields{
		endpointField.Args[requestArg.Name] = &graphql.ArgumentConfig{
			Type: getFieldType(requestArg),
		}
	}

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

func getFieldType(fieldType services.Field) graphql.Type {
	field := typesMapping[fmt.Sprintf(fieldType.Type)]
	if field == nil && len(fieldType.SubFields) > 0 {
		var fieldMap = make(graphql.Fields)
		var fieldObject = graphql.ObjectConfig{
			Name: strings.Title(fieldType.Name),
		}

		for _, subField := range fieldType.SubFields {
			fieldMap[subField.Name] = &graphql.Field{
				Name: subField.Name,
				Type: getFieldType(subField),
			}

		}
		fieldObject.Fields = fieldMap

		field := graphql.NewObject(fieldObject)
		typesMapping[field.Name()] = field
		return field
	}

	if field == nil && len(fieldType.SubFields) == 0 {
		return nil
	}

	return field
}