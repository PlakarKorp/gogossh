package ssh2

/*
#include <libssh2.h>
#include <libssh2_sftp.h>
#include <stdlib.h>
*/
import "C"

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"syscall"
	"time"
	"unsafe"
)

type SftpSession struct {
	parent *SshSession
	ptr    *C.LIBSSH2_SFTP
}

// Implements io.Reader interface
type SftpFile struct {
	parent *SftpSession
	ptr    *C.LIBSSH2_SFTP_HANDLE
}

type SftpDir struct {
	parent *SftpSession
	ptr    *C.LIBSSH2_SFTP_HANDLE
}

type sftpFileInfo struct {
	name  string
	attrs *C.LIBSSH2_SFTP_ATTRIBUTES
}

func (fi *sftpFileInfo) Name() string {
	return fi.name
}

func (fi *sftpFileInfo) Size() int64 {
	if fi.attrs.flags&C.LIBSSH2_SFTP_ATTR_SIZE != 0 {
		return int64(fi.attrs.filesize)
	} else {
		return 0
	}
}

func (fi *sftpFileInfo) Mode() os.FileMode {
	if fi.attrs.flags&C.LIBSSH2_SFTP_ATTR_PERMISSIONS != 0 {
		mode := os.FileMode(fi.attrs.permissions & (C.LIBSSH2_SFTP_S_IRWXU | C.LIBSSH2_SFTP_S_IRWXG | C.LIBSSH2_SFTP_S_IRWXO))

		switch fi.attrs.permissions & C.LIBSSH2_SFTP_S_IFMT {
		case C.LIBSSH2_SFTP_S_IFIFO:
			mode |= os.ModeNamedPipe
		case C.LIBSSH2_SFTP_S_IFCHR:
			mode |= os.ModeDevice | os.ModeCharDevice
		case C.LIBSSH2_SFTP_S_IFDIR:
			mode |= os.ModeDir
		case C.LIBSSH2_SFTP_S_IFBLK:
			mode |= os.ModeDevice
		case C.LIBSSH2_SFTP_S_IFREG:
		case C.LIBSSH2_SFTP_S_IFLNK:
			mode |= os.ModeSymlink
		case C.LIBSSH2_SFTP_S_IFSOCK:
			mode |= os.ModeSocket
		}

		return mode
	} else {
		return 0
	}
}

func (fi *sftpFileInfo) ModTime() time.Time {
	if fi.attrs.flags&C.LIBSSH2_SFTP_ATTR_ACMODTIME != 0 {
		return time.Unix(int64(fi.attrs.mtime), 0)
	} else {
		return time.Time{}
	}
}

func (fi *sftpFileInfo) IsDir() bool {
	return fi.Mode().IsDir()
}

func (fi *sftpFileInfo) Sys() any {
	return nil
}

func (ss *SshSession) SftpInit() (*SftpSession, error) {
	sftpSession := &SftpSession{parent: ss}
	sftpSession.ptr = C.libssh2_sftp_init(ss.ptr)

	if sftpSession.ptr == nil {
		return nil, fmt.Errorf("failed to initialize sftp session")
	}

	return sftpSession, nil
}

func (ss *SftpSession) Shutdown() error {
	return wrapSshError(C.libssh2_sftp_shutdown(ss.ptr))
}

type OpenFlags int

const (
	FXFRead   OpenFlags = C.LIBSSH2_FXF_READ
	FXFWrite  OpenFlags = C.LIBSSH2_FXF_WRITE
	FXFAppend OpenFlags = C.LIBSSH2_FXF_APPEND
	FXFCreat  OpenFlags = C.LIBSSH2_FXF_CREAT
	FXFTrunc  OpenFlags = C.LIBSSH2_FXF_TRUNC
	FXFExcl   OpenFlags = C.LIBSSH2_FXF_EXCL
)

type FileMode int

