package main

import (
	context "context"
	"encoding/json"
	"log"
	"net"
	"strings"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

// тут вы пишете код
// обращаю ваше внимание - в этом задании запрещены глобальные переменные
type AdminHandler struct {
}

type BizHandler struct {
}

type ACL map[string][]string

type Server struct {
	acl ACL
	AdminHandler
	BizHandler
}

func StartMyMicroservice(ctx context.Context, addr, acl string) error {
	srv := &Server{}
	if err := json.Unmarshal([]byte(acl), &srv.acl); err != nil {
		return err
	}

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(srv.streamInterceptor),
		grpc.UnaryInterceptor(srv.unaryInterceptor),
	)
	RegisterBizServer(grpcServer, srv)
	RegisterAdminServer(grpcServer, srv)

	go func() {
		err = grpcServer.Serve(lis)
		if err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		<-ctx.Done()
		grpcServer.Stop()
	}()

	return nil
}

func (s *Server) checks(ctx context.Context, fullMethod string) error {
	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return grpc.Errorf(codes.Unauthenticated, "can't get metadata")
	}

	consumer, ok := meta["consumer"]
	if !ok || len(consumer) != 1 {
		return grpc.Errorf(codes.Unauthenticated, "can't get metadata")
	}

	allowedPaths, ok := s.acl[consumer[0]]
	if !ok {
		return grpc.Errorf(codes.Unauthenticated, "NO! means NO!")
	}

	splittedMethod := strings.Split(fullMethod, "/")
	if len(splittedMethod) != 3 {
		return grpc.Errorf(codes.Unauthenticated, "NO! means NO!")
	}

	path, method := splittedMethod[1], splittedMethod[2]
	isAllowed := false
	for _, al := range allowedPaths {
		splitted := strings.Split(al, "/")
		if len(splitted) != 3 {
			continue
		}
		allowedPath, allowedMethod := splitted[1], splitted[2]
		if path != allowedPath {
			continue
		}
		if allowedMethod == "*" || method == allowedMethod {
			isAllowed = true
			break
		}
	}
	if !isAllowed {
		return grpc.Errorf(codes.Unauthenticated, "NO! means NO!")
	}
	return nil
}

func (s *Server) unaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if err := s.checks(ctx, info.FullMethod); err != nil {
		return nil, err
	}
	return handler(ctx, req)
}

func (s *Server) streamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if err := s.checks(ss.Context(), info.FullMethod); err != nil {
		return err
	}
	return handler(srv, ss)
}

func (b *BizHandler) Check(ctx context.Context, n *Nothing) (*Nothing, error) {
	return &Nothing{}, nil
}

func (b *BizHandler) Add(ctx context.Context, n *Nothing) (*Nothing, error) {
	return &Nothing{}, nil
}

func (b *BizHandler) Test(ctx context.Context, n *Nothing) (*Nothing, error) {
	return &Nothing{}, nil
}

func (s *AdminHandler) Logging(nothing *Nothing, srv Admin_LoggingServer) error {
	return nil
}

func (s *AdminHandler) Statistics(interval *StatInterval, srv Admin_StatisticsServer) error {
	return nil
}
