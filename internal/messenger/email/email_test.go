package email

import (
	"net/smtp"
	"strings"
	"testing"

	"github.com/knadh/smtppool/v2"
	"golang.org/x/oauth2"
)

// The XOAUTH2 initial response is a fixed format Google parses strictly:
// base64 of user=<mailbox>^Aauth=Bearer <token>^A^A.
func TestXOAUTH2Start(t *testing.T) {
	a := &xoauth2Auth{
		user: "xnxsimon@gmail.com",
		ts:   oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "tok123"}),
	}

	proto, resp, err := a.Start(&smtp.ServerInfo{Name: "smtp.gmail.com", TLS: true})
	if err != nil {
		t.Fatalf("Start returned an error: %v", err)
	}
	if proto != "XOAUTH2" {
		t.Fatalf("auth mechanism = %q, want XOAUTH2", proto)
	}

	want := "user=xnxsimon@gmail.com\x01auth=Bearer tok123\x01\x01"
	if string(resp) != want {
		t.Fatalf("initial response = %q, want %q", string(resp), want)
	}
}

// A server challenge means the token was refused, so the error has to carry
// the server's own message for the send log.
func TestXOAUTH2ChallengeIsAnError(t *testing.T) {
	a := &xoauth2Auth{user: "u", ts: oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "t"})}

	if _, err := a.Next(nil, false); err != nil {
		t.Fatalf("Next with more=false must not error, got %v", err)
	}
	if _, err := a.Next([]byte("invalid_grant"), true); err == nil {
		t.Fatal("Next with a challenge must return an error")
	} else if !strings.Contains(err.Error(), "invalid_grant") {
		t.Fatalf("error lost the server message: %v", err)
	}
}

func TestOAuthTokenSourceNeedsCredentials(t *testing.T) {
	if _, err := oauthTokenSource(Server{Opt: smtppool.Opt{Host: "smtp.gmail.com"}}); err == nil {
		t.Fatal("expected an error without oauth_client_id and oauth_refresh_token")
	}

	ts, err := oauthTokenSource(Server{
		Opt:               smtppool.Opt{Host: "smtp.gmail.com"},
		OAuthClientID:     "client",
		OAuthClientSecret: "secret",
		OAuthRefreshToken: "refresh",
	})
	if err != nil {
		t.Fatalf("unexpected error with full credentials: %v", err)
	}
	if ts == nil {
		t.Fatal("nil token source")
	}
}

// New must accept the xoauth2 protocol and build a pool without dialing.
func TestNewAcceptsXOAUTH2(t *testing.T) {
	e, err := New("", Server{
		Opt:               smtppool.Opt{Host: "smtp.gmail.com", Port: 587, MaxConns: 1},
		AuthProtocol:      "xoauth2",
		Username:          "xnxsimon@gmail.com",
		OAuthClientID:     "client",
		OAuthClientSecret: "secret",
		OAuthRefreshToken: "refresh",
		TLSType:           "STARTTLS",
	})
	if err != nil {
		t.Fatalf("New with xoauth2 failed: %v", err)
	}
	if e == nil {
		t.Fatal("nil emailer")
	}
}

// An unknown protocol must still be refused rather than silently sending
// unauthenticated.
func TestNewRejectsUnknownAuth(t *testing.T) {
	if _, err := New("", Server{Opt: smtppool.Opt{Host: "h", Port: 25}, AuthProtocol: "bogus"}); err == nil {
		t.Fatal("expected an error for an unknown auth protocol")
	}
}