const (
	/* Read, write, execute/search by owner */
	S_IRWXU FileMode = C.LIBSSH2_SFTP_S_IRWXU
	S_IRUSR FileMode = C.LIBSSH2_SFTP_S_IRUSR
	S_IWUSR FileMode = C.LIBSSH2_SFTP_S_IWUSR
	S_IXUSR FileMode = C.LIBSSH2_SFTP_S_IXUSR

	/* Read, write, execute/search by group */
	S_IRWXG FileMode = C.LIBSSH2_SFTP_S_IRWXG
	S_IRGRP FileMode = C.LIBSSH2_SFTP_S_IRGRP
	S_IWGRP FileMode = C.LIBSSH2_SFTP_S_IWGRP
	S_IXGRP FileMode = C.LIBSSH2_SFTP_S_IXGRP

	/* Read, write, execute/search by others */
	S_IRWXO FileMode = C.LIBSSH2_SFTP_S_IRWXO
	S_IROTH FileMode = C.LIBSSH2_SFTP_S_IROTH
	S_IWOTH FileMode = C.LIBSSH2_SFTP_S_IWOTH
	S_IXOTH FileMode = C.LIBSSH2_SFTP_S_IXOTH
)

// Only call this when there was an error, other you'll get an unspecified error.
func (ss *SftpSession) GetLastError() error {
	code := SftpErrorCode(C.libssh2_sftp_last_error(ss.ptr))

	switch code {
	case FX_OK:
		return nil
	case FX_EOF:
		return io.EOF
	case FX_NO_SUCH_FILE:
		return os.ErrNotExist
	case FX_PERMISSION_DENIED:
		return os.ErrPermission
	default:
		return errors.New(code.String())
	}
}

func (ss *SftpSession) OpenFile(path string, flags OpenFlags, mode FileMode) (*SftpFile, error) {
	pathnameCStr := C.CString(path)
	defer C.free(unsafe.Pointer(pathnameCStr))

	handle := &SftpFile{parent: ss}
	handle.ptr = C.libssh2_sftp_open_ex(ss.ptr, pathnameCStr, C.uint(len(path)), C.ulong(flags), C.long(mode), C.LIBSSH2_SFTP_OPENFILE)

	if handle.ptr == nil {
		// The API is a bit weird here, you are supposed to get an SFTP error
		// (for lack of privileges etc) but if the transport layer fails you
		// get no error here and you have to check the session level error.
		// To be on the safe side (avoiding a NULL deref later on) we even
		// introduce a fail safe error so that no one uses the resulting
		// SftpHandle.
		if err := ss.GetLastError(); err != nil {
			return nil, ss.GetLastError()
		} else if err := ss.parent.GetLastError(); err != nil {
			return nil, ss.parent.GetLastError()
		} else {
			return nil, fmt.Errorf("unable to open file %s, unknown error\n", path)
		}
	}

	return handle, nil
}

func (ss *SftpSession) OpenDir(path string, flags OpenFlags, mode FileMode) (*SftpDir, error) {
	pathnameCStr := C.CString(path)
	defer C.free(unsafe.Pointer(pathnameCStr))

	handle := &SftpDir{parent: ss}
	handle.ptr = C.libssh2_sftp_open_ex(ss.ptr, pathnameCStr, C.uint(len(path)), C.ulong(flags), C.long(mode), C.LIBSSH2_SFTP_OPENDIR)

	if handle.ptr == nil {
		// ditto
		if err := ss.GetLastError(); err != nil {
			return nil, ss.GetLastError()
		} else if err := ss.parent.GetLastError(); err != nil {
			return nil, ss.parent.GetLastError()
		} else {
			return nil, fmt.Errorf("unable to open dir %s, unknown error\n", path)
		}
	}

	return handle, nil
}

func (ss *SftpSession) Mkdir(path string, mode FileMode) error {
	pathnameCStr := C.CString(path)
	defer C.free(unsafe.Pointer(pathnameCStr))

	rc := C.libssh2_sftp_mkdir_ex(ss.ptr, pathnameCStr, C.uint(len(path)), C.long(mode))
	if rc < 0 {
		if err := ss.GetLastError(); err != nil {
			return ss.GetLastError()
		} else {
			return wrapSshError(rc)
		}
	}

	return nil
}

func (ss *SftpSession) MkdirAll(path string, perm FileMode) error {
	// Fast path: if we can tell whether path is a directory or file, stop with success or error.
	dir, err := ss.Stat(path)
	if err == nil {
		if dir.IsDir() {
			return nil
		}
		return &os.PathError{Op: "mkdir", Path: path, Err: syscall.ENOTDIR}
	}

	// Slow path: make sure parent exists and then call Mkdir for path.

	// Extract the parent folder from path by first removing any trailing
	// path separator and then scanning backward until finding a path
	// separator or reaching the beginning of the string.
	i := len(path) - 1
	for i >= 0 && path[i] == '/' {
		i--
	}
	for i >= 0 && path[i] != '/' {
		i--
	}
	if i < 0 {
		i = 0
	}

	// If there is a parent directory, and it is not the volume name,
	// recurse to ensure parent directory exists.
	if parent := path[:i]; len(parent) > 0 {
		err = ss.MkdirAll(parent, perm)
		if err != nil {
			return err
		}
	}

	// Parent now exists; invoke Mkdir and use its result.
	err = ss.Mkdir(path, perm)
	if err != nil {
		// Handle arguments like "foo/." by
		// double-checking that directory doesn't exist.
		dir, err1 := ss.Lstat(path)
		if err1 == nil && dir.IsDir() {
			return nil
		}
		return err
	}
	return nil
}

