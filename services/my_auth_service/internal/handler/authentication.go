package handler

import (
	"context"
	"errors"
	"github.com/hughbliss/my_auth_service/internal/service/authn"
	authnv1 "github.com/hughbliss/my_protobuf/go/pkg/gen/authn/v1"
	"github.com/hughbliss/my_toolkit/fault"
	"github.com/hughbliss/my_toolkit/reporter"
)

func NewAuthenticationHandler(s authn.AuthenticationService) authnv1.AuthenticationServiceServer {
	return &AuthenticationHandler{
		rep:     reporter.InitReporter("AuthenticationHandler"),
		service: s,
	}
}

type AuthenticationHandler struct {
	rep     reporter.Reporter
	service authn.AuthenticationService
}

func (a AuthenticationHandler) SignUp(ctx context.Context, request *authnv1.SignUpRequest) (*authnv1.SignUpResponse, error) {
	ctx, log, end := a.rep.Start(ctx, "SignUp")
	defer end()

	tokens, err := a.service.SignUp(ctx, &authn.SignUp{
		Email:    request.Email,
		Password: request.Password,
		Name:     request.Name,
	})
	if err != nil {
		log.Error().Err(err).Send()
		f := new(fault.Fault)
		if errors.As(err, &f) {
			return nil, f.ToProto()
		}
		return nil, fault.UnhandledError.Err().ToProto()
	}

	return &authnv1.SignUpResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func (a AuthenticationHandler) SignIn(ctx context.Context, request *authnv1.SignInRequest) (*authnv1.SignInResponse, error) {
	ctx, log, end := a.rep.Start(ctx, "SignIn")
	defer end()

	tokens, err := a.service.SignIn(ctx, &authn.SignIn{
		Email:    request.Email,
		Password: request.Password,
	})
	if err != nil {
		log.Error().Err(err).Send()
		f := new(fault.Fault)
		if errors.As(err, &f) {
			return nil, f.ToProto()
		}
		return nil, fault.UnhandledError.Err().ToProto()
	}

	return &authnv1.SignInResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func (a AuthenticationHandler) RefreshToken(ctx context.Context, request *authnv1.RefreshTokenRequest) (*authnv1.RefreshTokenResponse, error) {
	ctx, log, end := a.rep.Start(ctx, "RefreshToken")
	defer end()

	tokens, err := a.service.RefreshToken(ctx, request.RefreshToken)
	if err != nil {
		log.Error().Err(err).Send()
		f := new(fault.Fault)
		if errors.As(err, &f) {
			return nil, f.ToProto()
		}
		return nil, fault.UnhandledError.Err().ToProto()
	}

	return &authnv1.RefreshTokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func (a AuthenticationHandler) Authorize(ctx context.Context, request *authnv1.AuthorizeRequest) (*authnv1.AuthorizeResponse, error) {
	ctx, log, end := a.rep.Start(ctx, "Authorize")
	defer end()

	meta, err := a.service.Authorize(ctx, request.AccessToken)
	if err != nil {
		log.Error().Err(err).Send()
		f := new(fault.Fault)
		if errors.As(err, &f) {
			return nil, f.ToProto()
		}
		return nil, fault.UnhandledError.Err().ToProto()
	}

	return &authnv1.AuthorizeResponse{
		UserId:          meta.UserId.String(),
		Email:           meta.Email,
		CurrentDomainId: meta.DomainId.String(),
		CurrentRoleId:   meta.RoleId.String(),
		Permissions:     meta.Permissions,
	}, nil
}
