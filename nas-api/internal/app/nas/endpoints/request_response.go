package endpoints

import (
	"nas/internal/app/nas"
)

type ApiResponse struct {
	Code    int32  `json:"code,omitempty"`
	Type_   string `json:"type,omitempty"`
	Message string `json:"message,omitempty"`
}

type IpAccessRequestApiRequest struct {
	IpAddress nas.IpAddress `json:"ipAddress,omitempty"`
}

type IpAccessRequestApiResponse struct {
	// Request ID
	Id  string `json:"id,omitempty"`
	Err error  `json:"error,omitempty"`
}

type FindIpAccessRequestApiRequest struct {
	RequestId string `json:"requestId,omitempty"`
}

type FindIpAccessRequestApiResponse struct {
	IpAccessRequest nas.IpAccessRequest `json:"ipAccessRequest,omitempty"`
	Err             error               `json:"error,omitempty"`
}

type FindIpAddressByIpApiRequest struct {
	IpAddress    string `json:"ipAddress,omitempty"`
	AwsAccountId string `json:"awsAccountId,omitempty"`
}

type FindIpAddressByIpApiResponse struct {
	IpAddress nas.IpAddress `json:"ipAddress,omitempty"`
	Err       error         `json:"error,omitempty"`
}

type ServiceStatusRequest struct{}

type ServiceStatusResponse struct {
	Err error `json:"err,omitempty"`
}
