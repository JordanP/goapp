package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jordanp/goapp/app"
	"github.com/jordanp/goapp/entity"
	"github.com/jordanp/goapp/pkg/auth"
	"github.com/jordanp/goapp/pkg/handlers"
	"github.com/jordanp/goapp/pkg/log"
	"github.com/stretchr/testify/suite"
)

var (
	httpClient = http.Client{Timeout: 2000 * time.Millisecond}
	fixtures   = struct {
		u []entity.User
	}{
		u: []entity.User{
			{Login: "admin", Password: "$2y$10$CpVqJK/usJ8K8musmkaM1u3K7agJ0m/YOGQPLuwiBZ1M15cDHbkcu", Email: "admin@goapp", Role: "admin"},
			{Login: "user", Password: "$2y$10$CpVqJK/usJ8K8musmkaM1u3K7agJ0m/YOGQPLuwiBZ1M15cDHbkcu", Email: "user@goapp", Role: "user"},
			{Login: "user1Company1", Password: "$2y$10$CpVqJK/usJ8K8musmkaM1u3K7agJ0m/YOGQPLuwiBZ1M15cDHbkcu", Email: "user1@company1", Role: "user"},
			{Login: "user2Company1", Password: "$2y$10$CpVqJK/usJ8K8musmkaM1u3K7agJ0m/YOGQPLuwiBZ1M15cDHbkcu", Email: "user2@company1", Role: "user"},
		},
	}
)

type ApplicationTestSuite struct {
	suite.Suite
	app        *app.Application
	testServer *httptest.Server

	fixtures struct {
		u []entity.User
		c []entity.Company
	}
}

func TestMainApplication(t *testing.T) {
	if os.Getenv("CIRCLECI") == "true" { // CircleCI workers are damn slow
		httpClient.Timeout = 5 * time.Second
	}
	suite.Run(t, new(ApplicationTestSuite))
}

func (t *ApplicationTestSuite) SetupSuite() {
	log := log.New("mygoapp", "test", log.ErrorLevel)
	app, err := app.NewApplication(log, app.NewConfig(os.Getenv("SECRET_KEY"), os.Getenv("SQL_DSN")))
	t.Require().NoError(err)
	t.app = app
	t.testServer = httptest.NewServer(t.app.Routes())
}

func (t *ApplicationTestSuite) SetupTest() {
	t.Require().NoError(t.app.UserStore.DeleteAll())
	t.Require().NoError(t.app.CompanyStore.DeleteAll())
	ctx := context.Background()

	for _, u := range fixtures.u {
		_, err := t.app.UserStore.Add(ctx, u.Login, u.Password, u.Email, u.Role)
		t.Require().NoError(err)
	}
	t.fixtures.u, _ = t.app.UserStore.GetAll(context.Background())

	c, err := t.app.CompanyStore.Add(ctx, "company1", []uuid.UUID{t.fixtures.u[2].ID, t.fixtures.u[3].ID})
	t.Require().NoError(err)
	t.fixtures.c = []entity.Company{c}

	c, err = t.app.CompanyStore.Add(ctx, "company2", nil)
	t.Require().NoError(err)
	t.fixtures.c = append(t.fixtures.c, c)
}

func (t *ApplicationTestSuite) TearDownSuite() {
	t.testServer.Close()
	t.app.Stop()
}

func (t *ApplicationTestSuite) TestStatus() {
	var status handlers.StatusResponse
	t.get("/status", nil, http.StatusOK, &status)
	t.Require().WithinDuration(time.Now(), time.Now().Add(status.Uptime.Duration), 60*time.Second)

	t.options("/status", http.StatusOK)
}

func (t *ApplicationTestSuite) TestMetrics() {
	var resp []byte
	t.get("/metrics", nil, http.StatusOK, &resp)
	t.Require().Contains(string(resp), "request_duration_seconds")
	t.Require().Contains(string(resp), "go_memstats")
}

func (t *ApplicationTestSuite) TestUnauthorized() {
	t.delete("/admin/users/foo", nil, http.StatusUnauthorized, nil)
	t.get("/admin/users/all", nil, http.StatusUnauthorized, nil)
	t.post("/admin/users/new", nil, nil, http.StatusUnauthorized, nil)
	t.post("/admin/companies/new", nil, nil, http.StatusUnauthorized, nil)
	t.get("/users/me", nil, http.StatusUnauthorized, nil)
}

