package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/blakewilliams/medium"
	"github.com/stretchr/testify/require"
)

type MyData struct {
	Value int
}

type TeamData struct {
	Team int
}

type MyController struct {
	Name string
}

func (c MyController) RegisterRoutes() medium.Routes[*TeamData] {
	return medium.Routes[*TeamData]{
		"GET /": c.Index,
	}
}

func (c MyController) Index(r *medium.Request[*TeamData]) medium.Response {
	res := &medium.ResponseBuilder{}

	res.WriteStatus(http.StatusOK)
	res.WriteString("hello " + c.Name)

	return res
}

func TestController(t *testing.T) {
	router := medium.New[*TeamData](func(r *medium.RootRequest) *TeamData {
		return &TeamData{}
	})

	c := MyController{Name: "Fox Mulder"}

	Mount(router, c)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rw := httptest.NewRecorder()
	router.ServeHTTP(rw, req)

	require.Equal(t, http.StatusOK, rw.Code)
	require.Equal(t, "hello Fox Mulder", rw.Body.String())
}
