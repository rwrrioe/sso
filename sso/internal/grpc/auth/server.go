package auth

import (
	"context"

	ssov1 "github.com/GolangLessons/protos/gen/go/sso"
	"google.golang.org/grpc"
)

type serverAPI struct {
	ssov1.UnimplementedAuthServer
}

type Auth interface {
	Login(
		ctx context.Context,
		email string,
		password string,
		appID int,
	) (token string, err error)

	RegisterNewUser(
		ctx context.Context,
		email string,
		passwords string,
	) (userID int64, err error)
}

func Register(gGRPCServer *grpc.Server) {
	ssov1.RegisterAuthServer(gGRPCServer, &serverAPI{})
}

func (s *serverAPI) Login(
	ctx context.Context,
	req *ssov1.LoginRequest,
) (*ssov1.LoginResponse, error) {
	// if req.Email == "" {
	// 	return nil, status.Error(codes.InvalidArgument, "email is required")
	// }

	// if req.Password == "" {
	// 	return nil, status.Error(codes.InvalidArgument, "password is required")
	// }

	// if req.GetAppId() == 0 {
	// 	return nil, status.Error(codes.InvalidArgument, "app_id is required")
	// }

	// token, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword(), int(req.GetAppId()))
	// if err != nil {
	// 	if errors.Is(err, auth.ErrInvlidCredentials) {
	// 		return nil, status.Error(codes.InvalidArgument, "invalid email or password")
	// 	}

	// 	return nil, status.Error(codes.Internal, "failed to login")
	// }

	return nil, nil
}

func (s *serverAPI) Register(
	ctx context.Context,
	req *ssov1.RegisterRequest,
) (*ssov1.RegisterResponse, error) {
	// if req.Email == "" {
	// 	return nil, status.Error(codes.InvalidArgument, "email is required")
	// }

	// if req.Password == "" {
	// 	return nil, status.Error(codes.InvalidArgument, "password is required")
	// }

	// uid, err := s.auth.RegisterNewUser(ctx, req.GetEmail(), req.GetPassword())
	// if err != nil {
	// 	if errors.Is(err, storage.ErrUserExists) {
	// 		return nil, status.Error(codes.AlreadyExists, "user already exists")
	// 	}

	// 	return nil, status.Error(codes.Internal, "failed to register user")
	// }

	return nil, nil
}
