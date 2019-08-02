package app

import (
	"context"
	"github.com/DATA-DOG/go-txdb"
	"github.com/dgrijalva/jwt-go"
	"github.com/jmoiron/sqlx"
	auth "github.com/netology-group/ulms-auth-go"
	"github.com/rubenv/sql-migrate"
	"github.com/stretchr/testify/suite"
	"sync"
	"time"
)

var once sync.Once

func init() {
	once = sync.Once{}
}

// TestSuite is a test suite
// Usage example:
//   func TestHandlers(t *testing.T) {
//	     suite.Run(t, new(TestSuite))
//   }
type TestSuite struct {
	suite.Suite
	app        *App
	token      string
	wrongToken string
	claims     *jwt.StandardClaims
}

// SetupSuite prepares global context for tests
func (s *TestSuite) SetupSuite() {
	s.claims = &jwt.StandardClaims{
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
		Issuer:    "iam.example.com",
		Audience:  "example.com",
		Subject:   "account_id",
	}
	s.token, _ = jwt.NewWithClaims(
		jwt.GetSigningMethod("HS256"),
		s.claims,
	).SignedString([]byte("secret"))
	s.wrongToken, _ = jwt.NewWithClaims(
		jwt.GetSigningMethod("HS256"),
		&jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour).Unix(),
			Issuer:    "iam.example-wrong.com",
			Audience:  "example-wrong.com",
			Subject:   "account_id",
		},
	).SignedString([]byte("secret"))
	s.app = New("fixtures/config.yml")
	s.app.Db.MustExec("DROP SCHEMA public CASCADE")
	s.app.Db.MustExec("CREATE SCHEMA public")
	migrations := &migrate.FileMigrationSource{Dir: "../../migrations"}
	if _, err := migrate.Exec(s.app.Db.DB, "postgres", migrations, migrate.Up); err != nil {
		s.Fail(err.Error())
	}
	s.NoError(s.app.Db.Close())
	once.Do(func() { txdb.Register("txdb", s.app.Config.Db.Driver, s.app.Config.Db.DataSource) })
}

// SetupTest prepares local context for test
func (s *TestSuite) SetupTest() {
	s.app.Db = sqlx.MustOpen("txdb", s.app.Config.Db.DataSource)
	s.app.auth = &fakeAuth{Auth: s.app.auth, permission: &fakePermission{results: []bool{true}}}
}

// TearDownTest clears local context after test
func (s *TestSuite) TearDownTest() {
	s.NoError(s.app.Db.Close())
}

type fakeAuth struct {
	auth.Auth
	permission auth.Permission
}

func (auth *fakeAuth) Permission(audience string) auth.Permission {
	return auth.permission
}

type fakePermission struct {
	auth.Permission
	results    []bool
	resultFunc func(claims *jwt.StandardClaims, action auth.Action, objectValues ...string) bool
}

func (permission *fakePermission) Check(claims *jwt.StandardClaims, action auth.Action, objectValues ...string) error {
	if permission.resultFunc != nil {
		if permission.resultFunc(claims, action, objectValues...) {
			return nil
		}
	} else {
		var isAuthorized bool
		if len(permission.results) > 0 {
			isAuthorized = permission.results[0]
			if len(permission.results) > 1 {
				permission.results = permission.results[1:]
			}
		}
		if isAuthorized {
			return nil
		}
	}
	return auth.ErrorNotAuthorized
}

func (permission *fakePermission) CheckWithContext(ctx context.Context, cancel context.CancelFunc, claims *jwt.StandardClaims, action auth.Action, objectValues ...string) error {
	defer cancel()
	return permission.Check(claims, action, objectValues...)
}
