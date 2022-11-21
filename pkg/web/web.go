package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/jatalocks/opsilon/internal/concurrency"
	"github.com/jatalocks/opsilon/internal/config"
	"github.com/jatalocks/opsilon/internal/get"
	"github.com/jatalocks/opsilon/internal/internaltypes"
	"github.com/jatalocks/opsilon/pkg/repo"
	"github.com/jatalocks/opsilon/pkg/run"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pangpanglabs/echoswagger/v2"
)

var ver string

func App(port int64, v string) {
	ver = v
	// Echo instance
	e := echoswagger.New(echo.New(), "/api/v1/docs", nil)

	// Middleware
	e.Echo().Use(middleware.Logger())
	e.Echo().Use(middleware.Recover())
	// Routes
	e.GET("/api/v1/version", version).
		AddResponse(http.StatusOK, "the opsilon binary version", nil, nil)
	e.GET("/api/v1/list", list).
		AddResponse(http.StatusOK, "list of available workflows", nil, nil).
		AddParamQuery("", "repos", "comma seperated list of repositories", false)

	rg := e.Group("repo", "/api/v1/repo")
	rg.GET("/list", rlist).
		AddResponse(http.StatusOK, "list of added repositories", nil, nil)
	rg.POST("/add", radd).
		AddResponse(http.StatusCreated, "add a repository", nil, nil).
		AddParamBody(config.Repo{}, "repo", "repository to add", true)
	rg.DELETE("/delete/:repo", rdelete).
		AddResponse(http.StatusOK, "delete a repository", nil, nil).
		AddParamPath("", "repo", "repository to delete")

	e.POST("/api/v1/run", wrun).
		AddResponse(http.StatusOK, "run a workflow", nil, nil).
		AddParamBody(internaltypes.WorkflowArgument{}, "workflow", "workflow to run", true)
	// e.GET("/api/v1/swagger/*", echoSwagger.WrapHandler)
	// Start server
	e.Echo().Logger.Fatal(e.Echo().Start(":" + fmt.Sprint(port)))
}

func version(c echo.Context) error {
	return c.String(http.StatusOK, ver)
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

func wrun(c echo.Context) error {
	u := new(internaltypes.WorkflowArgument)
	if err := c.Bind(u); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	missing, chosenAct := run.ValidateWorkflowArgs(u.Repo, u.Workflow, u.Args)
	if len(missing) > 0 {
		return c.String(http.StatusBadRequest, fmt.Sprint("You have a problem in the following fields:", missing))
	}

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c.Response().WriteHeader(http.StatusOK)

	concurrency.ToGraph(chosenAct, c)

	return nil
}
