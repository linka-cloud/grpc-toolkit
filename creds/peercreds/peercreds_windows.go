//go:build windows

package peercreds

import (
	"fmt"
	"net"
	"reflect"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Get returns peer creds for the client connected to this server-side pipe
// connection. The conn must be a net.Conn returned from go-winio's ListenPipe.
func Get(conn net.Conn) (*Creds, error) {
	if conn == nil {
		return nil, ErrUnsupportedConnType
	}

	h, err := winioPipeHandle(conn)
	if err != nil {
		return nil, err
	}

	// Get client PID for this pipe instance.
	var pid uint32
	if err := windows.GetNamedPipeClientProcessId(h, &pid); err != nil {
		return nil, fmt.Errorf("GetNamedPipeClientProcessId: %w", err)
	}
	if pid == 0 {
		return nil, fmt.Errorf("GetNamedPipeClientProcessId returned pid=0")
	}

	// Open the client process with query rights.
	const processQueryLimitedInfo = windows.PROCESS_QUERY_LIMITED_INFORMATION
	ph, err := windows.OpenProcess(processQueryLimitedInfo, false, pid)
	if err != nil {
		return nil, fmt.Errorf("OpenProcess(%d): %w", pid, err)
	}
	defer windows.CloseHandle(ph)

	// Open the process token.
	var token windows.Token
	if err := windows.OpenProcessToken(ph, windows.TOKEN_QUERY, &token); err != nil {
		return nil, fmt.Errorf("OpenProcessToken: %w", err)
	}
	defer token.Close()

	// Get the token's user SID.
	tu, err := token.GetTokenUser()
	if err != nil {
		return nil, fmt.Errorf("GetTokenUser: %w", err)
	}

	return &Creds{
		uid: tu.User.Sid.String(),
		pid: int(pid),
	}, nil
}

// winioPipeHandle digs the underlying syscall HANDLE out of a go-winio
// pipe connection using reflect + unsafe. This depends on the current
// internal layout of github.com/Microsoft/go-winio:
//
//	type win32Pipe struct {
//	    *win32File
//	    path string
//	}
//
//	type win32MessageBytePipe struct {
//	    win32Pipe
//	    writeClosed bool
//	    readEOF     bool
//	}
//
//	type win32File struct {
//	    handle syscall.Handle
//	    ...
//	}
//
// See pipe.go + file.go in go-winio. :contentReference[oaicite:1]{index=1}
func winioPipeHandle(conn net.Conn) (windows.Handle, error) {
	v := reflect.ValueOf(conn)
	if !v.IsValid() {
		return 0, ErrUnsupportedConnType
	}

	// Peel off interface & pointer layers: net.Conn is an interface and the
	// concrete type is *win32Pipe or *win32MessageBytePipe.
	for v.Kind() == reflect.Interface || v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return 0, ErrUnsupportedConnType
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return 0, ErrUnsupportedConnType
	}

	var wfField reflect.Value

	// Case 1: *win32Pipe { *win32File; path string }
	if f := v.FieldByName("win32File"); f.IsValid() && f.Kind() == reflect.Ptr {
		wfField = f
	} else if v.NumField() > 0 {
		// Case 2: *win32MessageBytePipe { win32Pipe; ... }
		embedded := v.Field(0)
		if embedded.IsValid() && embedded.Kind() == reflect.Struct {
			if f2 := embedded.FieldByName("win32File"); f2.IsValid() && f2.Kind() == reflect.Ptr {
				wfField = f2
			}
		}
	}

	if !wfField.IsValid() || wfField.IsNil() {
		return 0, ErrUnsupportedConnType
	}

	// wfField is a *win32File. Its first field is "handle syscall.Handle".
	// We only need the first field, so we define a 1-field header type with
	// compatible layout and reinterpret the pointer.
	type win32FileHeader struct {
		Handle windows.Handle // same underlying type as syscall.Handle
	}

	ptr := unsafe.Pointer(wfField.Pointer())
	h := (*win32FileHeader)(ptr).Handle
	if h == 0 {
		return 0, fmt.Errorf("winio pipe handle is 0")
	}
	return h, nil
}
