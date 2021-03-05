package mock

import (
	"context"
	"errors"
	"net"
	"strconv"

	"github.com/cybozu-go/moco/agentrpc"
	"github.com/cybozu-go/moco/test_utils"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

type MockAgentServer struct {
	agentrpc.UnimplementedBackupBinlogServiceServer
	agentrpc.UnimplementedCloneServiceServer
	healthpb.HealthServer

	ReturnErr        bool
	LastCloneRequest *agentrpc.CloneRequest
}

var M = &MockAgentServer{}

type cloneRequestParam struct {
	Token     string
	DonorHost string
	DonorPort int32
	External  bool
}

func Start(ctx context.Context) error {
	s, err := net.Listen("tcp", test_utils.Host+":"+strconv.Itoa(test_utils.AgentPort))
	if err != nil {
		return err
	}
	grpcServer := grpc.NewServer()
	healthpb.RegisterHealthServer(grpcServer, M)
	agentrpc.RegisterCloneServiceServer(grpcServer, M)
	agentrpc.RegisterBackupBinlogServiceServer(grpcServer, M)

	go func() {
		err := grpcServer.Serve(s)
		if err != nil {
			panic(err)
		}
	}()
	go func(ctx context.Context) {
		<-ctx.Done()
		grpcServer.Stop()
	}(ctx)

	return nil
}

func (m *MockAgentServer) Clone(ctx context.Context, req *agentrpc.CloneRequest) (*agentrpc.CloneResponse, error) {
	m.LastCloneRequest = req
	if M.ReturnErr {
		return &agentrpc.CloneResponse{}, errors.New("clone api dummy error")
	} else {
		return &agentrpc.CloneResponse{}, nil
	}
}

func CompareWithLastCloneRequest(req *agentrpc.CloneRequest) string {
	base := &cloneRequestParam{
		Token:     M.LastCloneRequest.Token,
		DonorHost: M.LastCloneRequest.DonorHost,
		DonorPort: M.LastCloneRequest.DonorPort,
		External:  M.LastCloneRequest.External,
	}
	target := &cloneRequestParam{
		Token:     req.Token,
		DonorHost: req.DonorHost,
		DonorPort: req.DonorPort,
		External:  req.External,
	}

	return cmp.Diff(base, target)
}
