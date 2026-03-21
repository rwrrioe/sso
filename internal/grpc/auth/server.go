package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/rwrrioe/sso/internal/usecase/auth"
	ssov2 "github.com/rwrrioe/sso_protos/v2/gen/go/sso/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type serverAPI struct {
	auth Auth
	ssov2.UnimplementedAuthServer
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
	) (userID uuid.UUID, err error)

	IsAdmin(ctx context.Context, userID uuid.UUID) (bool, error)
}

func Register(gGRPCServer *grpc.Server, auth Auth) {
	ssov2.RegisterAuthServer(gGRPCServer, &serverAPI{auth: auth})
}

func (s *serverAPI) Login(
	ctx context.Context,
	req *ssov2.LoginRequest,
) (*ssov2.LoginResponse, error) {
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	if req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	if req.GetAppId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "app_id is required")
	}

	token, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword(), int(req.GetAppId()))
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid email or password")
		}

		return nil, status.Error(codes.Internal, "failed to login")
	}

	return &ssov2.LoginResponse{Token: token}, nil
}

func (s *serverAPI) Register(
	ctx context.Context,
	req *ssov2.RegisterRequest,
) (*ssov2.RegisterResponse, error) {
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	if req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	uid, err := s.auth.RegisterNewUser(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, auth.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}

		return nil, status.Error(codes.Internal, "failed to register user")
	}

	return &ssov2.RegisterResponse{UserId: uid.String()}, nil
}

func (s *serverAPI) IsAdmin(
	ctx context.Context,
	in *ssov2.IsAdminRequest,
) (*ssov2.IsAdminResponse, error) {
	const op = "auth.serverAPI.IsAdmin"

	if in.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is empty")
	}

	uid, err := uuid.Parse(in.UserId)
	if err != nil {
		return nil, fmt.Errorf("%s:%s", op, "failed to parse uid to uuid")
	}

	isAdmin, err := s.auth.IsAdmin(ctx, uid)
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "failed to check whether the user is admin")
	}

	return &ssov2.IsAdminResponse{IsAdmin: isAdmin}, nil
}
