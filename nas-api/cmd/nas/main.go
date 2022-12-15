package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"nas/internal/app/nas"
	"nas/internal/app/nas/adapters"
	"nas/internal/app/nas/endpoints"
	"nas/internal/app/nas/transport"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

const (
	defaultHTTPPort = "8080"
	//defaultGRPCPort = "8082"
)

func main() {
	var (
		logger   log.Logger
		httpAddr = net.JoinHostPort("0.0.0.0", envString("HTTP_PORT", defaultHTTPPort))
		//grpcAddr = net.JoinHostPort("0.0.0.0", envString("GRPC_PORT", defaultGRPCPort))
	)

	//TODO: Add config mgmt https://github.com/spf13/viper
	logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = log.With(logger,
		"svc", "nas",
		"ts", log.DefaultTimestampUTC,
		"caller", log.DefaultCaller,
	)

	level.Info(logger).Log("msg", "service started")
	defer level.Info(logger).Log("msg", "service ended")

	// Create nas Service
	var s nas.Service
	{
		level.Info(logger).Log("msg", "loading AWS SDK")
		sdkConfig, err := config.LoadDefaultConfig(context.TODO())
		if err != nil {
			level.Error(logger).Log("error", err)
			//log.Fatalf("unable to load SDK config, %v", err)
		}
		db := dynamodb.NewFromConfig(sdkConfig)
		level.Info(logger).Log("msg", "initializing Requests Table DynamoDB")
		requestsRepository := adapters.NewRequestsDynamoDbRepository(db)

		level.Info(logger).Log("msg", "initializing Allow List Table DynamoDB")
		allowListRepository := adapters.NewAllowListDynamoDbRepository(db)

		s = nas.NewService(requestsRepository, allowListRepository)
	}

	// Create Go kit endpoints for the Service
	var e endpoints.Endpoints
	{
		e = endpoints.MakeServerEndpoints(s)
	}

	var h http.Handler
	{
		h = transport.NewHTTPHandler(e, logger)
	}

	errs := make(chan error)
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()

	go func() {
		level.Info(logger).Log("transport", "HTTP", "addr", httpAddr)
		errs <- http.ListenAndServe(httpAddr, h)
	}()

	level.Error(logger).Log("exit", <-errs)
}

func envString(env, fallback string) string {
	e := os.Getenv(env)
	if e == "" {
		return fallback
	}
	return e
}
