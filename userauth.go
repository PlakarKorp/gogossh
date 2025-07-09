package ssh2

/*
#include <libssh2.h>
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"unsafe"
)

// UserAuthList gets the remote's list of supported authentication methods
// It returns the list as a comma separated value
func (ss *SshSession) UserAuthList(username string) (string, error) {
	usernameCStr := C.CString(username)
	defer C.free(unsafe.Pointer(usernameCStr))

	digestCstr := C.libssh2_userauth_list(ss.ptr, usernameCStr, C.uint(len(username)))

	if digestCstr == nil {
		return "", fmt.Errorf("failed to list remote's supported authentication methods")
	}

	return C.GoString(digestCstr), nil
}

// UserAuthPassword auth the user with the given password.
// This function may return ErrorEagain in case it would block
func (ss *SshSession) UserAuthPassword(username, password string) error {
	usernameCStr := C.CString(username)
	defer C.free(unsafe.Pointer(usernameCStr))

	passwordCStr := C.CString(password)
	defer C.free(unsafe.Pointer(passwordCStr))

	return wrapSshError(C.libssh2_userauth_password_ex(ss.ptr, usernameCStr, C.uint(len(username)), passwordCStr, C.uint(len(password)), nil))
}
