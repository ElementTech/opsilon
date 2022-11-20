package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/jatalocks/opsilon/internal/config"
	"github.com/jatalocks/opsilon/internal/get"
	"github.com/jatalocks/opsilon/pkg/repo"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func App(port int64) {
	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/list", list)
	e.GET("/repo/list", rlist)
	e.POST("/repo/add", radd)
	e.DELETE("/repo/delete/:repo", rdelete)
	// Start server
	e.Logger.Fatal(e.Start(":" + fmt.Sprint(port)))
}

// Handler
func list(c echo.Context) error {
	repos := c.QueryParam("repos")
	r := []string{}
	if repos != "" {
		r = strings.Split(repos, ",")
	}
	w, err := get.GetWorkflowsForRepo(r)

	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	} else {
		e, err := json.Marshal(w)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		return c.JSONBlob(http.StatusOK, e)
	}
}

// Handler
func rlist(c echo.Context) error {
	e, err := json.Marshal(config.GetConfig())
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSONBlob(http.StatusOK, e)
}
func radd(c echo.Context) error {
	u := new(config.Repo)
	if err := c.Bind(u); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	if err := repo.InsertRepositoryIfValid(*u); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, u)
}

func rdelete(c echo.Context) error {
	repository := c.Param("repo")
	if err := repo.Delete([]string{repository}); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	return c.String(http.StatusOK, repository)
}
