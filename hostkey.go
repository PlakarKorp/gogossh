package ssh2

/*
#include <libssh2.h>
*/
import "C"
import "fmt"

type HashType int

const (
	HashTypeMD5    HashType = C.LIBSSH2_HOSTKEY_HASH_MD5
	HashTypeSHA1   HashType = C.LIBSSH2_HOSTKEY_HASH_SHA1
	HashTypeSHA256 HashType = C.LIBSSH2_HOSTKEY_HASH_SHA256
)

func (ss *SshSession) HostKeyHash(htype HashType) (string, error) {
	digestCstr := C.libssh2_hostkey_hash(ss.ptr, C.int(htype))

	if digestCstr == nil {
		return "", fmt.Errorf("failed to get remote's computed digest hostkey")
	}

	return C.GoString(digestCstr), nil
}
