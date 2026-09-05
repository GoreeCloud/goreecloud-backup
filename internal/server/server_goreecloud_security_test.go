package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	"github.com/kopia/kopia/internal/apiclient"
	"github.com/kopia/kopia/internal/auth"
	"github.com/kopia/kopia/internal/clock"
)

func TestGoreeCloudAuthenticationCookieSecurity(t *testing.T) {
	const (
		username = "goreecloud-admin"
		password = "test-password"
	)

	s := &Server{
		authenticator:        auth.AuthenticateSingleUser(username, password),
		authCookieSigningKey: []byte("goreecloud-auth-cookie-test-signing-key"),
		options: Options{
			LogRequests: true,
		},
	}

	req := httptest.NewRequest(http.MethodGet, "https://backup.goreecloud.test/api/v1/repo/status", http.NoBody)
	req.SetBasicAuth(username, password)

	rr := httptest.NewRecorder()
	require.True(t, s.isAuthenticated(requestContext{w: rr, req: req, srv: s}))

	var authCookie *http.Cookie

	for _, cookie := range rr.Result().Cookies() {
		if cookie.Name == kopiaAuthCookie {
			authCookie = cookie

			break
		}
	}

	require.NotNil(t, authCookie, "successful authentication must issue the bounded optimization cookie")
	require.True(t, authCookie.HttpOnly, "authentication cookie must not be readable by browser scripts")
	require.True(t, authCookie.Secure, "authentication cookie must only be sent over secure transport")
	require.Equal(t, http.SameSiteStrictMode, authCookie.SameSite, "authentication cookie must use strict same-site isolation")
	require.Equal(t, "/", authCookie.Path)
	require.False(t, authCookie.Expires.IsZero())
	require.True(t, s.isAuthCookieValid(username, authCookie.Value))
	require.False(t, s.isAuthCookieValid("different-user", authCookie.Value), "cookie subject must be bound to the authenticated user")
}

func TestGoreeCloudAuthenticationCookieRejectsWrongIssuerAndAudience(t *testing.T) {
	s := &Server{authCookieSigningKey: []byte("goreecloud-auth-cookie-test-signing-key")}
	now := clock.Now()

	makeToken := func(issuer string, audience jwt.ClaimStrings) string {
		t.Helper()

		token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.RegisteredClaims{
			Subject:   "goreecloud-admin",
			NotBefore: jwt.NewNumericDate(now.Add(-kopiaAuthCookieTTL)),
			ExpiresAt: jwt.NewNumericDate(now.Add(kopiaAuthCookieTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			Audience:  audience,
			Issuer:    issuer,
		}).SignedString(s.authCookieSigningKey)
		require.NoError(t, err)

		return token
	}

	require.False(t, s.isAuthCookieValid("goreecloud-admin", makeToken("unexpected-issuer", jwt.ClaimStrings{kopiaAuthCookieAudience})))
	require.False(t, s.isAuthCookieValid("goreecloud-admin", makeToken(kopiaAuthCookieIssuer, jwt.ClaimStrings{"unexpected-audience"})))
}

func TestGoreeCloudAuthenticationDeniesInvalidCredentials(t *testing.T) {
	s := &Server{
		authenticator:        auth.AuthenticateSingleUser("goreecloud-admin", "correct-password"),
		authCookieSigningKey: []byte("goreecloud-auth-cookie-test-signing-key"),
	}

	req := httptest.NewRequest(http.MethodGet, "https://backup.goreecloud.test/api/v1/repo/status", http.NoBody)
	req.SetBasicAuth("submitted-user-must-not-be-logged", "wrong-password")

	rr := httptest.NewRecorder()
	require.False(t, s.isAuthenticated(requestContext{w: rr, req: req, srv: s}))
	require.Equal(t, http.StatusUnauthorized, rr.Code)
	require.Equal(t, `Basic realm="Kopia"`, rr.Header().Get("WWW-Authenticate"))
}

func TestGoreeCloudRequestIntegrityValidatesCSRFTokenWithoutExposingSecrets(t *testing.T) {
	const sessionID = "goreecloud-session-secret-test-value"

	s := &Server{authCookieSigningKey: []byte("goreecloud-csrf-test-signing-key")}

	validRequest := httptest.NewRequest(http.MethodPost, "https://backup.goreecloud.test/api/v1/sources", http.NoBody)
	validRequest.AddCookie(&http.Cookie{Name: kopiaSessionCookie, Value: sessionID})
	validRequest.Header.Set(apiclient.CSRFTokenHeader, s.generateCSRFToken(sessionID))
	require.True(t, s.validateCSRFToken(validRequest))

	invalidRequest := httptest.NewRequest(http.MethodPost, "https://backup.goreecloud.test/api/v1/sources", http.NoBody)
	invalidRequest.AddCookie(&http.Cookie{Name: kopiaSessionCookie, Value: sessionID})
	invalidRequest.Header.Set(apiclient.CSRFTokenHeader, "submitted-csrf-secret-test-value")
	require.False(t, s.validateCSRFToken(invalidRequest))

	missingTokenRequest := httptest.NewRequest(http.MethodPost, "https://backup.goreecloud.test/api/v1/sources", http.NoBody)
	missingTokenRequest.AddCookie(&http.Cookie{Name: kopiaSessionCookie, Value: sessionID})
	require.False(t, s.validateCSRFToken(missingTokenRequest))

	missingSessionRequest := httptest.NewRequest(http.MethodPost, "https://backup.goreecloud.test/api/v1/sources", http.NoBody)
	missingSessionRequest.Header.Set(apiclient.CSRFTokenHeader, "submitted-csrf-secret-test-value")
	require.False(t, s.validateCSRFToken(missingSessionRequest))
}

func TestGoreeCloudUISessionCookieSecurity(t *testing.T) {
	s := &Server{authCookieSigningKey: []byte("goreecloud-ui-session-test-signing-key")}
	router := mux.NewRouter()
	uiFS := http.FS(fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<html><head></head><body>GoreeCloud Backup</body></html>")},
	})

	s.ServeStaticFiles(router, uiFS)

	req := httptest.NewRequest(http.MethodGet, "https://backup.goreecloud.test/", http.NoBody)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var sessionCookie *http.Cookie
	for _, cookie := range rr.Result().Cookies() {
		if cookie.Name == kopiaSessionCookie {
			sessionCookie = cookie
			break
		}
	}

	require.NotNil(t, sessionCookie, "UI bootstrap must issue a CSRF-bound session cookie")
	require.NotEmpty(t, sessionCookie.Value)
	require.Equal(t, "/", sessionCookie.Path)
	require.True(t, sessionCookie.HttpOnly, "UI session cookie must not be readable by browser scripts")
	require.True(t, sessionCookie.Secure, "UI session cookie must only be sent over secure transport")
	require.Equal(t, http.SameSiteStrictMode, sessionCookie.SameSite, "UI session cookie must use strict same-site isolation")
}
