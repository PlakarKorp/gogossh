package ssh2

/*
#include <libssh2.h>
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"net"
	"reflect"
	"unsafe"
)

type DisconnectCode int

const (
	DisconnectHostNotAllowedToConnect     DisconnectCode = C.SSH_DISCONNECT_HOST_NOT_ALLOWED_TO_CONNECT
	DisconnectProtocolError               DisconnectCode = C.SSH_DISCONNECT_PROTOCOL_ERROR
	DisconnectKeyExchangeFailed           DisconnectCode = C.SSH_DISCONNECT_KEY_EXCHANGE_FAILED
	DisconnectReserved                    DisconnectCode = C.SSH_DISCONNECT_RESERVED
	DisconnectMacError                    DisconnectCode = C.SSH_DISCONNECT_MAC_ERROR
	DisconnectCompressionError            DisconnectCode = C.SSH_DISCONNECT_COMPRESSION_ERROR
	DisconnectServiceNotAvailable         DisconnectCode = C.SSH_DISCONNECT_SERVICE_NOT_AVAILABLE
	DisconnectProtocolVersionNotSupported DisconnectCode = C.SSH_DISCONNECT_PROTOCOL_VERSION_NOT_SUPPORTED
	DisconnectHostKeyNotVerifiable        DisconnectCode = C.SSH_DISCONNECT_HOST_KEY_NOT_VERIFIABLE
	DisconnectConnectionLost              DisconnectCode = C.SSH_DISCONNECT_CONNECTION_LOST
	DisconnectByApplication               DisconnectCode = C.SSH_DISCONNECT_BY_APPLICATION
	DisconnectTooManyConnections          DisconnectCode = C.SSH_DISCONNECT_TOO_MANY_CONNECTIONS
	DisconnectAuthCancelledByUser         DisconnectCode = C.SSH_DISCONNECT_AUTH_CANCELLED_BY_USER
	DisconnectNoMoreAuthMethodsAvailable  DisconnectCode = C.SSH_DISCONNECT_NO_MORE_AUTH_METHODS_AVAILABLE
	DisconnectIllegalUserName             DisconnectCode = C.SSH_DISCONNECT_ILLEGAL_USER_NAME
)

type SshSession struct {
	ptr *C.LIBSSH2_SESSION
}

func SessionInit() (*SshSession, error) {
	sess := &SshSession{}
	sess.ptr = C.libssh2_session_init_ex(nil, nil, nil, nil)

	if sess.ptr == nil {
		return nil, fmt.Errorf("failed to create ssh session")
	}

	// Opinionated, but this is goland, everything is blocking we just punt to
	// goroutines.
	C.libssh2_session_set_blocking(sess.ptr, 1)

	return sess, nil
}

func (ss *SshSession) Handshake(conn net.Conn) error {
	v := reflect.Indirect(reflect.ValueOf(conn))
	con := v.FieldByName("conn")
	netFD := reflect.Indirect(con.FieldByName("fd"))
	pfd := netFD.FieldByName("pfd")
	fd := int(pfd.FieldByName("Sysfd").Int())
	return wrapSshError(C.libssh2_session_handshake(ss.ptr, C.libssh2_socket_t(fd)))
}

func (ss *SshSession) Disconnect(desc string) error {
	langCstr := C.CString("")
	descCstr := C.CString(desc)
	defer C.free(unsafe.Pointer(langCstr))
	defer C.free(unsafe.Pointer(descCstr))

	return wrapSshError(C.libssh2_session_disconnect_ex(ss.ptr, C.int(DisconnectByApplication), descCstr, langCstr))
}

func (ss *SshSession) Close() error {
	return wrapSshError(C.libssh2_session_free(ss.ptr))
}

func (ss *SshSession) GetLastError() error {
	return wrapSshError(C.libssh2_session_last_error(ss.ptr, nil, nil, 0))
}
