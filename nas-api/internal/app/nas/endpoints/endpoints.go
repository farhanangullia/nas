package endpoints

import (
	"context"
	"github.com/go-kit/kit/endpoint"
	"nas/internal/app/nas"
	"log"
)

// Endpoints holds all Go kit endpoints for nas.
type Endpoints struct {
	AddIpAccessRequest             endpoint.Endpoint
	FindIpAccessRequestByRequestId endpoint.Endpoint
	ServiceStatusRequest           endpoint.Endpoint
	FindIpAddressByIp              endpoint.Endpoint
}

// // MakeServerEndpoints initializes all Go kit endpoints for nas.
func MakeServerEndpoints(s nas.Service) Endpoints {
	return Endpoints{
		AddIpAccessRequest:             makeAddIpAccessRequestEndpoint(s),
		FindIpAccessRequestByRequestId: makeFindIpAccessRequestByRequestId(s),
		ServiceStatusRequest:           makeServiceStatusRequestEndpoint(s),
		FindIpAddressByIp:              makeFindIpAddressByIp(s),
	}
}

func makeServiceStatusRequestEndpoint(s nas.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		_ = request.(ServiceStatusRequest)
		err := s.ServiceStatus(ctx)
		if err != nil {
			return ServiceStatusResponse{Err: err}, nil
		}
		return ServiceStatusResponse{}, nil
	}
}

func makeAddIpAccessRequestEndpoint(s nas.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(IpAccessRequestApiRequest) // type assertion
		log.Printf("endpoints.go makeAddIpAccessRequestEndpoint: %+v\n", req)

		id, err := s.AddIpAccessRequest(ctx, req.IpAddress)
		if err != nil {
			return IpAccessRequestApiResponse{Err: err}, nil
		}
		return IpAccessRequestApiResponse{Id: id}, nil
	}
}

func makeFindIpAccessRequestByRequestId(s nas.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(FindIpAccessRequestApiRequest) // type assertion
		log.Printf("endpoints.go makeFindIpAccessRequestByRequestId: %+v\n", req)

		ipAccessRequest, err := s.FindIpAccessRequestByRequestId(ctx, req.RequestId)

		if err != nil {
			return FindIpAccessRequestApiResponse{Err: err}, nil
		}
		return FindIpAccessRequestApiResponse{IpAccessRequest: *ipAccessRequest}, nil
	}
}

func makeFindIpAddressByIp(s nas.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(FindIpAddressByIpApiRequest) // type assertion
		log.Printf("endpoints.go makeFindIpAddressByIp: %+v\n", req)

		ipAddress, err := s.FindIpAddressByIp(ctx, req.IpAddress, req.AwsAccountId)

		if err != nil {
			return FindIpAddressByIpApiResponse{Err: err}, nil
		}
		return FindIpAddressByIpApiResponse{IpAddress: *ipAddress}, nil
	}
}

// errors
func (r ServiceStatusResponse) Error() error          { return r.Err }
func (r IpAccessRequestApiResponse) Error() error     { return r.Err }
func (r FindIpAccessRequestApiResponse) Error() error { return r.Err }
func (r FindIpAddressByIpApiResponse) Error() error   { return r.Err }
