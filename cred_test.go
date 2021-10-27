package giocredentials

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"runtime"
	"testing"
	"time"
)

func TestManagerWindows(t *testing.T) {
	manager := testNewManager(t)
	creds := []*Credential{
		testNewCredential(t),
		testNewCredential(t),
		testNewCredential(t),
	}

	if err := manager.Add(creds[0]); err != nil {
		t.Fatal(err)
	}

	time.Sleep(5 * time.Second)

	if err := manager.Add(creds[1]); err != nil {
		t.Fatal(err)
	}

	list, err := manager.View()
	if err != nil {
		t.Fatal(err)
	}

	if found := helperFind(list, creds[0], creds[1]); !found {
		t.Fatal("credential not found in View()")
	}
	if found := helperFind(list, creds[2]); found {
		t.Fatal("credential that was never added was found")
	}

	if err := manager.Remove(creds[0]); err != nil {
		t.Fatal(err)
	}

	if runtime.GOOS == "js" {
		return
	}

	list, err = manager.View()
	if err != nil {
		t.Fatal(err)
	}
	if found := helperFind(list, creds[0]); found {
		t.Fatal("credential was not deleted")
	}
	if found := helperFind(list, creds[1]); !found {
		t.Fatal("credential was deleted wrongly")
	}
}

func helperFind(list []*Credential, expected ...*Credential) (found bool) {
	for _, cred := range list {
		for _, c := range expected {
			if cred.Username == c.Username && bytes.Equal(cred.Password, c.Password) {
				found = true
			}
		}
	}
	return found
}

func testNewCredential(t *testing.T) *Credential {
	t.Helper()
	u, p := make([]byte, 16), make([]byte, 32)
	rand.Read(u)
	rand.Read(p)
	return &Credential{
		Username: hex.EncodeToString(u),
		Password: p,
	}
}

func testNewManager(t *testing.T) *Manager {
	t.Helper()
	return &Manager{
		Identifier:  "GioCredentialsTest",
		Comment:     "Testing of GioCredentials",
		AllowUnsafe: false,
	}
}
