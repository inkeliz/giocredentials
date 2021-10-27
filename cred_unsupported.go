//+build !js,!android,!windows

package giocredentials

type config struct{}

func (m *Manager) view() (credentials []*Credential, err error) {
	return nil, ErrNotAvailableAPI
}

func (m *Manager) add(cred *Credential) error {
	return ErrNotAvailableAPI
}

func (m *Manager) remove(cred *Credential) error {
	return ErrNotAvailableAPI
}
