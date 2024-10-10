package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

// TokenResponse is a proto message that contains a token.
// It typically is a response from a login service.
type TokenResponse interface {
	proto.Message
	GetToken() string
}

// Handler is an interface that can be used to handle token forwarding.
// It can be used to forward tokens from grpc messages to cookies and vice versa.
// T is the token response type that creates the session and contains the token to be forwarded as a cookie.
// U is the message that closes the session.
type Handler interface {
	CookieToAuth(ctx context.Context) context.Context
	ForwardTokenAsCookie(ctx context.Context, w http.ResponseWriter, p proto.Message) error
	ForwardCookieAsToken(ctx context.Context, request *http.Request) metadata.MD
	SendAuthCookie(ctx context.Context, token string) error
}

func NewHandler(sessionName string, sessions *sessions.CookieStore) Handler {
	return &handler{sessionName: sessionName, sessions: sessions}
}

type handler struct {
	sessionName string
	sessions    *sessions.CookieStore
}

// ForwardTokenAsCookie forwards the token from the message to the cookie store.
func (h *handler) ForwardTokenAsCookie(_ context.Context, w http.ResponseWriter, p proto.Message) error {
	switch m := any(p).(type) {
	case TokenResponse:
		sess := h.newSession(h.sessionName)
		sess.Values["token"] = m.GetToken()
		// request is not available in the context, but session.Save does not actually use it
		return h.sessions.Save(nil, w, sess)
	default:
		sess := h.newSession(h.sessionName)
		sess.Options.MaxAge = -1
		return h.sessions.Save(nil, w, sess)
	}
}

// ForwardCookieAsToken forwards the token from the cookie store to the grpc metadata.
func (h *handler) ForwardCookieAsToken(_ context.Context, request *http.Request) metadata.MD {
	sess, err := h.sessions.Get(request, h.sessionName)
	if err != nil {
		return nil
	}
	tk, ok := sess.Values["token"]
	if !ok {
		return nil
	}
	return metadata.Pairs("authorization", fmt.Sprintf("Bearer %v", tk))
}

// SendAuthCookie sets the cookie in the response.
func (h *handler) SendAuthCookie(ctx context.Context, token string) error {
	sess := h.newSession(h.sessionName)
	sess.Values["token"] = token
	if token == "" {
		sess.Options.MaxAge = -1
	}
	encoded, err := securecookie.EncodeMulti(sess.Name(), sess.Values, h.sessions.Codecs...)
	if err != nil {
		return err
	}
	return grpc.SetHeader(ctx, metadata.Pairs("set-cookie", sessions.NewCookie(sess.Name(), encoded, sess.Options).String()))
}

func (h *handler) newSession(name string) *sessions.Session {
	session := sessions.NewSession(h.sessions, name)
	opts := *h.sessions.Options
	session.Options = &opts
	session.IsNew = true
	return session
}

// CookieToAuth forwards the cookie authorization to the grpc metadata.
func (h *handler) CookieToAuth(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx
	}
	// if we have an authorization header, we don't need to check for cookies here
	if len(md.Get("authorization")) > 0 {
		return ctx
	}
	// no other way to easily parse cookies
	header := http.Header{}
	for _, v := range md.Get("cookie") {
		header.Add("Cookie", v)
	}
	md.Delete("cookie")
	sess, err := h.sessions.Get(&http.Request{Header: header}, h.sessionName)
	if err != nil {
		return ctx
	}
	if v, ok := sess.Values["token"]; ok {
		md["authorization"] = []string{fmt.Sprintf("Bearer %s", v)}
	}
	return metadata.NewIncomingContext(ctx, md)
}
