package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	domainerrors "github.com/rwrrioe/sso/internal/domain/errors"
	"github.com/rwrrioe/sso/internal/domain/models"
	"github.com/rwrrioe/sso/internal/usecase/code"
	ssov3 "github.com/rwrrioe/sso_protos/v3/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type serverAPI struct {
	auth Auth
	code Code
	ssov3.UnimplementedAuthServer
}

type Auth interface {
	Login(
		ctx context.Context,
		email string,
		password string,
		appID int,
	) (*models.TokenPair, error)

	RegisterNewUser(
		ctx context.Context,
		email string,
		password string,
	) (uuid.UUID, error)

	RegenerateToken(
		ctx context.Context,
		refreshToken string,
	) (*models.TokenPair, error)

	GenerateResetToken(
		ctx context.Context,
		email string,
	) (string, error)

	CreateNewPassword(
		ctx context.Context,
		resetToken, password, email string) error

	Logout(ctx context.Context, refreshToken string) error
	IsAdmin(ctx context.Context, userID uuid.UUID) (bool, error)
}

type Code interface {
	VerifyCode(ctx context.Context, code string) error
	SendCode(
		ctx context.Context,
		email string,
		opts *code.Options,
	) error
}

// auth

func Register(
	gGRPCServer *grpc.Server,
	auth Auth,
	code Code) {
	ssov3.RegisterAuthServer(
		gGRPCServer, &serverAPI{
			auth: auth,
			code: code,
		})
}

func (s *serverAPI) Login(
	ctx context.Context,
	req *ssov3.LoginRequest,
) (*ssov3.LoginResponse, error) {
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	if req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	if req.GetAppId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "app_id is required")
	}

	tokenPair, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword(), int(req.GetAppId()))
	if err != nil {
		if errors.Is(err, domainerrors.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid email or password")
		}

		return nil, status.Error(codes.Internal, "failed to login")
	}

	return &ssov3.LoginResponse{
		AccToken: tokenPair.AccessToken,
		RefToken: tokenPair.RefreshToken,
	}, nil
}

func (s *serverAPI) Register(
	ctx context.Context,
	req *ssov3.RegisterRequest,
) (*ssov3.RegisterResponse, error) {
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	if req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	uid, err := s.auth.RegisterNewUser(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, domainerrors.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}

		return nil, status.Error(codes.Internal, "failed to register user")
	}

	return &ssov3.RegisterResponse{UserId: uid.String()}, nil
}

func (s *serverAPI) RegenerateToken(
	ctx context.Context,
	req *ssov3.RegenerateTokenRequest,
) (*ssov3.RegenerateTokenResponse, error) {

	if req.RefToken == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh token is required")
	}

	tokenPair, err := s.auth.RegenerateToken(ctx, req.RefToken)
	if err != nil {
		if errors.Is(err, domainerrors.ErrInvalidToken) {
			return nil, status.Error(codes.InvalidArgument, "invalid refresh token")
		}

		if errors.Is(err, domainerrors.ErrAppNotFound) {
			return nil, status.Error(codes.NotFound, "app not found")
		}

		if errors.Is(err, domainerrors.ErrTokenExpired) {
			return nil, status.Error(codes.InvalidArgument, "token is expired")
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &ssov3.RegenerateTokenResponse{
		AccToken: tokenPair.AccessToken,
		RefToken: tokenPair.RefreshToken,
	}, nil
}

// reset password flow

func (s *serverAPI) SendResetCode(
	ctx context.Context,
	req *ssov3.SendCodeRequest,
) (*ssov3.SendCodeResponse, error) {
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	if err := s.code.SendCode(ctx, req.Email, &code.Options{
		CodeType: code.TypeResetCode}); err != nil {
		if errors.Is(err, domainerrors.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}

		return nil, status.Error(codes.Internal, "failed to send the code")
	}

	return &ssov3.SendCodeResponse{}, nil
}

func (s *serverAPI) VerifyResetCode(
	ctx context.Context,
	req *ssov3.VerifyCodeRequest,
) (*ssov3.VerifyCodeResponse, error) {
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	if req.Code == "" {
		return nil, status.Error(codes.InvalidArgument, "reset code is required")
	}

	if err := s.code.VerifyCode(ctx, req.Code); err != nil {
		if errors.Is(err, domainerrors.ErrInvalidCode) {
			return nil, status.Error(codes.InvalidArgument, "invalid reset code")
		}

		if errors.Is(err, domainerrors.ErrCodeAlreadyUsed) {
			return nil, status.Error(codes.InvalidArgument, "reset code already used")
		}

		if errors.Is(err, domainerrors.ErrCodeExpired) {
			return nil, status.Error(codes.InvalidArgument, "reset code expired")
		}

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	resetToken, err := s.auth.GenerateResetToken(ctx, req.Email)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate reset token")
	}

	return &ssov3.VerifyCodeResponse{
		ResetToken: resetToken,
	}, nil
}

func (s *serverAPI) CreateNewPassword(
	ctx context.Context,
	req *ssov3.CreateNewPasswordRequest,
) (*ssov3.CreateNewPasswordResponse, error) {

	if err := s.auth.CreateNewPassword(ctx, req.ResetToken, req.Password, req.Email); err != nil {
		if errors.Is(err, domainerrors.ErrInvalidResetToken) {
			return nil, status.Error(codes.InvalidArgument, "invalid reset token")
		}

		if errors.Is(err, domainerrors.ErrResetTokenExpired) {
			return nil, status.Error(codes.InvalidArgument, "reset token expired ")
		}

		if errors.Is(err, domainerrors.ErrResetTokenAlreadyUsed) {
			return nil, status.Error(codes.InvalidArgument, "reset token already used")
		}

		return nil, status.Error(codes.Internal, "failed to create password")
	}

	return &ssov3.CreateNewPasswordResponse{}, nil
}

// authz

func (s *serverAPI) IsAdmin(
	ctx context.Context,
	in *ssov3.IsAdminRequest,
) (*ssov3.IsAdminResponse, error) {
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
		if errors.Is(err, domainerrors.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "failed to check whether the user is admin")
	}

	return &ssov3.IsAdminResponse{IsAdmin: isAdmin}, nil
}
