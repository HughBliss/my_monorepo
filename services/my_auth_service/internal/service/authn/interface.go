package authn

import (
	"context"
	"github.com/rs/xid"
)

type UserMeta struct {
	UserId      xid.ID   // UserId уникальный идентификатор пользователя.
	Email       string   // Email адрес электронной почты пользователя.
	DomainId    xid.ID   // DomainId идентификатор текущего домена пользователя.
	RoleId      xid.ID   // RoleId
	Permissions []string // Permissions доступы
}

type SignUp struct {
	Email    string // Email адрес электронной почты пользователя.
	Password string // Password пароль пользователя.
	Name     string // Name имя пользователя.
}

type SignIn struct {
	Email    string // Email адрес электронной почты пользователя.
	Password string // Password пароль пользователя.
}

type TokenPair struct {
	AccessToken  string // AccessToken JWT токен доступа.
	RefreshToken string // RefreshToken токен для обновления access токена.
}

type AuthenticationService interface {
	Authorize(ctx context.Context, accessToken string) (*UserMeta, error)
	SignUp(ctx context.Context, request *SignUp) (*TokenPair, error)
	SignIn(ctx context.Context, request *SignIn) (*TokenPair, error)
	RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error)
}
