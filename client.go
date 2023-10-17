package nmail

import (
	"context"
	"crypto/tls"
	"errors"
	"github.com/smartwalle/npool"
	"net/mail"
	"net/smtp"
	"time"
)

type Dialer func(ctx context.Context, username, password, host, port string, tls *tls.Config) (*SMTPClient, error)

type SMTPClient struct {
	*smtp.Client
}

func (c *SMTPClient) Close() error {
	return c.Client.Quit()
}

type Option func(c *Client)

func WithDialer(dialer Dialer) Option {
	return func(c *Client) {
		c.dialer = dialer
	}
}

func WithMaxIdle(idle int) Option {
	return func(c *Client) {
		c.opts = append(c.opts, npool.WithMaxIdle(idle))
	}
}

func WithMaxActive(active int) Option {
	return func(c *Client) {
		c.opts = append(c.opts, npool.WithMaxActive(active))
	}
}

func WithIdleTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.opts = append(c.opts, npool.WithIdleTimeout(timeout))
	}
}

func WithMaxLifetime(lifetime time.Duration) Option {
	return func(c *Client) {
		c.opts = append(c.opts, npool.WithMaxLifetime(lifetime))
	}
}

func WithTLSConfig(config *tls.Config) Option {
	return func(c *Client) {
		c.tlsConfig = config
	}
}

type Client struct {
	username  string
	password  string
	host      string
	port      string
	tlsConfig *tls.Config
	dialer    Dialer

	opts []npool.Option
	pool *npool.Pool[*SMTPClient]
}

func NewClient(username, password, host, port string, opts ...Option) *Client {
	var nClient = &Client{}
	nClient.username = username
	nClient.password = password
	nClient.host = host
	nClient.port = port

	for _, opt := range opts {
		if opt != nil {
			opt(nClient)
		}
	}
	nClient.opts = append(nClient.opts, npool.WithWait(true))

	if nClient.tlsConfig == nil {
		nClient.tlsConfig = &tls.Config{
			ServerName:         host,
			InsecureSkipVerify: true,
		}
	}
	if nClient.dialer == nil {
		nClient.dialer = DefaultDialer
	}
	nClient.pool = npool.New[*SMTPClient](
		func(ctx context.Context) (*SMTPClient, error) {
			return nClient.dialer(ctx, nClient.username, nClient.password, nClient.host, nClient.port, nClient.tlsConfig)
		},
		nClient.opts...,
	)
	return nClient
}

func (c *Client) Send(message *Message) error {
	to := make([]string, 0, len(message.To)+len(message.Cc)+len(message.Bcc))
	to = append(append(append(to, message.To...), message.Cc...), message.Bcc...)
	for i := 0; i < len(to); i++ {
		addr, err := mail.ParseAddress(to[i])
		if err != nil {
			return err
		}
		to[i] = addr.Address
	}

	if message.From == "" || len(to) == 0 {
		return errors.New("must specify at least one From address and one To address")
	}
	sender, err := message.parseSender()
	if err != nil {
		return err
	}
	raw, err := message.Bytes()
	if err != nil {
		return err
	}

	conn, err := c.pool.Get(context.Background())
	if err != nil {
		return err
	}
	defer conn.Release()

	if err = conn.Element().Mail(sender); err != nil {
		return err
	}
	for _, addr := range to {
		if err = conn.Element().Rcpt(addr); err != nil {
			return err
		}
	}
	w, err := conn.Element().Data()
	if err != nil {
		return err
	}
	_, err = w.Write(raw)
	if err != nil {
		return err
	}
	return w.Close()
}

func (c *Client) Close() error {
	return c.pool.Close()
}