func (ss *SftpSession) Rmdir(path string) error {
	pathnameCStr := C.CString(path)
	defer C.free(unsafe.Pointer(pathnameCStr))

	return wrapSshError(C.libssh2_sftp_rmdir_ex(ss.ptr, pathnameCStr, C.uint(len(path))))
}

func (ss *SftpSession) Unlink(path string) error {
	pathnameCStr := C.CString(path)
	defer C.free(unsafe.Pointer(pathnameCStr))

	return wrapSshError(C.libssh2_sftp_unlink_ex(ss.ptr, pathnameCStr, C.uint(len(path))))
}

func (ss *SftpSession) Rename(oldname, newname string) error {
	oldnameCStr := C.CString(oldname)
	defer C.free(unsafe.Pointer(oldnameCStr))
	newnameCStr := C.CString(newname)
	defer C.free(unsafe.Pointer(newnameCStr))

	// Default behavior for rename, maybe we want RenameEx at some point but I'm not sure.
	flags := C.LIBSSH2_SFTP_RENAME_OVERWRITE | C.LIBSSH2_SFTP_RENAME_ATOMIC | C.LIBSSH2_SFTP_RENAME_NATIVE
	return wrapSshError(C.libssh2_sftp_rename_ex(ss.ptr, oldnameCStr, C.uint(len(oldname)), newnameCStr, C.uint(len(newname)), C.long(flags)))
}

func (ss *SftpSession) Stat(pathname string) (os.FileInfo, error) {
	pathnameCStr := C.CString(pathname)
	defer C.free(unsafe.Pointer(pathnameCStr))

	attrs := &C.LIBSSH2_SFTP_ATTRIBUTES{}
	ret := C.libssh2_sftp_stat_ex(ss.ptr, pathnameCStr, C.uint(len(pathname)), C.LIBSSH2_SFTP_STAT, attrs)
	if ret < 0 {
		return nil, wrapSshError(ret)
	}

	return &sftpFileInfo{name: path.Base(pathname), attrs: attrs}, nil
}

func (ss *SftpSession) Lstat(pathname string) (os.FileInfo, error) {
	pathnameCStr := C.CString(pathname)
	defer C.free(unsafe.Pointer(pathnameCStr))

	attrs := &C.LIBSSH2_SFTP_ATTRIBUTES{}
	ret := C.libssh2_sftp_stat_ex(ss.ptr, pathnameCStr, C.uint(len(pathname)), C.LIBSSH2_SFTP_LSTAT, attrs)
	if ret < 0 {
		return nil, wrapSshError(ret)
	}

	return &sftpFileInfo{name: path.Base(pathname), attrs: attrs}, nil
}

func (ss *SftpSession) Chown(pathname string, uid, gid int) error {
	pathnameCStr := C.CString(pathname)
	defer C.free(unsafe.Pointer(pathnameCStr))

	attrs := &C.LIBSSH2_SFTP_ATTRIBUTES{}
	attrs.flags = C.LIBSSH2_SFTP_ATTR_UIDGID
	attrs.uid = C.ulong(uid)
	attrs.gid = C.ulong(gid)

	ret := C.libssh2_sftp_stat_ex(ss.ptr, pathnameCStr, C.uint(len(pathname)), C.LIBSSH2_SFTP_SETSTAT, attrs)
	if ret < 0 {
		return wrapSshError(ret)
	}

	return nil
}

func (ss *SftpSession) Chmod(pathname string, mode os.FileMode) error {
	pathnameCStr := C.CString(pathname)
	defer C.free(unsafe.Pointer(pathnameCStr))

	attrs := &C.LIBSSH2_SFTP_ATTRIBUTES{}
	attrs.flags = C.LIBSSH2_SFTP_ATTR_PERMISSIONS
	attrs.permissions = C.ulong(mode & os.ModePerm)

	ret := C.libssh2_sftp_stat_ex(ss.ptr, pathnameCStr, C.uint(len(pathname)), C.LIBSSH2_SFTP_SETSTAT, attrs)
	if ret < 0 {
		return wrapSshError(ret)
	}

	return nil
}

