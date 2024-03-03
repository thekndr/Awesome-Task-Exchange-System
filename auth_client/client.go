package auth_client

import (
	"encoding/json"
	"fmt"
	"github.com/cenkalti/backoff/v4"
	"github.com/golang-jwt/jwt/v4"
	"io"
	"log"
	"net/http"
)

// Used both by auth clients as well as `auth` itself (for the sake of simplicity)
func VerifyToken(authPort int, token string) error {
	return backoff.Retry(func() error {
		return verifyTokenWithoutRetries(authPort, token)
	}, backoff.WithMaxRetries(backoff.NewExponentialBackOff(), 3))
}

func verifyTokenWithoutRetries(authPort int, token string) error {
	url := fmt.Sprintf(`http://localhost:%d/verify`, authPort)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf(`verification request failed: %w`, err)
	}

	q := req.URL.Query()
	q.Add("token", token)
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf(`failed to read response body: %w`, err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(`auth service failed (%d): %s`, resp.Status, respBody)
	}

	var validationResponse = struct {
		Role  string `json:"role"`
		Valid bool   `json:"valid"`
	}{}

	if err := json.Unmarshal(respBody, &validationResponse); err != nil {
		return fmt.Errorf(`validation response decoding failed: %w`, err)
	}

	if !validationResponse.Valid {
		return fmt.Errorf(`token is invalid`)
	}

	log.Printf("token verified, role=%s", validationResponse.Role)

	return nil
}

func WithTokenVerification(authPort int, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.URL.Query().Get("token")
		if err := VerifyToken(authPort, tokenString); err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			log.Printf(`token verification failed, err=%s`, err)
			return
		}

		handler(w, r)
	}
}

func GetClaimWithoutVerification[T any](tokenString string, key string) (T, error) {
	var null T
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return null, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		return claims[key].(T), nil
	}

	return null, fmt.Errorf(`no claim key=%s is found`, key)
}
