package giocredentials

import (
	"golang.org/x/sys/windows"
	"reflect"
	"syscall"
	"unsafe"
)

type config struct{}

var (
	_advapi32 = windows.NewLazyDLL("advapi32.dll")

	_CredReadW      = _advapi32.NewProc("CredReadW")
	_CredWriteW     = _advapi32.NewProc("CredWriteW")
	_CredDeleteW    = _advapi32.NewProc("CredDeleteW")
	_CredFree       = _advapi32.NewProc("CredFree")
	_CredEnumerateW = _advapi32.NewProc("CredEnumerateW")
)

type credential struct {
	Flags          uint32
	Type           uint32
	TargetName     *uint16
	Comment        *uint16
	LastWritten    syscall.Filetime
	PasswordSize   uint32
	Password       uintptr
	Persist        uint32
	AttributeCount uint32
	Attributes     uintptr
	TargetAlias    *uint16
	UserName       *uint16
}

func (m *Manager) view() (credentials []*Credential, err error) {
	var (
		credsSize int
		creds     uintptr
	)

	resp, _, _ := _CredEnumerateW.Call(
		uintptr(unsafe.Pointer(windows.StringToUTF16Ptr(m.Identifier+"*"))),
		0,
		uintptr(unsafe.Pointer(&credsSize)),
		uintptr(unsafe.Pointer(&creds)),
	)
	if resp == 0 {
		return nil, ErrNotAvailableAPI
	}
	defer _CredFree.Call(creds)

	credentials = make([]*Credential, credsSize)
	for i, c := range *(*[]*credential)(createSlice(creds, credsSize, credsSize)) {
		if c == nil {
			continue
		}
		credentials[i] = &Credential{
			Username: windows.UTF16PtrToString(c.UserName),
			Password: make([]byte, c.PasswordSize),
		}
		copy(credentials[i].Password, *(*[]byte)(createSlice(c.Password, int(c.PasswordSize), int(c.PasswordSize))))
	}

	return credentials, nil
}

func (m *Manager) add(cred *Credential) error {
	resp, _, _ := _CredWriteW.Call(
		uintptr(unsafe.Pointer(&credential{
			Type:         0x01,
			TargetName:   windows.StringToUTF16Ptr(m.Identifier + ":" + cred.Username),
			Comment:      windows.StringToUTF16Ptr(m.Comment),
			PasswordSize: uint32(len(cred.Password)),
			Password:     uintptr(unsafe.Pointer(&cred.Password[0])),
			Persist:      0x02,
			UserName:     windows.StringToUTF16Ptr(cred.Username),
		})),
		0,
	)
	if resp == 0 {
		return ErrNotAvailableAPI
	}

	return nil
}

func (m *Manager) remove(cred *Credential) error {
	_CredDeleteW.Call(
		uintptr(unsafe.Pointer(windows.StringToUTF16Ptr(m.Identifier+":"+cred.Username))),
		0x01,
		0,
	)
	return nil
}

func createSlice(p uintptr, len int, cap int) unsafe.Pointer {
	h := &reflect.SliceHeader{}
	h.Data = p
	h.Len = len
	h.Cap = cap
	return unsafe.Pointer(h)
}
