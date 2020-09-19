package grpcDaemonServer

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"time"

	"github.com/zerops-io/zcli/src/vpn"
	"github.com/zerops-io/zcli/src/zeropsDaemonProtocol"
	"google.golang.org/grpc"
)

type Config struct {
	Socket string
}

type Handler struct {
	config Config
	vpn    *vpn.Handler
}

func New(config Config, vpn *vpn.Handler) *Handler {
	return &Handler{
		config: config,
		vpn:    vpn,
	}
}

func (h *Handler) Run(ctx context.Context) error {
	address, err := url.Parse(h.config.Socket)
	if err != nil {
		return err
	}

	err = removeUnusedServerSocket(address)
	if err != nil {
		return err
	}

	lis, err := net.Listen("unix", h.config.Socket)
	if err != nil {
		return errors.New(fmt.Sprintf("failed to listen: %v", err))
	}
	if err := os.Chmod(h.config.Socket, 0666); err != nil {
		return errors.New(fmt.Sprintf("failed to chmod: %v", err))
	}

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	zeropsDaemonProtocol.RegisterZeropsDaemonProtocolServer(grpcServer, h)

	go func() {
		grpcServer.Serve(lis)
	}()

	<-ctx.Done()

	err = lis.Close()
	if err != nil {
		return errors.New(fmt.Sprintf("failed to close listen: %v", err))
	}

	grpcServer.GracefulStop()

	return nil
}

func removeUnusedServerSocket(address *url.URL) error {
	if _, errFound := os.Stat(address.Path); errFound != nil {
		return nil
	}

	conn, err := net.DialTimeout("unix", address.Path, 1*time.Second)
	if serverIsRunning := err == nil; serverIsRunning {
		defer func() { _ = conn.Close() }()
		return fmt.Errorf("socket %s already in use", address.String())
	}

	_ = os.Remove(address.Path)
	if _, errFound := os.Stat(address.Path); errFound == nil {
		return fmt.Errorf("unused socket %s can't be deleted", address.String())
	}
	return nil
}
