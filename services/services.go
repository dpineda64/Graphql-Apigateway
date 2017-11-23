package services

import (
	"os"
	"context"
	"strings"
	micro "github.com/micro/go-micro/registry"
	"github.com/micro/go-micro/client"
	"github.com/micro/go-plugins/registry/kubernetes"
	"github.com/micro/go-plugins/broker/rabbitmq"
	"github.com/micro/go-plugins/transport/nats"
	"github.com/micro/go-micro/metadata"
)

type ServiceHelper struct {
	Services map[string]Service
}

type Service struct {
	Name string
	Endpoints []*micro.Endpoint
	Active bool
	ParsedEndpoints []Endpoint
}

type Endpoint struct {
	Name string
	Type string
	RequestFields []Field
	ResponseFields []Field
}

type Field struct {
	Name string
	Type string
	SubFields []Field
}

var cli = client.NewClient(
	client.Registry(kubernetes.NewRegistry()),
	client.Broker(rabbitmq.NewBroker()),
	client.Transport(nats.NewTransport()),
)

func (s *ServiceHelper) Analyze() error {
	err := s.FindServices()

	if err != nil {
		return err
	}

	return nil
}


func (s *ServiceHelper) FindServices() error {
	registry := kubernetes.NewRegistry()

	services, err := registry.ListServices()

	if err != nil {
		return err
	}

	for _, service := range services {
		if strings.Contains(service.Name, "graphql") {
			servicesInfo, err := registry.GetService(service.Name)

			if err != nil {
				return err
			}

			for _, serviceInfo := range servicesInfo {
				s.Services = map[string]Service{
					serviceInfo.Name: {Name: service.Name, Endpoints:serviceInfo.Endpoints, Active: true},
				}
			}
		}
	}

	for i, service := range s.Services {
		for _, endpoint := range service.Endpoints {
			name := strings.Split(endpoint.Name, ".")
			parsed := Endpoint{
				Name: name[1],
				RequestFields: []Field{},
				ResponseFields: []Field{},
			}
			requestFields := s.buildFields(endpoint.Request.Values, false)
			responseFields := s.buildFields(endpoint.Response.Values, false)
			parsed.RequestFields = requestFields
			parsed.ResponseFields = responseFields
			service.ParsedEndpoints = append(service.ParsedEndpoints, parsed)
		}

		s.Services[i] = service
	}
	return nil
}

func (s *ServiceHelper) buildFields(values []*micro.Value, subFields bool) []Field {
	var Fields []Field

	for _, value := range values {
		field := Field{
			Name: value.Name,
			Type: value.Type,
		}
		if len(value.Values) > 0 {
			subFields := s.buildFields(value.Values, true)
			field.SubFields = subFields
		}

		Fields = append(Fields, field)
	}

	return Fields
}

func (s *ServiceHelper) Communicate(service , endpoint string, endpointRequest map[string]interface{}) (*map[string]interface{}, error) {
	ResponseMap := make(map[string]interface {})

	req := cli.NewJsonRequest(service, endpoint,  endpointRequest)
	ctx := metadata.NewContext(context.Background(), map[string]string{
		"X-From-Id": os.Getenv("GATEWAY_FROM_ID"),
	})

	rsp := &ResponseMap

	if err := cli.Call(ctx, req, rsp); err != nil {
		return nil, err
	}

	return rsp, nil
}