func (t *ApplicationTestSuite) TestCreateCompany() {
	var err app.JSONError
	company := entity.Company{}
	t.post("/admin/companies/new", t.adminHeader("ut"), company, http.StatusBadRequest, &err)
	t.Require().Contains(err.Message, "missing or empty 'name'")

	var resp entity.Company
	company.Name = "foo"
	t.post("/admin/companies/new", t.adminHeader("ut"), company, http.StatusOK, &resp)
	t.Require().Equal("foo", resp.Name)
	t.Require().Len(company.ID.String(), 36)
	t.Require().InDelta(time.Now().UnixNano(), resp.CreatedAt.UnixNano(), float64(time.Second))
}

func (t *ApplicationTestSuite) TestCreateDuplicateCompany() {
	var err app.JSONError
	company := t.fixtures.c[0]
	t.post("/admin/companies/new", t.adminHeader("ut"), company, http.StatusUnprocessableEntity, &err)
	t.Require().Equal("company '"+t.fixtures.c[0].Name+"' already exists", err.Message)
}

func (t *ApplicationTestSuite) TestCreateCompanyWithInvalidUser() {
	var err app.JSONError
	notfoundID := uuid.New()
	company := entity.Company{Name: "toto", Users: []entity.User{{ID: notfoundID}}}
	t.post("/admin/companies/new", t.adminHeader("ut"), company, http.StatusUnprocessableEntity, &err)
	t.Require().Equal("user '"+notfoundID.String()+"' not found", err.Message)
}

func (t *ApplicationTestSuite) TestGetCompany() {
	var company1 entity.Company
	t.get("/admin/companies/"+t.fixtures.c[0].ID.String(), t.adminHeader("ut"), http.StatusOK, &company1)
	t.Require().Equal(t.fixtures.c[0].ID, company1.ID)
	t.Require().Len(company1.Users, 2)

	var company2 entity.Company
	t.get("/admin/companies/"+t.fixtures.c[1].ID.String(), t.adminHeader("ut"), http.StatusOK, &company2)
	t.Require().Equal(t.fixtures.c[1].ID, company2.ID)
	t.Require().Len(company2.Users, 0)

	t.get("/admin/companies/00000000-0000-0000-0000-000000000000", t.adminHeader("ut"), http.StatusNotFound, nil)
	var err app.JSONError
	t.get("/admin/companies/malformatedUUID", t.adminHeader("ut"), http.StatusNotFound, &err)
	t.Require().Equal("company 'malformatedUUID' not found", err.Message)
}

func (t *ApplicationTestSuite) TestCreateCompanyWithDuplicateUsers() {
	var err app.JSONError
	company := entity.Company{Name: "toto", Users: []entity.User{{ID: t.fixtures.u[0].ID}, {ID: t.fixtures.u[0].ID}}}
	t.post("/admin/companies/new", t.adminHeader("ut"), company, http.StatusUnprocessableEntity, &err)
	t.Require().Equal("duplicate user '"+t.fixtures.u[0].ID.String()+"' in company", err.Message)
}

func (t *ApplicationTestSuite) TestDeleteCompany() {
	t.delete("/admin/companies/"+t.fixtures.c[0].ID.String(), t.adminHeader("ut"), http.StatusOK, nil)
	t.delete("/admin/companies/"+t.fixtures.c[0].ID.String(), t.adminHeader("ut"), http.StatusNotFound, nil)
	var resp []byte
	t.delete("/admin/companies/malformatedUUID", t.adminHeader("ut"), http.StatusNotFound, &resp)
	t.Require().Contains(string(resp), "company 'malformatedUUID' not found")
}

func (t *ApplicationTestSuite) TestListAllUsers() {
	var resp []byte
	t.get("/admin/users/all", t.userHeader(entity.User{}), http.StatusUnauthorized, &resp)
	t.Require().Equal("jwt: aud claim is invalid\n", string(resp))

	var users entity.Users
	t.get("/admin/users/all", t.adminHeader("ut"), http.StatusOK, &users)
	t.Require().Len(users.Users, len(fixtures.u))
}

func (t *ApplicationTestSuite) TestDeleteUser() {
	t.delete("/admin/users/"+t.fixtures.u[0].ID.String(), t.adminHeader("ut"), http.StatusOK, nil)
	t.delete("/admin/users/"+t.fixtures.u[0].ID.String(), t.adminHeader("ut"), http.StatusNotFound, nil)
	var resp []byte
	t.delete("/admin/users/malformatedUUID", t.adminHeader("ut"), http.StatusNotFound, &resp)
	t.Require().Contains(string(resp), "user 'malformatedUUID' not found")
}

