package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jatalocks/opsilon/internal/concurrency"
	"github.com/jatalocks/opsilon/internal/config"
	"github.com/jatalocks/opsilon/internal/db"
	"github.com/jatalocks/opsilon/internal/get"
	"github.com/jatalocks/opsilon/internal/internaltypes"
	"github.com/jatalocks/opsilon/pkg/repo"
	"github.com/jatalocks/opsilon/pkg/run"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mitchellh/hashstructure/v2"
	"github.com/pangpanglabs/echoswagger/v2"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/exp/slices"
)

var ver string

func App(port int64, v string) {
	ver = v
	// Echo instance
	e := echoswagger.New(echo.New(), "/api/v1/docs", &echoswagger.Info{
		Title:       "Opsilon API",
		Description: "This API interface allows for interaction with Opsilon's components the same way CLI does.",
		Version:     ver,
		License: &echoswagger.License{
			Name: "GNU GPLv3",
			URL:  "https://spdx.org/licenses/GPL-3.0-or-later.html",
		},
	}).
		SetResponseContentType("application/json").
		SetScheme("https", "http").
		SetUI(echoswagger.UISetting{DetachSpec: true, HideTop: true})

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

	rgw := e.Group("workflow", "/api/v1/workflow")
	rgw.GET("/list", wlist).
		AddResponse(http.StatusOK, "list workflows that have been run by this server", nil, nil).
		AddParamQuery("", "workflow", "workflow id to view (generated by hashing the workflow), omit to view all", false)
	rgw.DELETE("/delete/:workflow", rgdelete).
		AddResponse(http.StatusOK, "delete a workflow", nil, nil).
		AddParamPath("", "workflow", "workflow to delete")

	rrgw := e.Group("run", "/api/v1/run")
	rrgw.GET("/list", wrlist).
		AddResponse(http.StatusOK, "list runs of a certain workflow", nil, nil).
		AddParamQuery("", "workflow", "workflow id to view (generated by hashing the workflow), omit to view all", false)
	rrgw.GET("/id", wrid).
		AddResponse(http.StatusOK, "get ID of workflow at its latest configuration", nil, nil).
		AddParamQuery("", "workflow", "workflow id", false).
		AddParamQuery("", "repo", "workflow name", false)
	rrgw.GET("/history", wrhistory).
		AddResponse(http.StatusOK, "get history of workflow runs", nil, nil).
		AddParamQuery("", "workflow", "workflow id", false).
		AddParamQuery("", "repo", "workflow name", false)
	rrgw.DELETE("/delete/:run", rrdelete).
		AddResponse(http.StatusOK, "delete a run", nil, nil).
		AddParamPath("", "run", "run to delete")

	e.POST("/api/v1/run", wrun).
		AddResponse(http.StatusOK, "run a workflow", nil, nil).
		AddParamBody(internaltypes.WorkflowArgument{}, "workflow", "workflow to run", true)
	// e.GET("/api/v1/swagger/*", echoSwagger.WrapHandler)
	// Start server
	e.GET("/api/v1/ws", runstream)

	e.Echo().Logger.Fatal(e.Echo().Start(":" + fmt.Sprint(port)))
}

var (
	upgrader = websocket.Upgrader{}
)

func runstream(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer ws.Close()
	for {

		// Read
		// msg := internaltypes.WorkflowQuery{}
		// err := ws.ReadJSON(&msg)
		// if err != nil {
		// 	c.Logger().Error(err)
		// }
		// _, hash := getID(msg.Name, msg.Repo)
		// fmt.Printf("%s\n", msg)
		db.WebSocket(ws)

	}
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
func wlist(c echo.Context) error {
	workflow := c.QueryParam("workflow")
	filter := bson.D{}
	if workflow != "" {
		filter = bson.D{{Key: "workflow", Value: workflow}}
	}
	docs, err := db.FindMany("workflows", filter)

	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	} else {
		e, err := json.Marshal(docs)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		return c.JSONBlob(http.StatusOK, e)
	}
}

// Handler
func wrlist(c echo.Context) error {
	workflow := c.QueryParam("workflow")
	filter := bson.D{}
	if workflow != "" {
		filter = bson.D{{Key: "workflow", Value: workflow}}
	}
	docs, err := db.FindMany("results", filter)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	} else {
		e, err := json.Marshal(docs)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		return c.JSONBlob(http.StatusOK, e)
	}
}

func getID(w, r string) (error, string) {
	wFlows, err := get.GetWorkflowsForRepo([]string{r})
	if err != nil {
		return err, ""
	}
	hashReturn := ""
	for _, v := range wFlows {
		if v.ID == w {
			v.Input = []internaltypes.Input{}
			hash, err := hashstructure.Hash(v, hashstructure.FormatV2, nil)
			if err != nil {
				return err, ""
			}
			hashReturn = fmt.Sprint(hash)

		}
	}
	return nil, fmt.Sprint(hashReturn)
}

