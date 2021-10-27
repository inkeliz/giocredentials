package giocredentials

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"syscall/js"
)

type config struct{}

func (m *Manager) view() ([]*Credential, error) {
	if !isAvailableCredentialsManagementAPI() {
		return m.viewUnsafe()
	}
	return m.viewDefault()
}

func (m *Manager) viewDefault() ([]*Credential, error) {
	var (
		success, failure js.Func

		value = make(chan js.Value, 1)
		err   = make(chan error, 1)
	)
	defer func() {
		close(value)
		close(err)
	}()

	success = js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		success.Release()
		failure.Release()

		if len(args) > 0 {
			value <- args[0]
			return nil
		}
		err <- ErrUserDecline
		return nil
	})
	failure = js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		success.Release()
		failure.Release()

		err <- ErrUserDecline
		return nil
	})

	go func() {
		api := js.Global().Get("navigator").Get("credentials").Call("get", map[string]interface{}{"password": true})
		api.Call("then", success, failure)
	}()

	for {
		select {
		case v := <-value:
			if !v.Truthy() {
				return nil, ErrUserDecline
			}

			user := v.Get("id").String()
			pass, _ := base64.StdEncoding.DecodeString(v.Get("password").String())

			return []*Credential{{Username: user, Password: pass}}, nil
		case v := <-err:
			return nil, v
		}
	}
}

func (m *Manager) viewUnsafe() (credentials []*Credential, err error) {
	if !m.AllowUnsafe {
		return nil, ErrNotAvailableAPI
	}

	credentials = make([]*Credential, 0, 8)

	api := js.Global().Get("localStorage")
	for i := 0; i < api.Get("length").Int(); i++ {
		key := api.Call("key", i).String()
		if !strings.HasPrefix(key, m.Identifier+":") {
			continue
		}

		_cred := &Credential{}
		if err := json.Unmarshal([]byte(api.Call("getItem", key).String()), _cred); err != nil {
			continue
		}

		credentials = append(credentials, _cred)
	}

	return credentials, nil
}

func (m *Manager) add(cred *Credential) error {
	if !isAvailableCredentialsManagementAPI() {
		return m.addUnsafe(cred)
	}
	return m.addDefault(cred)
}

func (m *Manager) addDefault(cred *Credential) error {
	var (
		cResp = make(chan js.Value, 1)
		cErr  = make(chan js.Value, 1)
	)

	_cred := js.Global().Get("PasswordCredential").New(
		map[string]interface{}{"id": cred.Username, "password": base64.StdEncoding.EncodeToString(cred.Password)},
	)

	s := js.FuncOf(func(_ js.Value, v []js.Value) interface{} {
		cResp <- js.Value{}
		return nil
	})
	e := js.FuncOf(func(_ js.Value, v []js.Value) interface{} {
		if len(v) > 0 {
			cErr <- v[0]
		}
		return nil
	})

	go func() {
		api := js.Global().Get("navigator").Get("credentials").Call("store", _cred)
		api.Call("then", s, e)
	}()

	for {
		select {
		case <-cResp:
			return nil
		case <-cErr:
			return ErrUserDecline
		}
	}
}

func (m *Manager) addUnsafe(cred *Credential) error {
	if !m.AllowUnsafe {
		return ErrNotAvailableAPI
	}

	_cred, err := json.Marshal(cred)
	if err != nil {
		return nil
	}

	js.Global().Get("localStorage").Call("setItem", m.Identifier+":"+cred.Username, string(_cred))
	return nil
}

func (m *Manager) remove(cred *Credential) error {
	if isAvailableCredentialsManagementAPI() {
		return m.removeUnsafe(cred)
	}
	return m.removeDefault(cred)
}

func (m *Manager) removeDefault(_ *Credential) error {
	// It's not possible to remove credentials.
	return nil
}

func (m *Manager) removeUnsafe(cred *Credential) error {
	js.Global().Get("localStorage").Call("removeItem", m.Identifier+":"+cred.Username)
	return nil
}

func isAvailableCredentialsManagementAPI() (ok bool) {
	if _, ok := get(js.Global(), "navigator", "credentials", "get"); !ok {
		return false
	}
	if _, ok := get(js.Global(), "PasswordCredential"); !ok {
		return false
	}
	return true
}

func get(src js.Value, path ...string) (object js.Value, ok bool) {
	object = src
	for _, v := range path {
		if try := object.Get(v); try.Truthy() {
			object = try
		} else {
			return js.Value{}, false
		}
	}
	return object, true
}
