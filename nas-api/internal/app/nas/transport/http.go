package transport

import (
	"context"
	"encoding/json"
	kittransport "github.com/go-kit/kit/transport"
	kithttp "github.com/go-kit/kit/transport/http"
	kitlog "github.com/go-kit/log"
	"github.com/gorilla/mux"
	"nas/internal/app/nas"
	"nas/internal/app/nas/endpoints"
	"net/http"
)

func NewHTTPHandler(ep endpoints.Endpoints, logger kitlog.Logger) http.Handler {
	r := mux.NewRouter()

	options := []kithttp.ServerOption{
		kithttp.ServerErrorHandler(kittransport.NewLogErrorHandler(logger)),
		kithttp.ServerErrorEncoder(encodeError),
	}

	// Configure routes with Gorilla Mux package
	r.Methods("GET").Name("HealthCheck").Path("/nas/api/v2/healthz").Handler(kithttp.NewServer(
		ep.ServiceStatusRequest,
		decodeServiceStatusRequest,
		encodeResponse,
		options...,
	))

	r.Methods("POST").Name("AddIpAccessRequest").Path("/nas/api/v2/request").Handler(kithttp.NewServer(
		ep.AddIpAccessRequest,
		decodeAddIpAccessRequest,
		encodeResponse,
		options...,
	))

	r.Methods("GET").Name("FindIpAccessRequestByRequestId").Path("/nas/api/v2/request/findByRequestId").Handler(kithttp.NewServer(
		ep.FindIpAccessRequestByRequestId,
		decodeFindIpAccessRequestByRequestId,
		encodeResponse,
		options...,
	))

	r.Methods("GET").Name("FindIpAddressByIp").Path("/nas/api/v2/allowlist/findIpAddressByIp").Handler(kithttp.NewServer(
		ep.FindIpAddressByIp,
		decodeFindIpAddressByIp,
		encodeResponse,
		options...,
	))

	return r
}

func decodeServiceStatusRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	var req endpoints.ServiceStatusRequest
	if r.ContentLength == 0 {
		return req, nil
	}

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func decodeAddIpAccessRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	var req endpoints.IpAccessRequestApiRequest

	if e := json.NewDecoder(r.Body).Decode(&req); e != nil {
		return nil, e
	}
	return req, nil
}

func decodeFindIpAccessRequestByRequestId(_ context.Context, r *http.Request) (request interface{}, err error) {
	var req endpoints.FindIpAccessRequestApiRequest

	if e := json.NewDecoder(r.Body).Decode(&req); e != nil {
		return nil, e
	}
	return req, nil
}

func decodeFindIpAddressByIp(_ context.Context, r *http.Request) (request interface{}, err error) {
	var req endpoints.FindIpAddressByIpApiRequest

	if e := json.NewDecoder(r.Body).Decode(&req); e != nil {
		return nil, e
	}
	return req, nil
}

//func decodeHTTPGetRequest(_ context.Context, r *http.Request) (interface{}, error) {
//	var req endpoints.GetRequest
//	if r.ContentLength == 0 {
//		logger.Log("Get request with no body")
//		return req, nil
//	}
//	err := json.NewDecoder(r.Body).Decode(&req)
//	if err != nil {
//		return nil, err
//	}
//	return req, nil
//}

// Errorer is implemented by all concrete response types that may contain
// errors. It allows us to change the HTTP response code without needing to
// trigger an endpoint (transport-level) error.
type Errorer interface {
	Error() error
}

// encodeResponse is the common method to encode all response types to the
// client. I chose to do it this way because, since we're using JSON, there's no
// reason to provide anything more specific. It's certainly possible to
// specialize on a per-response (per-method) basis.
func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(Errorer); ok && e.Error() != nil { // Not a Go kit transport error, but a business-logic error.
		// Provide those as HTTP errors.
		encodeError(ctx, e.Error(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	switch err {
	case nas.ErrRequestNotFound:
		w.WriteHeader(http.StatusNotFound)
	case nas.ErrIpNotFound:
		w.WriteHeader(http.StatusNotFound)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}
