package authn

import (
	"context"
	zfg "github.com/chaindead/zerocfg"
	"github.com/hughbliss/my_database/pkg/gen/dbauth"
	entUser "github.com/hughbliss/my_database/pkg/gen/dbauth/user"
	"github.com/hughbliss/my_toolkit/fault"
	"github.com/hughbliss/my_toolkit/reporter"
	"github.com/rs/xid"
	"golang.org/x/crypto/bcrypt"

	"time"
)

import "github.com/golang-jwt/jwt/v5"

func New(db *dbauth.Client) AuthenticationService {
	return &impl{
		db:  db,
		rep: reporter.InitReporter("AuthenticationService"),
	}
}

type impl struct {
	rep reporter.Reporter
	db  *dbauth.Client
}

const (
	WrongEmailOrPassword fault.Code = "WrongEmailOrPassword" // WrongEmailOrPassword: "не верные данные для входа"
	UserDBErr            fault.Code = "UserDBErr"            // UserDBErr: "ошибка при обращении в базу данных"
	UserAlreadyExists    fault.Code = "UserAlreadyExists"    // UserAlreadyExists: "пользователь с таким email уже существует"
	InvalidToken         fault.Code = "InvalidToken"         // InvalidToken: "не валидный токен"
	UserNotFound         fault.Code = "UserNotFound"         // UserNotFound: "пользователь не найден"
	DomainNotFound       fault.Code = "DomainNotFound"       // DomainNotFound: "Домен по умолчанию не найден"
)

var (
	cfgGroup             = zfg.NewGroup("auth")
	secret               = zfg.Str("secret", "", "AUTH_SECRET", zfg.Required(), zfg.Secret(), zfg.Group(cfgGroup))
	refreshTokenLifetime = zfg.Dur("refresh_token_lifetime", 14*24*time.Hour, "AUTH_REFRESHTOKENLIFETIME", zfg.Group(cfgGroup))
	accessTokenLifetime  = zfg.Dur("access_token_lifetime", 15*time.Minute, "AUTH_ACCESSTOKENLIFETIME", zfg.Group(cfgGroup))
)

func (i impl) SignUp(ctx context.Context, request *SignUp) (*TokenPair, error) {
	ctx, log, end := i.rep.Start(ctx, "SignUp")
	defer end()

	// Проверяем существование пользователя
	exists, err := i.db.User.Query().Where(entUser.Email(request.Email)).Exist(ctx)
	if err != nil {
		log.Error().Err(err).Stack().Msg("failed to check user existence")
		return nil, UserDBErr.Err()
	}
	if exists {
		return nil, UserAlreadyExists.Err()
	}

	// Хэшируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Error().Err(err).Stack().Msg("failed to hash password")
		return nil, err
	}

	domain, err := i.db.Domain.Query().First(ctx)
	if err != nil {
		log.Error().Err(err).Stack().Msg("failed to query db domain")
		return nil, DomainNotFound.Err()
	}

	// Создаем пользователя
	user, err := i.db.User.Create().
		SetEmail(request.Email).
		SetName(request.Name).
		SetPasswordHash(string(hashedPassword)).
		SetCurrentDomain(domain).
		Save(ctx)
	if err != nil {
		log.Error().Err(err).Stack().Msg("failed to create user")
		return nil, UserDBErr.Err()
	}

	return i.generateTokenPair(ctx, user)
}

func (i impl) RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error) {
	ctx, log, end := i.rep.Start(ctx, "RefreshToken")
	defer end()

	// Парсим refresh token
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fault.UnhandledError.Err()
		}
		return []byte(*secret), nil
	})
	if err != nil {
		return nil, InvalidToken.Err()
	}

	// Проверяем claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, InvalidToken.Err()
	}

	// Проверяем тип токена
	if tokenType, ok := claims["token_type"].(string); !ok || tokenType != "refresh" {
		return nil, InvalidToken.Err()
	}

	// Получаем пользователя
	userId, ok := claims["user_id"].(string)
	if !ok {
		return nil, InvalidToken.Err()
	}

	userXID, err := xid.FromString(userId)
	if err != nil {
		return nil, InvalidToken.Err()
	}

	user, err := i.db.User.Get(ctx, userXID)
	if err != nil {
		if dbauth.IsNotFound(err) {
			return nil, UserNotFound.Err()
		}
		log.Error().Err(err).Stack().Msg("failed to get user")
		return nil, UserDBErr.Err()
	}

	return i.generateTokenPair(ctx, user)
}

