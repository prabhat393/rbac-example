package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/protsack-stephan/gin-toolkit/httperr"
	"github.com/protsack-stephan/gin-toolkit/httpmw"
)

type RBACAuthorizeFunc func(*gin.Context) (bool, error)

// RBAC implements RBAC using the provided authorizer function.
func RBAC(fn RBACAuthorizeFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		ok, err := fn(c)

		if err != nil {
			httperr.InternalServerError(c, err.Error())
			c.Abort()
			return
		}

		if !ok {
			httperr.Unauthorized(c)
			c.Abort()
			return
		}

		c.Next()
	}
}

func CasbinRBACAuthorizer(e *casbin.Enforcer) RBACAuthorizeFunc {
	return func(c *gin.Context) (bool, error) {
		var user *httpmw.CognitoUser

		if val, ok := c.Get("user"); ok && val != nil {
			user, _ = val.(*httpmw.CognitoUser)
		}

		if user == nil {
			return false, errors.New("missing user in request context")
		}
		for _, role := range user.GetGroups() {
			res, err := e.Enforce(role, c.Request.URL.Path, c.Request.Method)

			if err != nil {
				fmt.Println(err)
				return false, err
			}

			if res {
				return true, nil
			}
		}

		return false, nil
	}
}

func createHandler(endpoint string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// fmt.Printf("Allowed access to %s\n", endpoint)
		c.Status(http.StatusOK)
	}
}

func setupCasbinRBACMWUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		grp := c.Request.Header.Get("group")
		user := new(httpmw.CognitoUser)
		user.SetUsername("username")
		user.SetGroups([]string{grp})

		c.Set("user", user)
		// fmt.Println("url", c.Request.URL.Path)
		// fmt.Println("method", c.Request.Method)

		// fmt.Println("role", user.GetGroups())
	}
}

func getRBACRouter(middlewares ...gin.HandlerFunc) *gin.Engine {
	r := gin.New()

	// Setup middlewares
	for _, mw := range middlewares {
		r.Use(mw)
	}

	// exports
	r.GET("/v1/exports/download/:namespace/:project", createHandler("/v1/exports/download/:namespace/:project"))
	r.GET("/v1/exports/meta/:namespace", createHandler("/v1/exports/meta/:namespace"))
	r.GET("/v1/exports/meta/:namespace/:project", createHandler("/v1/exports/meta/:namespace/:project"))

	// diffs
	r.GET("/v1/diffs/download/:date/:namespace/:project", createHandler("/v1/diffs/download/:date/:namespace/:project"))
	r.GET("/v1/diffs/meta/:date/:namespace", createHandler("/v1/diffs/meta/:date/:namespace"))
	r.GET("/v1/diffs/meta/:date/:namespace/:project", createHandler("/v1/diffs/meta/:date/:namespace/:project"))

	// realtime
	r.GET("/v1/page-delete", createHandler("/v1/page-delete"))
	r.GET("/v1/page-update", createHandler("/v1/page-update"))
	r.GET("/v1/page-visibility", createHandler("/v1/page-visibility"))

	// on-demand
	r.GET("/v1/pages/meta/:project/*name", createHandler("/v1/pages/meta/:project/*name"))

	// meta
	r.GET("/v1/projects", createHandler("/v1/projects"))
	r.GET("/v1/namespaces", createHandler("/v1/namespaces"))

	//open
	r.GET("/v1/docs", createHandler("/v1/docs"))
	r.GET("/v1/status", createHandler("/v1/status"))

	return r
}

func runTest(router *gin.Engine, endpoint string, group string) {
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Set("group", group)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	fmt.Printf("\nUser of group %s accessing %s \t -> Status: %d\n", group, endpoint, w.Code)
}

func main() {
	gin.SetMode(gin.TestMode)
	// e, err := casbin.NewEnforcer("./model.conf", "./policy.csv")
	e, err := casbin.NewEnforcer("./model_1grp.conf", "./policy_1grp.csv")
	if err != nil {
		fmt.Println(err)
	}

	router := getRBACRouter(
		setupCasbinRBACMWUser(),
		RBAC(CasbinRBACAuthorizer(e)),
	)

	runTest(router, "/v1/exports/download/0/enwiki", "free")
	runTest(router, "/v1/exports/meta/1", "unlimited")
	runTest(router, "/v1/exports/meta/1", "new")
	runTest(router, "/v1/projects", "free")
	runTest(router, "/v1/page-delete", "free")
	runTest(router, "/v1/page-delete", "unlimited")
	runTest(router, "/v1/exports/unknown", "unlimited")
	runTest(router, "/v1/docs", "new")
}
