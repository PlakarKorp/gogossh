package ssh2

/*
#cgo pkg-config: libssh2
#include <libssh2.h>
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"iter"
	"unsafe"
)

type Agent struct {
	ptr *C.LIBSSH2_AGENT
}

func (ss *SshSession) AgentInit() (*Agent, error) {
	agent := &Agent{}
	agent.ptr = C.libssh2_agent_init(ss.ptr)

	if agent.ptr == nil {
		return nil, fmt.Errorf("failed to initialize agent")
	}

	return agent, nil
}

func (ss *Agent) Free() {
	C.libssh2_agent_free(ss.ptr)
}

// Connect to the agent running at path, if it's empty uses $SSH_AUTH_SOCK
func (a *Agent) Connect(path string) error {
	if len(path) > 0 {
		pathnameCStr := C.CString(path)
		defer C.free(unsafe.Pointer(pathnameCStr))

		C.libssh2_agent_set_identity_path(a.ptr, pathnameCStr)
	}

	return wrapSshError(C.libssh2_agent_connect(a.ptr))
}

func (a *Agent) Disconnect() error {
	return wrapSshError(C.libssh2_agent_disconnect(a.ptr))
}

type AgentPublicKey struct {
	blob    string
	comment string
	ptr     *C.struct_libssh2_agent_publickey
}

// List identities. Note that it's fine to hold references to the yielded
// AgentPublicKey, they are only freed when calling Agent.Free()
func (a *Agent) ListIdentities() (iter.Seq2[*AgentPublicKey, error], error) {
	rc := C.libssh2_agent_list_identities(a.ptr)

	if rc < 0 {
		return nil, wrapSshError(rc)
	}

	return func(yield func(*AgentPublicKey, error) bool) {
		var cur, prev *C.struct_libssh2_agent_publickey
		for {
			rc = C.libssh2_agent_get_identity(a.ptr, &cur, prev)

			if rc < 0 {
				if !yield(nil, wrapSshError(rc)) {
					return
				}
			} else if rc == 0 {
				apk := &AgentPublicKey{}
				if !yield(apk, nil) {
					return
				}
			} else {
				// End of the identities list, rc == 1
				break
			}
		}
	}, nil
}

func (a *Agent) UserAuth(username string, apk *AgentPublicKey) error {
	usernameCStr := C.CString(username)
	defer C.free(unsafe.Pointer(usernameCStr))

	return wrapSshError(C.libssh2_agent_userauth(a.ptr, usernameCStr, apk.ptr))
}
