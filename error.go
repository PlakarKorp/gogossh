package ssh2

/*
#include <libssh2.h>
#include <libssh2_sftp.h>
*/
import "C"

import (
	"errors"
)

//go:generate stringer -type=ErrorCode
type ErrorCode int

/* Error Codes (defined by libssh2) */
const (
	ErrorNone                  ErrorCode = C.LIBSSH2_ERROR_NONE
	ErrorSocketNone            ErrorCode = C.LIBSSH2_ERROR_SOCKET_NONE
	ErrorBannerRecv            ErrorCode = C.LIBSSH2_ERROR_BANNER_RECV
	ErrorBannerSend            ErrorCode = C.LIBSSH2_ERROR_BANNER_SEND
	ErrorInvalidMac            ErrorCode = C.LIBSSH2_ERROR_INVALID_MAC
	ErrorKexFailure            ErrorCode = C.LIBSSH2_ERROR_KEX_FAILURE
	ErrorAlloc                 ErrorCode = C.LIBSSH2_ERROR_ALLOC
	ErrorSocketSend            ErrorCode = C.LIBSSH2_ERROR_SOCKET_SEND
	ErrorKeyExchangeFailure    ErrorCode = C.LIBSSH2_ERROR_KEY_EXCHANGE_FAILURE
	ErrorTimeout               ErrorCode = C.LIBSSH2_ERROR_TIMEOUT
	ErrorHostKeyInit           ErrorCode = C.LIBSSH2_ERROR_HOSTKEY_INIT
	ErrorHostKeySign           ErrorCode = C.LIBSSH2_ERROR_HOSTKEY_SIGN
	ErrorDecrypt               ErrorCode = C.LIBSSH2_ERROR_DECRYPT
	ErrorSocketDisconnect      ErrorCode = C.LIBSSH2_ERROR_SOCKET_DISCONNECT
	ErrorProto                 ErrorCode = C.LIBSSH2_ERROR_PROTO
	ErrorPasswordExpired       ErrorCode = C.LIBSSH2_ERROR_PASSWORD_EXPIRED
	ErrorFile                  ErrorCode = C.LIBSSH2_ERROR_FILE
	ErrorMethodNone            ErrorCode = C.LIBSSH2_ERROR_METHOD_NONE
	ErrorAuthenticationFailed  ErrorCode = C.LIBSSH2_ERROR_AUTHENTICATION_FAILED
	ErrorPublickeyUnrecognized ErrorCode = C.LIBSSH2_ERROR_PUBLICKEY_UNRECOGNIZED
	ErrorPublickeyUnverified   ErrorCode = C.LIBSSH2_ERROR_PUBLICKEY_UNVERIFIED
	ErrorChannelOutoforder     ErrorCode = C.LIBSSH2_ERROR_CHANNEL_OUTOFORDER
	ErrorChannelFailure        ErrorCode = C.LIBSSH2_ERROR_CHANNEL_FAILURE
	ErrorChannelRequestDenied  ErrorCode = C.LIBSSH2_ERROR_CHANNEL_REQUEST_DENIED
	ErrorChannelUnknown        ErrorCode = C.LIBSSH2_ERROR_CHANNEL_UNKNOWN
	ErrorChannelWindowExceeded ErrorCode = C.LIBSSH2_ERROR_CHANNEL_WINDOW_EXCEEDED
	ErrorChannelPacketExceeded ErrorCode = C.LIBSSH2_ERROR_CHANNEL_PACKET_EXCEEDED
	ErrorChannelClosed         ErrorCode = C.LIBSSH2_ERROR_CHANNEL_CLOSED
	ErrorChannelEofSent        ErrorCode = C.LIBSSH2_ERROR_CHANNEL_EOF_SENT
	ErrorScpProtocol           ErrorCode = C.LIBSSH2_ERROR_SCP_PROTOCOL
	ErrorZlib                  ErrorCode = C.LIBSSH2_ERROR_ZLIB
	ErrorSocketTimeout         ErrorCode = C.LIBSSH2_ERROR_SOCKET_TIMEOUT
	ErrorSftpProtocol          ErrorCode = C.LIBSSH2_ERROR_SFTP_PROTOCOL
	ErrorRequestDenied         ErrorCode = C.LIBSSH2_ERROR_REQUEST_DENIED
	ErrorMethodNotSupported    ErrorCode = C.LIBSSH2_ERROR_METHOD_NOT_SUPPORTED
	ErrorInval                 ErrorCode = C.LIBSSH2_ERROR_INVAL
	ErrorInvalidPollType       ErrorCode = C.LIBSSH2_ERROR_INVALID_POLL_TYPE
	ErrorPublickeyProtocol     ErrorCode = C.LIBSSH2_ERROR_PUBLICKEY_PROTOCOL
	ErrorEagain                ErrorCode = C.LIBSSH2_ERROR_EAGAIN
	ErrorBufferTooSmall        ErrorCode = C.LIBSSH2_ERROR_BUFFER_TOO_SMALL
	ErrorBadUse                ErrorCode = C.LIBSSH2_ERROR_BAD_USE
	ErrorCompress              ErrorCode = C.LIBSSH2_ERROR_COMPRESS
	ErrorOutOfBoundary         ErrorCode = C.LIBSSH2_ERROR_OUT_OF_BOUNDARY
	ErrorAgentProtocol         ErrorCode = C.LIBSSH2_ERROR_AGENT_PROTOCOL
	ErrorSocketRecv            ErrorCode = C.LIBSSH2_ERROR_SOCKET_RECV
	ErrorEncrypt               ErrorCode = C.LIBSSH2_ERROR_ENCRYPT
	ErrorBadSocket             ErrorCode = C.LIBSSH2_ERROR_BAD_SOCKET
	ErrorKnownHosts            ErrorCode = C.LIBSSH2_ERROR_KNOWN_HOSTS
	ErrorChannelWindowFull     ErrorCode = C.LIBSSH2_ERROR_CHANNEL_WINDOW_FULL
	ErrorKeyfileAuthFailed     ErrorCode = C.LIBSSH2_ERROR_KEYFILE_AUTH_FAILED
	ErrorRandgen               ErrorCode = C.LIBSSH2_ERROR_RANDGEN
	ErrorMissingUserauthBanner ErrorCode = C.LIBSSH2_ERROR_MISSING_USERAUTH_BANNER
	ErrorAlgoUnsupported       ErrorCode = C.LIBSSH2_ERROR_ALGO_UNSUPPORTED
	ErrorMacFailure            ErrorCode = C.LIBSSH2_ERROR_MAC_FAILURE
	ErrorHashInit              ErrorCode = C.LIBSSH2_ERROR_HASH_INIT
	ErrorHashCalc              ErrorCode = C.LIBSSH2_ERROR_HASH_CALC
	ErrorBannerNone            ErrorCode = C.LIBSSH2_ERROR_BANNER_NONE
)

