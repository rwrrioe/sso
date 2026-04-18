package server

import (
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	EmptyEmail, _ = status.New(codes.InvalidArgument, "email is required").
			WithDetails(&errdetails.ErrorInfo{
			Reason: "EMPTY_EMAIL",
			Domain: "sso",
		})

	EmptyPassword, _ = status.New(codes.InvalidArgument, "password is required").
				WithDetails(&errdetails.ErrorInfo{
			Reason: "EMPTY_PASSWORD",
			Domain: "sso",
		})

	EmptyAppID, _ = status.New(codes.InvalidArgument, "app_id is required").
			WithDetails(&errdetails.ErrorInfo{
			Reason: "EMPTY_APP_ID",
			Domain: "sso",
		})

	EmptyRefreshToken, _ = status.New(codes.InvalidArgument, "refresh token is required").
				WithDetails(&errdetails.ErrorInfo{
			Reason: "EMPTY_REFRESH_TOKEN",
			Domain: "sso",
		})

	EmptyCode, _ = status.New(codes.InvalidArgument, "reset code is required").
			WithDetails(&errdetails.ErrorInfo{
			Reason: "EMPTY_CODE",
			Domain: "sso",
		})

	EmptyUserID, _ = status.New(codes.InvalidArgument, "user_id is empty").
			WithDetails(&errdetails.ErrorInfo{
			Reason: "EMPTY_USER_ID",
			Domain: "sso",
		})

	InvalidCredentials, _ = status.New(codes.InvalidArgument, "invalid email or password").
				WithDetails(&errdetails.ErrorInfo{
			Reason: "INVALID_CREDENTIALS",
			Domain: "sso",
		})

	InvalidRefreshToken, _ = status.New(codes.InvalidArgument, "invalid refresh token").
				WithDetails(&errdetails.ErrorInfo{
			Reason: "INVALID_REFRESH_TOKEN",
			Domain: "sso",
		})

	InvalidResetToken, _ = status.New(codes.InvalidArgument, "invalid reset token").
				WithDetails(&errdetails.ErrorInfo{
			Reason: "INVALID_RESET_TOKEN",
			Domain: "sso",
		})

	InvalidCode, _ = status.New(codes.InvalidArgument, "invalid reset code").
			WithDetails(&errdetails.ErrorInfo{
			Reason: "INVALID_CODE",
			Domain: "sso",
		})

	TokenExpired, _ = status.New(codes.InvalidArgument, "token is expired").
			WithDetails(&errdetails.ErrorInfo{
			Reason: "TOKEN_EXPIRED",
			Domain: "sso",
		})

	CodeAlreadyUsed, _ = status.New(codes.InvalidArgument, "reset code already used").
				WithDetails(&errdetails.ErrorInfo{
			Reason: "CODE_ALREADY_USED",
			Domain: "sso",
		})

	CodeExpired, _ = status.New(codes.InvalidArgument, "reset code expired").
			WithDetails(&errdetails.ErrorInfo{
			Reason: "CODE_EXPIRED",
			Domain: "sso",
		})

	ResetTokenAlreadyUsed, _ = status.New(codes.InvalidArgument, "reset token already used").
					WithDetails(&errdetails.ErrorInfo{
			Reason: "RESET_TOKEN_ALREADY_USED",
			Domain: "sso",
		})

	UserAlreadyExists, _ = status.New(codes.AlreadyExists, "user already exists").
				WithDetails(&errdetails.ErrorInfo{
			Reason: "USER_ALREADY_EXISTS",
			Domain: "sso",
		})

	UserNotFound, _ = status.New(codes.NotFound, "user not found").
			WithDetails(&errdetails.ErrorInfo{
			Reason: "USER_NOT_FOUND",
			Domain: "sso",
		})

	AppNotFound, _ = status.New(codes.NotFound, "app not found").
			WithDetails(&errdetails.ErrorInfo{
			Reason: "APP_NOT_FOUND",
			Domain: "sso",
		})

	LoginFailed, _ = status.New(codes.Internal, "failed to login").
			WithDetails(&errdetails.ErrorInfo{
			Reason: "LOGIN_FAILED",
			Domain: "sso",
		})

	RegisterFailed, _ = status.New(codes.Internal, "failed to register user").
				WithDetails(&errdetails.ErrorInfo{
			Reason: "REGISTER_FAILED",
			Domain: "sso",
		})

	LogoutFailed, _ = status.New(codes.Internal, "failed to logout").
			WithDetails(&errdetails.ErrorInfo{
			Reason: "LOGOUT_FAILED",
			Domain: "sso",
		})

	SendCodeFailed, _ = status.New(codes.Internal, "failed to send the code").
				WithDetails(&errdetails.ErrorInfo{
			Reason: "SEND_CODE_FAILED",
			Domain: "sso",
		})

	GenerateResetTokenFailed, _ = status.New(codes.Internal, "failed to generate reset token").
					WithDetails(&errdetails.ErrorInfo{
			Reason: "GENERATE_RESET_TOKEN_FAILED",
			Domain: "sso",
		})

	CreatePasswordFailed, _ = status.New(codes.Internal, "failed to create password").
				WithDetails(&errdetails.ErrorInfo{
			Reason: "CREATE_PASSWORD_FAILED",
			Domain: "sso",
		})

	IsAdminFailed, _ = status.New(codes.Internal, "failed to check whether the user is admin").
				WithDetails(&errdetails.ErrorInfo{
			Reason: "IS_ADMIN_FAILED",
			Domain: "sso",
		})
)
