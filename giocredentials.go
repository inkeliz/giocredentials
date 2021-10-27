package giocredentials

import (
	"errors"
	"strings"
)

var (
	ErrNotAvailableAPI   = errors.New("cred: no compatible API available")
	ErrInvalidCredential = errors.New("cred: username have invalid character")
	ErrInvalidManager    = errors.New("cred: manager identifier/comment is invalid")
	ErrUserDecline       = errors.New("cred: user refuse to choose the user or don't allow to store the credentials")
)

type Manager struct {
	// Identifier must be unique per app, like the name of the app.
	Identifier string

	// Comment should describe the app.
	Comment string

	// AllowUnsafe defines if it should use less safe methods to store passwords, if there's no better option.
	AllowUnsafe bool

	config
}

func NewManager(identifier string) *Manager {
	return &Manager{Identifier:  identifier}
}

func (m *Manager) valid() error {
	if strings.Contains(m.Identifier, "\x00") || strings.Contains(m.Comment, "\x00") {
		return ErrInvalidManager
	}
	return nil
}

// View list all credentials available for the app.
func (m *Manager) View() ([]*Credential, error) {
	if err := m.valid(); err != nil {
		return nil, err
	}
	return m.view()
}

// Add adds the given Credential to the credential manager.
func (m *Manager) Add(acc *Credential) error {
	if err := m.valid(); err != nil {
		return err
	}
	if err := acc.valid(); err != nil {
		return err
	}
	return m.add(acc)
}

// Remove removes the given Credential from the credential manager.
// It may delete the credential even if the Credential.Password didn't match.
func (m *Manager) Remove(acc *Credential) error {
	if err := m.valid(); err != nil {
		return err
	}
	return m.remove(acc)
}

type Credential struct {
	// Username defines the username of the user.
	Username string

	// Password defines the password (or equivalent credentials) of the user.
	Password []byte
}

func NewCredential(identifier string, password []byte) *Credential {
	return &Credential{Username: identifier, Password: password}
}

func (c *Credential) valid() error {
	if strings.Contains(c.Username, "\x00") {
		return ErrInvalidCredential
	}
	if len(c.Username) > 512 || len(c.Password) > 512 {
		return ErrInvalidCredential
	}
	return nil
}