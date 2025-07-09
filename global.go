package ssh2

/*
#cgo pkg-config: libssh2
#include <libssh2.h>
*/
import "C"

type InitFlags int

const (
	INIT_NO_CRYPTO InitFlags = C.LIBSSH2_INIT_NO_CRYPTO
)

func Init(flags InitFlags) error {
	return wrapSshError(C.libssh2_init(C.int(flags)))
}

func Exit() {
	C.libssh2_exit()
}
