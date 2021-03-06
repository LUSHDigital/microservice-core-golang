package auth_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/LUSHDigital/microservice-core-golang/keys"

	"github.com/LUSHDigital/microservice-core-golang/auth"
	"github.com/LUSHDigital/microservice-core-golang/response"
	jwt "github.com/dgrijalva/jwt-go"
)

func TestHandlerValidateJWT(t *testing.T) {
	defaultConsumer := auth.Consumer{
		ID:     5,
		Grants: []string{"test.grant"},
	}

	cases := []struct {
		name                 string
		broker               keys.RSAPublicKeyCopierRenewer
		claims               auth.Claims
		expectedStatusCode   int
		expectedErrorMessage string
	}{
		{
			name:   "token is good",
			broker: keys.MockRSAPublicKey(*correctPK),
			claims: auth.Claims{
				StandardClaims: jwt.StandardClaims{
					IssuedAt:  time.Now().Add(-2 * time.Hour).Unix(),
					NotBefore: time.Now().Add(-1 * time.Hour).Unix(),
					ExpiresAt: time.Now().Add(1 * time.Hour).Unix(),
				},
				Consumer: defaultConsumer,
			},
			expectedStatusCode:   http.StatusOK,
			expectedErrorMessage: "",
		},
		{
			name:   "token has expired",
			broker: keys.MockRSAPublicKey(*correctPK),
			claims: auth.Claims{
				StandardClaims: jwt.StandardClaims{
					IssuedAt:  time.Now().Add(-2 * time.Hour).Unix(),
					NotBefore: time.Now().Add(-1 * time.Hour).Unix(),
					ExpiresAt: time.Now().Add(-1 * time.Minute).Unix(),
				},
				Consumer: defaultConsumer,
			},
			expectedStatusCode:   http.StatusUnauthorized,
			expectedErrorMessage: "token is expired by 1m0s",
		},
		{
			name:   "token is not ready yet",
			broker: keys.MockRSAPublicKey(*correctPK),
			claims: auth.Claims{
				StandardClaims: jwt.StandardClaims{
					IssuedAt:  time.Now().Add(-2 * time.Hour).Unix(),
					NotBefore: time.Now().Add(1 * time.Minute).Unix(),
					ExpiresAt: time.Now().Add(1 * time.Hour).Unix(),
				},
				Consumer: defaultConsumer,
			},
			expectedStatusCode:   http.StatusUnauthorized,
			expectedErrorMessage: "token is not valid yet",
		},
		{
			name:   "issuedAt is in the future",
			broker: keys.MockRSAPublicKey(*correctPK),
			claims: auth.Claims{
				StandardClaims: jwt.StandardClaims{
					IssuedAt:  time.Now().Add(1 * time.Hour).Unix(),
					NotBefore: time.Now().Add(1 * time.Minute).Unix(),
					ExpiresAt: time.Now().Add(1 * time.Hour).Unix(),
				},
				Consumer: defaultConsumer,
			},
			expectedStatusCode:   http.StatusUnauthorized,
			expectedErrorMessage: "token is not valid yet",
		},
		{
			name:   "token not signed with matching key",
			broker: keys.MockRSAPublicKey(*incorrectPK),
			claims: auth.Claims{
				StandardClaims: jwt.StandardClaims{
					IssuedAt:  time.Now().Add(-2 * time.Hour).Unix(),
					NotBefore: time.Now().Add(-1 * time.Hour).Unix(),
					ExpiresAt: time.Now().Add(1 * time.Hour).Unix(),
				},
				Consumer: defaultConsumer,
			},
			expectedStatusCode:   http.StatusUnauthorized,
			expectedErrorMessage: "token signature invalid: crypto/rsa: verification error",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			token, err := issuer.IssueWithClaims(c.claims)
			if err != nil {
				t.Fatal(err)
			}
			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Add("Authorization", "Bearer "+token)

			recorder := httptest.NewRecorder()
			handler := auth.HandlerValidateJWT(c.broker, func(w http.ResponseWriter, r *http.Request) {
				consumer := auth.ConsumerFromContext(r.Context())
				response.New(http.StatusOK, "", &response.Data{Type: "consumer", Content: consumer}).WriteTo(w)
			})
			handler.ServeHTTP(recorder, req)

			equals(t, c.expectedStatusCode, recorder.Code)

			var body response.Response
			if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}

			equals(t, body.Code, c.expectedStatusCode)
			equals(t, body.Message, c.expectedErrorMessage)

			if c.expectedStatusCode == http.StatusOK {
				var consumer auth.Consumer
				err := body.ExtractData("consumer", &consumer)
				if err != nil {
					t.Fatal(err)
				}
				equals(t, c.claims.Consumer.ID, consumer.ID)
			}
		})
	}
}
