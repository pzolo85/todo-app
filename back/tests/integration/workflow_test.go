// run integration tests
package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/pzolo85/todo-app/back/internal/auth"
	"github.com/pzolo85/todo-app/back/internal/config"
	"github.com/pzolo85/todo-app/back/internal/mail"
	"github.com/pzolo85/todo-app/back/internal/user"
	"github.com/stretchr/testify/assert"
)

var AdminToken string
var cfg *config.Config
var (
	basePath  = "/api/v1"
	adminPath = "/admin"
	mailPath  = "/mail"
	userPath  = "/user"
	host      string
)

func loadConfig() error {
	if cfg != nil {
		return nil
	}
	var err error
	if cfg, err = config.Load(); err != nil {
		return err
	}

	AdminToken = os.Getenv("I_ADMIN_TOKEN")
	if AdminToken == "" {
		return fmt.Errorf("admin token missing")
	}

	host = fmt.Sprintf("http://%s:%d", cfg.Address, cfg.Port)

	return nil
}

func setAdmin(req *http.Request) {
	req.Header.Add(auth.AuthHeader, AdminToken)
}
func setJSON(req *http.Request) {
	req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
}

func Test_User(t *testing.T) {
	assert.Nil(t, loadConfig())
	email := "john@test.com"

	t.Run("create user", func(t *testing.T) {
		reqBody := user.UserCreateRequest{
			Email:      email,
			Salt:       "abc123",
			HashedPass: "abc123",
		}

		reqBodyBytes, err := json.Marshal(reqBody)
		assert.Nil(t, err)
		reqBodyReader := bytes.NewReader(reqBodyBytes)

		req, err := http.NewRequest(
			http.MethodPost,
			host+basePath+userPath+"/create",
			reqBodyReader,
		)
		setJSON(req)

		res, err := http.DefaultClient.Do(req)
		assert.Nil(t, err)

		assert.Equal(t, res.StatusCode, http.StatusOK)

		resByte, err := io.ReadAll(res.Body)
		assert.Nil(t, err)
		defer res.Body.Close()

		var u user.User
		err = json.Unmarshal(resByte, &u)
		assert.Nil(t, err)

		assert.Equal(t, email, u.Email)
	})

	t.Run("validate email", func(t *testing.T) {
		req, err := http.NewRequest(
			http.MethodGet,
			host+basePath+adminPath+mailPath+"/list",
			nil,
		)
		assert.Nil(t, err)
		setAdmin(req)

		res, err := http.DefaultClient.Do(req)
		assert.Nil(t, err)

		resByte, err := io.ReadAll(res.Body)
		assert.Nil(t, err)
		defer res.Body.Close()

		var m mail.Mails
		err = json.Unmarshal(resByte, &m)
		assert.Nil(t, err)

		for _, m := range m.Mails {
			if m.To == email {
				r, err := http.NewRequest(http.MethodGet, m.Link, nil)
				assert.Nil(t, err)
				resp, err := http.DefaultClient.Do(r)
				assert.Equal(t, resp.StatusCode, http.StatusOK)
				break
			}
		}

	})

}