func wrapSshError(ci C.int) error {
	if ci < 0 {
		code := ErrorCode(ci)
		return errors.New(code.String())
	} else {
		return nil
	}
}

//go:generate stringer -type=SftpErrorCode
type SftpErrorCode int

const (
	FX_OK                     SftpErrorCode = C.LIBSSH2_FX_OK
	FX_EOF                    SftpErrorCode = C.LIBSSH2_FX_EOF
	FX_NO_SUCH_FILE           SftpErrorCode = C.LIBSSH2_FX_NO_SUCH_FILE
	FX_PERMISSION_DENIED      SftpErrorCode = C.LIBSSH2_FX_PERMISSION_DENIED
	FX_FAILURE                SftpErrorCode = C.LIBSSH2_FX_FAILURE
	FX_BAD_MESSAGE            SftpErrorCode = C.LIBSSH2_FX_BAD_MESSAGE
	FX_NO_CONNECTION          SftpErrorCode = C.LIBSSH2_FX_NO_CONNECTION
	FX_CONNECTION_LOST        SftpErrorCode = C.LIBSSH2_FX_CONNECTION_LOST
	FX_OP_UNSUPPORTED         SftpErrorCode = C.LIBSSH2_FX_OP_UNSUPPORTED
	FX_INVALID_HANDLE         SftpErrorCode = C.LIBSSH2_FX_INVALID_HANDLE
	FX_NO_SUCH_PATH           SftpErrorCode = C.LIBSSH2_FX_NO_SUCH_PATH
	FX_FILE_ALREADY_EXISTS    SftpErrorCode = C.LIBSSH2_FX_FILE_ALREADY_EXISTS
	FX_WRITE_PROTECT          SftpErrorCode = C.LIBSSH2_FX_WRITE_PROTECT
	FX_NO_MEDIA               SftpErrorCode = C.LIBSSH2_FX_NO_MEDIA
	FX_NO_SPACE_ON_FILESYSTEM SftpErrorCode = C.LIBSSH2_FX_NO_SPACE_ON_FILESYSTEM
	FX_QUOTA_EXCEEDED         SftpErrorCode = C.LIBSSH2_FX_QUOTA_EXCEEDED
	FX_UNKNOWN_PRINCIPAL      SftpErrorCode = C.LIBSSH2_FX_UNKNOWN_PRINCIPAL
	FX_LOCK_CONFLICT          SftpErrorCode = C.LIBSSH2_FX_LOCK_CONFLICT
	FX_DIR_NOT_EMPTY          SftpErrorCode = C.LIBSSH2_FX_DIR_NOT_EMPTY
	FX_NOT_A_DIRECTORY        SftpErrorCode = C.LIBSSH2_FX_NOT_A_DIRECTORY
	FX_INVALID_FILENAME       SftpErrorCode = C.LIBSSH2_FX_INVALID_FILENAME
	FX_LINK_LOOP              SftpErrorCode = C.LIBSSH2_FX_LINK_LOOP
)