func (i impl) Authorize(ctx context.Context, accessToken string) (*UserMeta, error) {
	ctx, log, end := i.rep.Start(ctx, "Authorize")
	defer end()

	// Парсим access token
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, InvalidToken.Err()
		}
		return []byte(*secret), nil
	})
	if err != nil {
		return nil, InvalidToken.Err()
	}

	// Проверяем claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, InvalidToken.Err()
	}

	// Проверяем тип токена
	if tokenType, ok := claims["token_type"].(string); !ok || tokenType != "access" {
		return nil, InvalidToken.Err()
	}

	// Получаем пользователя с доменом и ролью
	userId := claims["user_id"].(string)

	userXID, err := xid.FromString(userId)
	if err != nil {
		return nil, InvalidToken.Err()
	}

	user, err := i.db.User.Query().
		WithUserDomain(func(query *dbauth.UserDomainQuery) {
			query.WithRole()
		}).
		Where(entUser.ID(userXID)).
		Only(ctx)
	if err != nil {
		if dbauth.IsNotFound(err) {
			return nil, UserNotFound.Err()
		}
		log.Error().Err(err).Stack().Msg("failed to get user")
		return nil, UserDBErr.Err()
	}

	// Собираем permissions из роли
	var permissions []string
	var roleID xid.ID
	for _, domain := range user.Edges.UserDomain {
		if domain.DomainID != user.CurrentDomainID {
			continue
		}
		roleID = domain.RoleID
		permissions = domain.Edges.Role.Permissions
	}

	return &UserMeta{
		UserId:      user.ID,
		Email:       user.Email,
		DomainId:    user.CurrentDomainID,
		RoleId:      roleID,
		Permissions: permissions,
	}, nil
}

func (i impl) SignIn(ctx context.Context, request *SignIn) (*TokenPair, error) {
	ctx, log, end := i.rep.Start(ctx, "SignIn")
	defer end()

	user, err := i.db.User.Query().WithUserDomain(func(query *dbauth.UserDomainQuery) {
		query.WithDomain().WithRole()
	}).Where(entUser.Email(request.Email)).Only(ctx)
	if err != nil {
		if dbauth.IsNotFound(err) {
			return nil, WrongEmailOrPassword.Err()
		}
		log.Error().Err(err).Stack().Msg("failed to query entUser")
		return nil, UserDBErr.Err()
	}

	if err := i.verifyPassword(user.PasswordHash, request.Password); err != nil {
		return nil, WrongEmailOrPassword.Err()
	}

	return i.generateTokenPair(ctx, user)

}

func (i impl) verifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
func (i impl) generateTokenPair(ctx context.Context, user *dbauth.User) (*TokenPair, error) {
	_, log, end := i.rep.Start(ctx, "generateTokenPair")
	defer end()

	// Формируем основные клеймы для токенов
	now := time.Now()
	accessClaims := jwt.MapClaims{
		"user_id":    user.ID.String(),
		"email":      user.Email,
		"exp":        now.Add(*accessTokenLifetime).Unix(), // access token живет 15 минут
		"iat":        now.Unix(),
		"token_type": "access",
	}

	refreshClaims := jwt.MapClaims{
		"user_id":    user.ID.String(),
		"exp":        now.Add(*refreshTokenLifetime).Unix(),
		"iat":        now.Unix(),
		"token_type": "refresh",
	}

	// Создаем токены
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)

	// Подписываем токены
	accessTokenString, err := accessToken.SignedString([]byte(*secret))
	if err != nil {
		log.Error().Err(err).Stack().Msg("failed to sign access token")
		return nil, err
	}

	refreshTokenString, err := refreshToken.SignedString([]byte(*secret))
	if err != nil {
		log.Error().Err(err).Stack().Msg("failed to sign refresh token")
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
	}, nil
}
