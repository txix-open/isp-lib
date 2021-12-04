package grpc_test

import (
	"context"
	"net"
	"testing"

	"github.com/integration-system/isp-lib/v3/grpc"
	grpcCli "github.com/integration-system/isp-lib/v3/grpc/client"
	"github.com/integration-system/isp-lib/v3/grpc/endpoint"
	"github.com/integration-system/isp-lib/v3/log"
	"github.com/integration-system/isp-lib/v3/requestid"
	"github.com/stretchr/testify/require"
)

type reqBody struct {
	A string
	B bool
	C int32
}

type respBody struct {
	Ok bool
}

func TestGrpcBasic(t *testing.T) {
	require, srv, cli := prepareTest(t)
	reqId := requestid.Next()
	ctx := requestid.ToContext(context.Background(), reqId)
	expectedReq := reqBody{
		A: "string",
		B: true,
		C: 123,
	}
	handler := func(ctx context.Context, data grpc.AuthData, req reqBody) (*respBody, error) {
		receivedReqId := requestid.FromContext(ctx)
		require.EqualValues(reqId, receivedReqId)

		appId, err := data.ApplicationId()
		require.NoError(err)
		require.EqualValues(123, appId)

		require.EqualValues(expectedReq, req)

		return &respBody{Ok: true}, nil
	}
	logger, err := log.New()
	require.NoError(err)
	mapper := endpoint.Default(logger)
	srv.Upgrade(grpc.NewMux().Handle("endpoint", mapper.Endpoint(handler)))

	resp := respBody{}

	err = cli.Invoke("endpoint").
		ApplicationId(123).
		JsonRequestBody(expectedReq).
		ReadJsonResponse(&resp).
		Do(ctx)
	require.NoError(err)
	require.True(resp.Ok)
}

func prepareTest(t *testing.T) (*require.Assertions, *grpc.Server, *grpcCli.Client) {
	required := require.New(t)

	listener, err := net.Listen("tcp", "127.0.0.1:")
	required.NoError(err)
	srv := grpc.NewServer()
	cli, err := grpcCli.Default()
	required.NoError(err)
	t.Cleanup(func() {
		err := cli.Close()
		required.NoError(err)
		srv.Shutdown()
	})
	go func() {
		err := srv.Serve(listener)
		required.NoError(err)
	}()

	cli.Upgrade([]string{listener.Addr().String()})
	return required, srv, cli
}
