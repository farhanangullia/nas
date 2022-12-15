package nas

import (
	"context"
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"nas/internal/app/nas/common"
	"time"
)

type Status string

const (
	Pending   Status = "Pending"
	Completed Status = "Completed"
	Failed    Status = "Failed"
)

type IpAccessRequest struct {

	// Request ID
	Id string `json:"id,omitempty"`

	IpAddress *IpAddress `json:"ipAddress,omitempty"`

	Status string `json:"status,omitempty"`

	DateRequested time.Time `json:"dateRequested,omitempty"`
}

type IpAddress struct {
	Ip string `json:"ip,omitempty" validate:"required,cidrv4"`

	Type_ string `json:"type,omitempty" validate:"required,oneof=persistent time-bounded"`

	AwsAccountId string `json:"awsAccountId,omitempty" validate:"required,number"`

	// expiry is in Epochs
	Expiry int `json:"expiry,omitempty" validate:"required_if=Type_ time-bounded,number"`

	Requestor string `json:"requestor,omitempty" validate:"required"`

	Approver string `json:"approver,omitempty" validate:"required"`
}

var (
	ErrIpNotFound          = errors.New("IP address not found")
	ErrRequestNotFound     = errors.New("Request not found")
	ErrInvalidIp           = errors.New("Invalid ip address")
	ErrInvalidRequestType  = errors.New("Invalid ip request type")
	ErrInvalidAwsAccountId = errors.New("Invalid AWS Account id")
	ErrInvalidInput        = errors.New("Invalid inputs provided")
)

var (
	validate = validator.New()
)

type RequestsRepository interface {
	CreateIpAccessRequest(ctx context.Context, request IpAccessRequest) error
	RetrieveIpAccessRequest(ctx context.Context, requestId string) (*IpAccessRequest, error)
}

type AllowListRepository interface {
	RetrieveIpAddress(ctx context.Context, ipAddress string, awsAccountId string) (*IpAddress, error)
}

type Service interface {
	AddIpAccessRequest(ctx context.Context, ipAddress IpAddress) (string, error)
	FindIpAccessRequestByRequestId(ctx context.Context, requestId string) (*IpAccessRequest, error)
	FindIpAddressByIp(ctx context.Context, ipAddress string, awsAccountId string) (*IpAddress, error)
	ServiceStatus(ctx context.Context) error
}

type service struct {
	requestsRepository  RequestsRepository
	allowListRepository AllowListRepository
}

func NewService(requestsRepository RequestsRepository, allowListRepository AllowListRepository) *service {
	return &service{
		requestsRepository:  requestsRepository,
		allowListRepository: allowListRepository,
	}
}

func (s *service) AddIpAccessRequest(ctx context.Context, ipAddress IpAddress) (string, error) {
	requestDate := time.Now()
	id := uuid.New().String()

	err := common.ValidateStruct(*validate, ipAddress)
	if err != nil {
		return "", err
	}

	// create request
	ipAccessRequestObj := IpAccessRequest{
		Id:            id,
		Status:        string(Pending),
		DateRequested: requestDate,
		IpAddress:     &ipAddress,
	}

	err = s.requestsRepository.CreateIpAccessRequest(ctx, ipAccessRequestObj)

	if err != nil {
		return "", err
	}

	return id, nil
}

func (s *service) FindIpAccessRequestByRequestId(ctx context.Context, requestId string) (*IpAccessRequest, error) {

	err := common.ValidateVar(*validate, requestId, "required")
	if err != nil {
		return nil, err
	}

	request, err := s.requestsRepository.RetrieveIpAccessRequest(ctx, requestId)

	if err != nil {
		return request, err
	}

	if *request == (IpAccessRequest{}) {
		return request, ErrRequestNotFound
	}

	return request, nil
}

func (s *service) FindIpAddressByIp(ctx context.Context, ipAddress string, awsAccountId string) (*IpAddress, error) {

	err := common.ValidateVar(*validate, ipAddress, "required,cidrv4")

	if err != nil {
		return nil, err
	}

	err = common.ValidateVar(*validate, awsAccountId, "required,number")

	if err != nil {
		return nil, err
	}

	ip, err := s.allowListRepository.RetrieveIpAddress(ctx, ipAddress, awsAccountId)

	if err != nil {
		return ip, err
	}

	// If IP not found
	if *ip == (IpAddress{}) {
		return ip, ErrIpNotFound
	}

	return ip, nil
}

func (s *service) ServiceStatus(ctx context.Context) error {
	return nil
}
