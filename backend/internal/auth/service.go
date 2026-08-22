package auth

import (
	"context"
	"errors"
	"log/slog"
	"net/mail"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"moderndvwa/backend/internal/users"
)

const (
	minPasswordLength = 10
	defaultRole       = "user"
)

var (
	ErrInvalidEmail       = errors.New("invalid email address")
	ErrWeakPassword       = errors.New("password does not meet the length policy")
	ErrEmailTaken         = users.ErrEmailTaken
	ErrInvalidCredentials = errors.New("invalid email or password")
)

type UserStore interface {
	Create(ctx context.Context, email, passwordHash, role string) (*users.User, error)
	ByEmail(ctx context.Context, email string) (*users.User, error)
	ByID(ctx context.Context, id string) (*users.User, error)
}

type Service struct {
	store     UserStore
	decoyHash []byte
	log       *slog.Logger
}

func NewService(store UserStore, log *slog.Logger) *Service {
	decoy, err := bcrypt.GenerateFromPassword([]byte("decoy-password-for-unknown-users"), bcryptCost)
	if err != nil {
		panic("auth: unable to precompute decoy hash: " + err.Error())
	}
	return &Service{store: store, decoyHash: decoy, log: log}
}

type RegisterInput struct {
	Email    string
	Password string
}

func (s *Service) Register(ctx context.Context, in RegisterInput) (*users.User, error) {
	email := strings.ToLower(strings.TrimSpace(in.Email))
	if !validEmail(email) {
		return nil, ErrInvalidEmail
	}
	if len(in.Password) < minPasswordLength || len(in.Password) > maxPasswordBytes {
		return nil, ErrWeakPassword
	}
	hash, err := HashPassword(in.Password)
	if err != nil {
		return nil, err
	}
	user, err := s.store.Create(ctx, email, hash, defaultRole)
	if errors.Is(err, users.ErrEmailTaken) {
		return nil, ErrEmailTaken
	}
	if err != nil {
		return nil, err
	}
	s.log.Info("user registered", slog.String("user_id", user.ID))
	return user, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (*users.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	user, err := s.store.ByEmail(ctx, email)
	if errors.Is(err, users.ErrNotFound) {
		_ = VerifyPassword(string(s.decoyHash), password)
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}
	if err := VerifyPassword(user.PasswordHash, password); err != nil {
		return nil, ErrInvalidCredentials
	}
	return user, nil
}

func validEmail(email string) bool {
	if len(email) < 3 || len(email) > 254 || strings.ContainsAny(email, " \t\r\n") {
		return false
	}
	addr, err := mail.ParseAddress(email)
	return err == nil && addr.Address == email && strings.Contains(email, "@")
}