// Low level Setstat function, will only set w/e is present in the map.
/*
func (ss *SftpSession) Setstat(pathname string, attrs map[StatAttrFlags]any) error {
	pathnameCStr := C.CString(pathname)
	defer C.free(unsafe.Pointer(pathnameCStr))

	cAttrs := &C.LIBSSH2_SFTP_ATTRIBUTES{}
	for f, v := range attrs {
		switch f {
		case AttrFlagSize:
			cAttrs.
		case AttrFlagUGID:
		case AttrFlagPerm:
		case AttrFlagAMTime:
		case AttrFlagExtended:
		}

	}

	ret := C.libssh2_sftp_stat_ex(ss.ptr, pathnameCStr, C.uint(len(pathname)), C.LIBSSH2_SFTP_SETSTAT, attrs)
	if ret <= 0 {
		return nil, wrapSshError(ret)
	}

	return nil
}
*/

/* Directory specific methods */
func (ss *SftpSession) ReadDir(path string) ([]os.FileInfo, error) {
	dirfp, err := ss.OpenDir(path, 0, 0)
	if err != nil {
		return nil, err
	}
	defer dirfp.Close()

	var out []os.FileInfo
	buf := make([]byte, 512)
	for {
		attrs := &C.LIBSSH2_SFTP_ATTRIBUTES{}
		ret := C.libssh2_sftp_readdir_ex(dirfp.ptr, (*C.char)(unsafe.Pointer(&buf[0])), C.size_t(512), nil, 0, attrs)
		if ret <= 0 {
			break
		}

		name := string(buf[:ret])

		// For compatibility with the old go-sftp library skip cur/parent dir.
		if name == "." || name == ".." {
			continue
		}

		out = append(out, &sftpFileInfo{name, attrs})
	}

	return out, nil
}

func (ss *SftpDir) Stat() (os.FileInfo, error) {
	attrs := &C.LIBSSH2_SFTP_ATTRIBUTES{}
	ret := C.libssh2_sftp_fstat_ex(ss.ptr, attrs, 0)
	if ret < 0 {
		return nil, wrapSshError(ret)
	}

	return &sftpFileInfo{name: path.Base(""), attrs: attrs}, nil
}

func (ss *SftpDir) Close() error {
	return wrapSshError(C.libssh2_sftp_close_handle(ss.ptr))
}

/* File specific methods */
func (ss *SftpFile) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	// The library doesn't hold a ref to the buffer so it's safe to just pass
	// the pointer here it can't be garbage collected in between.
	n := C.libssh2_sftp_read(ss.ptr, (*C.char)(unsafe.Pointer(&p[0])), C.size_t(len(p)))
	if n < 0 {
		return 0, wrapSshError(C.int(n))
	} else if n == 0 {
		return 0, io.EOF
	} else {
		return int(n), nil
	}
}

func (ss *SftpFile) Write(p []byte) (int, error) {
	// XXX Review int types in here.
	var written int = 0
	for written < len(p) {
		leftover := len(p) - written
		n := int(C.libssh2_sftp_write(ss.ptr, (*C.char)(unsafe.Pointer(&p[written])), C.size_t(leftover)))
		if n < 0 {
			return n, wrapSshError(C.int(n))
		}

		written += n
	}

	return written, nil
}

func (ss *SftpFile) Stat() (os.FileInfo, error) {
	attrs := &C.LIBSSH2_SFTP_ATTRIBUTES{}
	ret := C.libssh2_sftp_fstat_ex(ss.ptr, attrs, 0)
	if ret < 0 {
		return nil, wrapSshError(ret)
	}

	return &sftpFileInfo{name: path.Base(""), attrs: attrs}, nil
}

func (ss *SftpFile) Seek(offset int64) {
	C.libssh2_sftp_seek64(ss.ptr, C.libssh2_uint64_t(offset))
}

func (ss *SftpFile) Rewind() {
	ss.Seek(0)
}

func (ss *SftpFile) Tell() int64 {
	return int64(C.libssh2_sftp_tell64(ss.ptr))
}

func (ss *SftpFile) Close() error {
	return wrapSshError(C.libssh2_sftp_close_handle(ss.ptr))
}