func (t *ApplicationTestSuite) TestCreateUser() {
	user := entity.User{Login: "test", Password: "test", Email: "test"}
	var resp []byte
	t.post("/admin/users/new", t.adminHeader("ut"), user, http.StatusBadRequest, &resp)
	t.Require().Contains(string(resp), "missing or empty 'role'")

	user.Role = "user"
	t.post("/admin/users/new", t.adminHeader("ut"), user, http.StatusOK, &resp)
	t.post("/admin/users/new", t.adminHeader("ut"), user, http.StatusUnprocessableEntity, &resp)
	t.Require().Contains(string(resp), "login 'test' already exists")
}

func (t *ApplicationTestSuite) TestGetAdminToken() {
	t.post("/token/admin", nil, nil, http.StatusBadRequest, nil)
	t.post("/token/admin", nil, entity.UserCredentials{}, http.StatusBadRequest, nil)
	t.post("/token/admin", nil, entity.UserCredentials{Login: "test", Password: "test"}, http.StatusUnauthorized, nil)
	t.post("/token/admin", nil, entity.UserCredentials{Login: "admin", Password: "wrongpass"}, http.StatusUnauthorized, nil)

	var resp []byte
	t.post("/token/admin", nil, entity.UserCredentials{Login: "user", Password: "admin"}, http.StatusForbidden, &resp)
	t.Require().Equal(`{"message":"you don't have the admin role"}`+"\n", string(resp))

	var token entity.Token
	t.post("/token/admin", nil, entity.UserCredentials{Login: "admin", Password: "admin"}, http.StatusOK, &token)
	t.Require().True(len(token.Token) >= 30)
}

func (t *ApplicationTestSuite) TestMe() {
	var u entity.User
	t.get("/users/me", t.userHeader(entity.User{Login: "foobar"}), http.StatusOK, &u)
	t.Require().Equal("foobar", u.Login)
}

func (t *ApplicationTestSuite) adminHeader(login string) map[string]string {
	adminToken, err := t.app.TokenManager.GenerateAdminToken(auth.AdminUser{Login: login})
	t.Require().NoError(err)
	return map[string]string{"Authorization": "Bearer " + adminToken}
}

func (t *ApplicationTestSuite) userHeader(user entity.User) map[string]string {
	accessToken, err := t.app.TokenManager.GenerateAccessToken(auth.NewUser(user.Login, user.Email, user.Role))
	t.Require().NoError(err)
	return map[string]string{"Authorization": "Bearer " + accessToken}
}

func (t *ApplicationTestSuite) options(path string, expectedStatusCode int) {
	headers := map[string]string{"Origin": "http://test.com", "Access-control-request-method": "POST"}
	t.Require().NoError(t.doRequest("OPTIONS", path, headers, nil, expectedStatusCode, nil))
}

func (t *ApplicationTestSuite) delete(path string, headers map[string]string, expectedStatusCode int, result interface{}) {
	t.Require().NoError(t.doRequest(http.MethodDelete, path, headers, nil, expectedStatusCode, result))
}

func (t *ApplicationTestSuite) get(path string, headers map[string]string, expectedStatusCode int, result interface{}) {
	t.Require().NoError(t.doRequest(http.MethodGet, path, headers, nil, expectedStatusCode, result))
}

func (t *ApplicationTestSuite) post(path string, headers map[string]string, body interface{}, expectedStatusCode int, result interface{}) {
	t.Require().NoError(t.doRequest(http.MethodPost, path, headers, body, expectedStatusCode, result))
}

func (t *ApplicationTestSuite) doRequest(method string, path string, headers map[string]string, body interface{}, expectedStatusCode int, result interface{}) error {
	var r io.Reader
	switch v := body.(type) {
	case []byte:
		r = bytes.NewReader(v)
	case nil:
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return err
		}
		r = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, t.testServer.URL+path, r)
	if err != nil {
		return err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bodyResp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	//log.Println(string(bodyResp))

	if resp.StatusCode != expectedStatusCode {
		return fmt.Errorf("unexpected status code: got %d but expected %d. (body: '%s')", resp.StatusCode, expectedStatusCode, strings.TrimSpace(string(bodyResp)))
	}
	switch v := result.(type) {
	case *[]byte:
		l := resp.ContentLength
		if l == -1 { // The value -1 indicates that the length is unknown.
			l = 1024 * 10
		}
		v2 := make([]byte, l)
		copy(v2, bodyResp)
		*v = v2
	case nil:
	default:
		err = json.Unmarshal(bodyResp, result)
	}
	return err
}
