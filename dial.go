package nmail

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
)

func DefaultDialer(ctx context.Context, username, password, host, port string, tlsConfig *tls.Config) (*SMTPClient, error) {
	var addr = fmt.Sprintf("%s:%s", host, port)
	nClient, err := smtp.Dial(addr)
	if err != nil {
		return nil, err
	}

	if err = nClient.Hello("localhost"); err != nil {
		nClient.Close()
		return nil, err
	}

	if ok, _ := nClient.Extension("STARTTLS"); ok {
		if err = nClient.StartTLS(tlsConfig); err != nil {
			nClient.Close()
			return nil, err
		}
	}

	if ok, auth := nClient.Extension("AUTH"); ok {
		var nAuth smtp.Auth
		if strings.Contains(auth, "CRAM-MD5") {
			nAuth = smtp.CRAMMD5Auth(username, password)
		} else if strings.Contains(auth, "LOGIN") && !strings.Contains(auth, "PLAIN") {
			nAuth = LoginAuth(username, password)
		} else {
			nAuth = smtp.PlainAuth("", username, password, host)
		}
		if err = nClient.Auth(nAuth); err != nil {
			nClient.Close()
			return nil, err
		}
	}
	return &SMTPClient{Client: nClient}, nil
}

func TLSDialer(ctx context.Context, username, password, host, port string, tlsConfig *tls.Config) (*SMTPClient, error) {
	var addr = fmt.Sprintf("%s:%s", host, port)
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return nil, err
	}

	nClient, err := smtp.NewClient(conn, host)
	if err != nil {
		return nil, err
	}

	if err = nClient.Hello("localhost"); err != nil {
		nClient.Close()
		return nil, err
	}

	if ok, auth := nClient.Extension("AUTH"); ok {
		var nAuth smtp.Auth
		if strings.Contains(auth, "CRAM-MD5") {
			nAuth = smtp.CRAMMD5Auth(username, password)
		} else if strings.Contains(auth, "LOGIN") && !strings.Contains(auth, "PLAIN") {
			nAuth = LoginAuth(username, password)
		} else {
			nAuth = smtp.PlainAuth("", username, password, host)
		}
		if err = nClient.Auth(nAuth); err != nil {
			nClient.Close()
			return nil, err
		}
	}
	return &SMTPClient{Client: nClient}, nil
}
