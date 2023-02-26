package tests

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gitamped/seed/auth"
	"github.com/gitamped/seed/keystore"
	"github.com/gitamped/seed/mid"
	"github.com/gitamped/seed/server"
	"github.com/golang-jwt/jwt/v4"
)

// Success and failure markers.
const (
	Success = "\u2713"
	Failed  = "\u2717"
)

type ServerTest struct {
	app          http.Handler
	adminToken   string
	userToken    string
	invalidToken string
}

func Test_ServerWithRegisteredEndpoints(t *testing.T) {
	// Configure seed server
	validAuth := GetAuth()
	authMid := mid.AuthMiddleware(validAuth)
	mw := append([]mid.Middleware{authMid}, mid.CommonMiddleware...)
	s := server.NewServer(mw)
	mid.MultipleMiddleware(s.DefaultHandler, mid.CommonMiddleware...)

	// Register GreeterServicer
	gs := NewSecretGreeterServicer()
	gs.Register(s)

	// Create tokens for tests
	claims := auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "seed project",
			Subject:   "5cf37266-3473-4006-984f-9325122678b7",
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		},
		Roles: []string{auth.RoleUser},
	}

	userToken, _ := validAuth.GenerateToken(claims)
	invalidAuth := GetAuth()
	invalidToken, _ := invalidAuth.GenerateToken(claims)

	tests := ServerTest{
		app:          s,
		userToken:    userToken,
		invalidToken: invalidToken,
	}

	t.Run("Server", tests.Test_Server)
}

func GetAuth() *auth.Auth {
	const keyID = "54bb2165-71e1-41a6-af3e-7da4a0e1e2c1"
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	ks := keystore.New()
	ks.Add(privateKey, keyID)
	a, _ := auth.New(keyID, ks)
	return a
}

func (st *ServerTest) Test_Server(t *testing.T) {
	t.Log("Given the need to call the rpc over http server endpoints")
	{

		ttable := []struct {
			TestTitle          string
			ExpectedStatusCode int
			ExpectedError      bool
			ExpectedErrorText  string
			Token              string
			URL                string
			RequestData        string
		}{
			// test valid alias
			{
				TestTitle:          "When passed a valid alias",
				ExpectedStatusCode: 200,
				ExpectedError:      false,
				Token:              st.userToken,
				URL:                "/v1/GreeterService.SecretGreet",
				RequestData:        `{"alias": "Seed Client"}`,
			},
			// test valid alias
			{
				TestTitle:          "When passed a invalid alias",
				ExpectedStatusCode: 200,
				ExpectedError:      true,
				ExpectedErrorText:  `validating data: [{"field":"alias","error":"alias must be at least 1 character in length"}]`,
				Token:              st.userToken,
				URL:                "/v1/GreeterService.SecretGreet",
				RequestData:        `{"alias": ""}`,
			},
			// test invalid user
			{
				TestTitle:          "When unauthenticated user attempts action",
				ExpectedStatusCode: 401,
				ExpectedError:      true,
				ExpectedErrorText:  "attempted action is not allowed",
				Token:              st.invalidToken,
				URL:                "/v1/GreeterService.SecretGreet",
				RequestData:        `{"name": "Suspicious Seed Client"}`,
			},
		}

		for i, td := range ttable {

			testID := i
			t.Logf("\tTest %d:\t%s", testID, td.TestTitle)
			{

				r := httptest.NewRequest(http.MethodPost, td.URL, bytes.NewBuffer([]byte(td.RequestData)))

				w := httptest.NewRecorder()
				r.Header.Set("Authorization", "Bearer "+td.Token)
				st.app.ServeHTTP(w, r)

				if w.Code != td.ExpectedStatusCode {
					t.Fatalf("\t%s\tTest %d:\tShould receive a status code of %d for the response : %v", Failed, testID, td.ExpectedStatusCode, w.Code)
				}
				t.Logf("\t%s\tTest %d:\tShould receive a status code of %d for the response.", Success, testID, td.ExpectedStatusCode)

				if w.Code != 200 {
					continue
				}

				var got SecretGreetResponse
				if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to unmarshal the response : %v", Failed, testID, err)
				}

				if td.ExpectedError {
					if got.Error != td.ExpectedErrorText {
						t.Fatalf("\t%s\tTest %d: Should return the error expected: %s: actual: %s", Failed, testID, td.ExpectedErrorText, got.Error)
					}
					t.Logf("\t%s\tTest: Should return the error.", Success)
				} else {
					if len(got.SecretGreeting) < 1 {
						t.Fatalf("\t%s\tTest %d: Should return the expected greeting.", Failed, testID)
					}
					t.Logf("\t%s\tTest %d: Should return the expected greeting.", Success, testID)
				}
			}
		}
	}

}
