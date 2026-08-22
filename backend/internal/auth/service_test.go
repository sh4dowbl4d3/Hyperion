package auth

import (
	"context"
	"errors"
	"testing"

	"moderndvwa/backend/internal/users"
)

type fakeStore struct {
	usersByEmail map[string]*users.User
	createErr    error
	created      *users.User
}

func newFakeStore() *fakeStore {
	return &fakeStore{usersByEmail: make(map[string]*users.User)}
}

func (f *fakeStore) Create(_ context.Context, email, passwordHash, role string) (*users.User, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	if _, exists := f.usersByEmail[email]; exists {
		return nil, users.ErrEmailTaken
	}
	u := &users.User{ID: "id-" + email, Email: email, PasswordHash: passwordHash, Role: role}
	f.usersByEmail[email] = u
	return u, nil
}

func (f *fakeStore) ByEmail(_ context.Context, email string) (*users.User, error) {
	if u, ok := f.usersByEmail[email]; ok {
		return u, nil
	}
	return nil, users.ErrNotFound
}

func (f *fakeStore) ByID(_ context.Context, id string) (*users.User, error) {
	for _, u := range f.usersByEmail {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, users.ErrNotFound
}

func newTestService(t *testing.T, store UserStore) *Service {
	t.Helper()
	svc := NewService(store, discardLogger())
	t.Cleanup(func() {})
	return svc
}

func TestRegisterCreatesUserWithLowercasedEmail(t *testing.T) {
	store := newFakeStore()
	svc := newTestService(t, store)

	user, err := svc.Register(context.Background(), RegisterInput{
		Email:    "  NewUser@Example.COM ",
		Password: "long-enough-password",
	})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if user.Email != "newuser@example.com" {
		t.Errorf("email = %q, want lowercased and trimmed", user.Email)
	}
	if user.Role != "user" {
		t.Errorf("role = %q, want user", user.Role)
	}
	stored := store.usersByEmail[user.Email]
	if stored == nil {
		t.Fatal("user not persisted under normalized email")
	}
	if stored.PasswordHash == "long-enough-password" || len(stored.PasswordHash) < 50 {
		t.Error("stored password must be a bcrypt hash, not plaintext or a short value")
	}
	if err := VerifyPassword(stored.PasswordHash, "long-enough-password"); err != nil {
		t.Errorf("hash does not verify against original password: %v", err)
	}
}

func TestRegisterValidationFailures(t *testing.T) {
	cases := []struct {
		name string
		in   RegisterInput
		want error
	}{
		{"missing at sign", RegisterInput{Email: "no-at-sign.example", Password: "long-enough-password"}, ErrInvalidEmail},
		{"empty email", RegisterInput{Email: "", Password: "long-enough-password"}, ErrInvalidEmail},
		{"email with space", RegisterInput{Email: "bad address@example.com", Password: "long-enough-password"}, ErrInvalidEmail},
		{"password too short", RegisterInput{Email: "a@example.com", Password: "short9"}, ErrWeakPassword},
		{"password over 72 bytes", RegisterInput{Email: "a@example.com", Password: string(make([]byte, 73))}, ErrWeakPassword},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := newTestService(t, newFakeStore()).Register(context.Background(), tc.in)
			if !errors.Is(err, tc.want) {
				t.Fatalf("error = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestRegisterPropagatesEmailTaken(t *testing.T) {
	store := newFakeStore()
	svc := newTestService(t, store)

	in := RegisterInput{Email: "dup@example.com", Password: "long-enough-password"}
	if _, err := svc.Register(context.Background(), in); err != nil {
		t.Fatalf("first register: %v", err)
	}
	_, err := svc.Register(context.Background(), in)
	if !errors.Is(err, ErrEmailTaken) {
		t.Fatalf("error = %v, want ErrEmailTaken", err)
	}
}

func TestLoginSuccessAndGenericFailure(t *testing.T) {
	store := newFakeStore()
	svc := newTestService(t, store)

	ctx := context.Background()
	if _, err := svc.Register(ctx, RegisterInput{Email: "login@example.com", Password: "long-enough-password"}); err != nil {
		t.Fatalf("register: %v", err)
	}

	user, err := svc.Login(ctx, "Login@Example.com", "long-enough-password")
	if err != nil {
		t.Fatalf("login with valid credentials failed: %v", err)
	}
	if user.Email != "login@example.com" {
		t.Errorf("logged in as %q, want login@example.com", user.Email)
	}

	for name, tc := range map[string]struct {
		email    string
		password string
	}{
		"unknown user":     {"ghost@example.com", "long-enough-password"},
		"wrong password":   {"login@example.com", "definitely-not-it"},
		"empty password":   {"login@example.com", ""},
		"case-insens mail": {"LOGIN@example.com", "wrong"},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := svc.Login(ctx, tc.email, tc.password)
			if !errors.Is(err, ErrInvalidCredentials) {
				t.Fatalf("error = %v, want ErrInvalidCredentials", err)
			}
		})
	}
}