// Handler
func wrid(c echo.Context) error {
	workflow := c.QueryParam("workflow")
	repo := c.QueryParam("repo")
	err, hash := getID(workflow, repo)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.String(http.StatusOK, fmt.Sprint(hash))
}

func GroupBy(arr interface{}, groupFunc interface{}) interface{} {
	groupMap := reflect.MakeMap(reflect.MapOf(reflect.TypeOf(groupFunc).Out(0), reflect.TypeOf(arr)))
	for i := 0; i < reflect.ValueOf(arr).Len(); i++ {
		groupPivot := reflect.ValueOf(groupFunc).Call([]reflect.Value{reflect.ValueOf(arr).Index(i)})[0]
		if !groupMap.MapIndex(groupPivot).IsValid() {
			groupMap.SetMapIndex(groupPivot, reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(arr).Elem()), 0, 0))
		}
		groupMap.SetMapIndex(groupPivot, reflect.Append(groupMap.MapIndex(groupPivot), reflect.ValueOf(arr).Index(i)))
	}
	return groupMap.Interface()
}

// Handler
func getHistory(workflow, repo string) (error, []internaltypes.RunResult) {
	groupedDocs := []internaltypes.RunResult{}
	err, hash := getID(workflow, repo)
	if err != nil {
		return err, nil
	}
	filter := bson.D{{Key: "workflow", Value: hash}}
	docs, err := db.FindManyResults("results", filter)
	if err != nil {
		// return c.String(http.StatusInternalServerError, err.Error())
		return err, nil
	}
	runIDS := []string{}
	for _, d := range docs {

		if !slices.Contains(runIDS, d.RunID) {
			runIDS = append(runIDS, d.RunID)
			groupedDocs = append(groupedDocs, internaltypes.RunResult{
				Outputs: d.Outputs,
				SkippedStages: func() uint32 {
					if d.Skipped {
						return 1
					} else {
						return 0
					}
				}(),
				FailedStages: func() uint32 {
					if !d.Skipped && !d.Result {
						return 1
					} else {
						return 0
					}
				}(),
				SuccessfulStages: func() uint32 {
					if !d.Skipped && d.Result {
						return 1
					} else {
						return 0
					}
				}(),
				Logs:     d.Logs,
				Workflow: d.Workflow,
				RunID:    d.RunID,
				Result: func() bool {
					if !d.Skipped && !d.Result {
						return false
					} else {
						return true
					}
				}(),
				RunTime:   time.Duration(d.UpdatedDate.Sub(d.CreatedDate).Seconds()),
				StartTime: d.CreatedDate,
				EndTime:   d.UpdatedDate,
			})
		} else {
			for i, g := range groupedDocs {
				if g.RunID == d.RunID {
					p := &groupedDocs[i]
					p.Outputs = append(p.Outputs, d.Outputs...)
					p.Logs = append(p.Logs, d.Logs...)
					p.RunTime += time.Duration(d.UpdatedDate.Sub(d.CreatedDate).Seconds())
					p.StartTime = func() time.Time {
						if p.StartTime.Unix() > d.CreatedDate.Unix() {
							return d.CreatedDate
						} else {
							return p.StartTime
						}
					}()
					p.EndTime = func() time.Time {
						if p.EndTime.Unix() > d.UpdatedDate.Unix() {
							return p.EndTime
						} else {
							return d.UpdatedDate
						}
					}()
					p.SkippedStages += func() uint32 {
						if d.Skipped {
							return 1
						} else {
							return 0
						}
					}()
					p.FailedStages += func() uint32 {
						if !d.Skipped && !d.Result {
							return 1
						} else {
							return 0
						}
					}()
					p.SuccessfulStages += func() uint32 {
						if !d.Skipped && d.Result {
							return 1
						} else {
							return 0
						}
					}()
					p.Result = (p.FailedStages == 0)

				}
			}
		}
	}
	return nil, groupedDocs
}

func wrhistory(c echo.Context) error {
	workflow := c.QueryParam("workflow")
	repo := c.QueryParam("repo")

	err, groupedDocs := getHistory(workflow, repo)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, groupedDocs)
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
func rgdelete(c echo.Context) error {
	workflow := c.Param("workflow")
	err := db.DeleteOne("workflows", bson.D{{Key: "_id", Value: workflow}})
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	} else {
		return c.String(http.StatusOK, workflow)
	}
}

func rrdelete(c echo.Context) error {
	run := c.Param("run")
	err := db.DeleteMany("results", bson.D{{Key: "runid", Value: run}})
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	} else {
		return c.String(http.StatusOK, run)
	}
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

	concurrency.ToGraph(chosenAct, c, internaltypes.SlackMesseger{Callback: nil})

	return nil
}
