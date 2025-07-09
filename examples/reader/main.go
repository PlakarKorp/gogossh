package main

import (
	"bytes"
	"crypto/rand"
	"fmt"
	ssh "gogossh"
	"io"
	"net"
	"strings"
)

const (
	server   = "localhost:2222"
	user     = "foo"
	password = "pass"
	file     = "/shared/foo.txt"
	basepath = "/shared"
)

func main() {
	conn, err := net.Dial("tcp", server)
	if err != nil {
		fmt.Printf("Failed to connect to %q: %s\n", server, err)
		return
	}

	if err := ssh.Init(0); err != nil {
		fmt.Printf("Failed to init ssh %s\n", err)
		return
	}
	defer ssh.Exit()

	session, err := ssh.SessionInit()
	if err != nil {
		fmt.Printf("Failed to init session %s\n", err)
		return
	}
	defer session.Close()

	if err := session.Handshake(conn); err != nil {
		fmt.Printf("Failed to handshake %s\n", err)
		return
	}
	defer session.Disconnect("Normal Shutdown")

	fingerprint, err := session.HostKeyHash(ssh.HashTypeSHA1)
	if err != nil {
		fmt.Printf("Failed to fingerprint server %s\n", err)
		return
	}

	for i, f := range strings.Split(fingerprint, ",") {
		fmt.Printf("[*] Fingerprint %d: %x \n", i, f)
	}

	authlist, err := session.UserAuthList(user)
	if err != nil {
		fmt.Printf("Failed to get auth list for user %s: %s\n", user, err)
		return
	}

	fmt.Printf("[*] Advertised auth methods:\n")
	for _, auth := range strings.Split(authlist, ",") {
		fmt.Printf("Auth method %q\n", auth)
	}

	if err := session.UserAuthPassword(user, password); err != nil {
		fmt.Printf("Failed to authenticate user %s\n", err)
		return
	}

	fmt.Printf("[*] User %q authenticated\n", user)

	sftpSession, err := session.SftpInit()
	if err != nil {
		fmt.Printf("Failed to init sftp session %s\n", err)
		return
	}
	defer sftpSession.Shutdown()

	fileHandle, err := sftpSession.OpenFile(file, ssh.FXFRead, 0)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer fileHandle.Close()

	out := make([]byte, 1000)
	fmt.Printf("[*] Reading file %q\n", file)
	n, err := fileHandle.Read(out)
	if err != nil {
		fmt.Printf("Failed to read file %q\n", err)
		return
	}
	fmt.Printf("Read %d bytes out of file\nData is %q\n", n, string(out[:n]))

	dirHandle, err := sftpSession.OpenDir(basepath, ssh.FXFRead, 0)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer dirHandle.Close()

	fileHandle2, err := sftpSession.OpenFile("/shared/caca.rand", ssh.FXFRead|ssh.FXFCreat|ssh.FXFWrite|ssh.FXFTrunc, 0)
	if err != nil {
		fmt.Printf("error opening file ? %s\n", err)
		return
	}
	defer fileHandle2.Close()

	buf := make([]byte, 23)
	if _, err := rand.Read(buf); err != nil {
		fmt.Println(err)
		return
	}

	written, err := io.Copy(fileHandle2, bytes.NewReader(buf))
	if err != nil {
		fmt.Printf("Failed to copy %s\n", err)
		return
	}

	fmt.Printf("Wrote %d\n", written)
}
