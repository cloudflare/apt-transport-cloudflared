package access

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/cloudflare/apt-transport-cloudflared/apt/exec"
)

// Transport takes a Token and applies it to any requests sent using it.
type Transport struct {
	authToken Token
	parent    http.RoundTripper
}

// NewTransport returns a new AccessRountTripper set to use the given
// token and parent round-tripper.
func NewTransport(token Token, rt http.RoundTripper) *Transport {
	if rt == nil {
		rt = http.DefaultTransport
	}

	return &Transport{
		authToken: token,
		parent:    rt,
	}
}

// RoundTrip applies the token headers to the request and gets a response.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.authToken.ModifyRequest(req)

	return t.parent.RoundTrip(req)
}

// Token is the interface to both types of token.
type Token interface {
	ModifyRequest(r *http.Request)
}

// GetToken attempts to get a token for the given uri.
//
// This function first attempts to load a service token for the requested URI,
// then attempts to load a user JWT using cloudflared if no service token was
// found.
//
// The writer argument is used to redirect os.Stderr from the subprocess (if
// one is spawned). If you want to silence the output from os.Stderr, use
// ioutil.Discard as the writer. If the writer is nil, it is implicitly
// converted to ioutil.Discard.
func GetToken(ctx context.Context, uri *url.URL, servicetokendir string,
	usecloudflared bool, w io.Writer) (Token, error) {
	// Attempt to load a service token
	if servicetokendir != "" {
		token, err := FindServiceToken(servicetokendir, uri.Host)
		if err == nil && token != nil {
			return token, nil
		}
	}

	// Attempt to get the user token
	return FindUserToken(ctx, uri, usecloudflared, w)
}

// ServiceToken is a Cloudflare Access token used for services which need
// long-lived access.
type ServiceToken struct {
	// ID is the service token ID. This is used to identify the service which
	// is being authenticated.
	ID string

	// Secret is the Client-Secret value for the service token.
	Secret string
}

// ParseServiceToken converts a compound service token consisting of a
// Client-ID and a Client-Secret into a ServiceToken struct.
//
// This function expects the client data to be stored on two lines as:
//   ${CLIENT_ID}
//   ${CLIENT_SECRET}
// Whitespace in the Client-ID and Client-Secret will be stripped. The
// Client-ID must be in the form of:
//   ${ID}.${HOST}
// where ${HOST} is the hostname being connected to.
func ParseServiceToken(data string) (*ServiceToken, error) {
	// Trim off trailing newlines, then split on newline
	parts := strings.Split(strings.TrimSpace(data), "\n")

	// We need exactly 2 parts - otherwise error
	if len(parts) != 2 {
		return nil, fmt.Errorf("parse expected two lines of input, got %d", len(parts))
	}

	return &ServiceToken{
		ID:     strings.TrimSpace(parts[0]),
		Secret: strings.TrimSpace(parts[1]),
	}, nil
}

// LoadServiceToken takes the given file path and parses a ServiceToken from
// the file contents.
//
// If the file does not exist, then this function returns an error. See the
// ParseServiceToken() function for more details on how service tokens are
// parsed.
func LoadServiceToken(filepath string) (*ServiceToken, error) {
	fdata, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	return ParseServiceToken(string(fdata))
}

// FindServiceToken takes the given directory and path and attempts to load a
// service token for the given host.
func FindServiceToken(directory, host string) (*ServiceToken, error) {
	return LoadServiceToken(path.Join(directory, host+"-Service-Token"))
}

// ModifyRequest sets the request headers to the given token values.
func (st *ServiceToken) ModifyRequest(req *http.Request) {
	req.Header.Set("Cf-Access-Client-Id", st.ID)
	req.Header.Set("Cf-Access-Client-Secret", st.Secret)
}

// UserToken represents a user token for the given service.
type UserToken struct {
	// JWT is the content of the user token.
	JWT string
}

func findTokenCloudflared(ctx context.Context, uri *url.URL, w io.Writer) (*UserToken, error) {
	baseuri := uri.Scheme + "://" + uri.Host

	var cmdLogin, cmdToken []string
	prog := "cloudflared"
	sudoUser := os.Getenv("SUDO_USER")
	if sudoUser != "" {
		prog = "su"
		cmdLogin = []string{sudoUser, "-c", "cloudflared access login " + baseuri}
		cmdToken = []string{sudoUser, "-c", "cloudflared access token --app " + baseuri}
	} else {
		cmdLogin = []string{"access", "login", baseuri}
		cmdToken = []string{"access", "token", "--app", baseuri}
	}

	login := exec.CommandContext(ctx, prog, cmdLogin...)
	login.Stderr = w
	err := login.Run()
	if err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, prog, cmdToken...)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	token := strings.TrimSpace(string(output))
	if token == "" {
		return nil, errors.New("no output from `cloudflared access token`")
	}

	if strings.HasPrefix(token, "Unable") {
		return nil, errors.New("bad output from `cloudflared access token`: unable to get token")
	}

	return &UserToken{token}, nil
}

func findToken(ctx context.Context, uri *url.URL, w io.Writer) (*UserToken, error) {
	// TODO: Use cloudflared library directly
	return findTokenCloudflared(ctx, uri, w)
}

// FindUserToken attempts to fetch a user token for the given URI.
func FindUserToken(ctx context.Context, uri *url.URL, cloudflared bool, w io.Writer) (*UserToken, error) {
	if w == nil {
		w = ioutil.Discard
	}

	if cloudflared {
		return findTokenCloudflared(ctx, uri, w)
	}
	return findToken(ctx, uri, w)
}

func (ut *UserToken) ModifyRequest(req *http.Request) {
	req.Header.Set("Cf-Access-Token", ut.JWT)
}
