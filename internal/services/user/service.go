package user

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"

	"github.com/vleukhin/gophermart/internal/storage"
)

type (
	UsersService struct {
		storage storage.Storage
		jwtKey  []byte
	}
	Claims struct {
		UserID int `json:"user_id"`
		jwt.StandardClaims
	}
	ContextKey string
)

const AuthUserID ContextKey = "userID"

var ErrUsernameTaken = errors.New("this username is already taken")

func NewUserService(storage storage.Storage, jwtKey string) *UsersService {
	return &UsersService{
		storage: storage,
		jwtKey:  []byte(jwtKey),
	}
}

func (s UsersService) Register(ctx context.Context, name, password string) (string, time.Time, error) {
	passwordHash, err := s.hashPassword(password)
	if err != nil {
		return "", time.Now(), err
	}

	created, err := s.storage.CreateUser(ctx, name, passwordHash)
	if err != nil {
		return "", time.Now(), err
	}

	if !created {
		return "", time.Now(), ErrUsernameTaken
	}

	log.Debug().Msgf("User %s created", name)

	user, err := s.storage.GetUser(ctx, name)
	if err != nil {
		return "", time.Now(), err
	}

	return s.authorize(user.ID)
}

func (s UsersService) Login(ctx context.Context, name, password string) (string, time.Time, error) {
	user, err := s.storage.GetUser(ctx, name)
	if err != nil {
		return "", time.Now(), err
	}

	if user == nil || !s.checkPasswordHash(password, user.Password) {
		return "", time.Now(), nil
	}

	return s.authorize(user.ID)
}

func (s UsersService) authorize(id int) (string, time.Time, error) {
	ttl := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: id,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: ttl.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtKey)
	if err != nil {
		return "", time.Now(), err
	}

	return tokenString, ttl, nil
}

func (s UsersService) CheckAuth(r *http.Request) (*Claims, error) {
	c, err := r.Cookie("token")
	if err != nil {
		return nil, err
	}

	tknStr := c.Value
	claims := &Claims{}

	tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
		return s.jwtKey, nil
	})
	if err != nil {
		return nil, err
	}
	if !tkn.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	return claims, nil
}

func (s UsersService) GetAuthUserID(ctx context.Context) int {
	value := ctx.Value(AuthUserID)
	if id, ok := value.(int); ok {
		return id
	}

	return 0
}

func (s UsersService) hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func (s UsersService) checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
