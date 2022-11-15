package get

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/jatalocks/opsilon/internal/config"
	"github.com/jatalocks/opsilon/internal/engine"
	"github.com/jatalocks/opsilon/internal/logger"
	"github.com/jatalocks/opsilon/internal/validate"
	"gopkg.in/yaml.v3"
)

// CheckArgs should be used to ensure the right command line arguments are
// passed before executing an example.
func CheckArgs(arg ...string) {
	if len(os.Args) < len(arg)+1 {
		Warning("Usage: %s %s", os.Args[0], strings.Join(arg, " "))
		os.Exit(1)
	}
}

// CheckIfError should be used to naively panics if an error is not nil.
func CheckIfError(err error) {
	if err == nil {
		return
	}

	fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %s", err))
	os.Exit(1)
}

// Info should be used to describe the example commands that are about to run.
func Info(format string, args ...interface{}) {
	fmt.Printf("\x1b[34;1m%s\x1b[0m\n", fmt.Sprintf(format, args...))
}

// Warning should be used to display a warning
func Warning(format string, args ...interface{}) {
	fmt.Printf("\x1b[36;1m%s\x1b[0m\n", fmt.Sprintf(format, args...))
}

func getWorkflows(location config.Location, repo string) *[]engine.Workflow {
	data := []engine.Workflow{}
	logger.Operation("Getting workflows from repo", repo, "in location", location.Path, "type", location.Type)
	if location.Type == "folder" {
		if location.Path[0:1] == "/" {

			err := filepath.Walk(location.Path,
				func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}

					if !info.IsDir() {
						yfile, err := ioutil.ReadFile(path)
						logger.HandleErr(err)
						temp := engine.Workflow{}
						temp.Repo = repo
						err2 := yaml.Unmarshal(yfile, &temp)
						logger.HandleErr(err2)

						data = append(data, temp)
					}

					return nil
				})
			if err != nil {
				logger.Fatal(err)
			}

		}
	} else if location.Type == "git" {

		CheckArgs(location.Path)

		r, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
			URL:           location.Path,
			ReferenceName: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", location.Branch)),
			SingleBranch:  true,
		})
		// ... retrieving the branch being pointed by HEAD
		ref, err := r.Head()
		CheckIfError(err)

		commit, err := r.CommitObject(ref.Hash())
		CheckIfError(err)

		tree, err := commit.Tree()
		CheckIfError(err)
		tree.Files().ForEach(func(f *object.File) error {
			if strings.Contains(f.Name, location.Subfolder) && (strings.Contains(f.Name, "yaml") || strings.Contains(f.Name, "yml")) {
				fReader, err := f.Blob.Reader()
				CheckIfError(err)
				bytes, err := io.ReadAll(fReader)
				CheckIfError(err)
				temp := engine.Workflow{}
				temp.Repo = repo
				err2 := yaml.Unmarshal(bytes, &temp)
				logger.HandleErr(err2)
				data = append(data, temp)
			}
			return nil
		})
	}

	// else {

	// 	yfile, err2 := ioutil.ReadFile(path.Join(path.Dir(viper.ConfigFileUsed()), location.Path))
	// 	logger.HandleErr(err2)
	// 	err3 := yaml.Unmarshal(yfile, &data)
	// 	logger.HandleErr(err3)
	// }

	// } else if location.Type == "url" {
	// 	resp, err := http.Get(location.Path)
	// 	logger.HandleErr(err)
	// 	defer resp.Body.Close()
	// 	buf := new(bytes.Buffer)
	// 	buf.ReadFrom(resp.Body)
	// 	err3 := yaml.Unmarshal(buf.Bytes(), &data)
	// 	logger.HandleErr(err3)
	// }
	return &data
}

func appendToWArray(v config.Repo, workflowArray *[]engine.Workflow) error {
	logger.Info("Repository", v.Name)
	w := *getWorkflows(v.Location, v.Name)
	validate.ValidateWorkflows(&w)
	if len(w) == 0 {
		return errors.New("Cannot fetch workflows from repository " + v.Name + " or it is empty.")
	}
	*workflowArray = append(*workflowArray, w...)
	return nil
}
func GetWorkflowsForRepo(repoList []string) ([]engine.Workflow, error) {
	data := config.GetConfig()
	workflowArray := []engine.Workflow{}
	skipRepoCheck := false
	if len(repoList) == 0 {
		skipRepoCheck = true
	}
	for _, v := range data {
		if skipRepoCheck {
			err := appendToWArray(v, &workflowArray)
			if err != nil {
				return nil, err
			}
		} else {
			if config.StringInSlice(v.Name, repoList) {
				err := appendToWArray(v, &workflowArray)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	return workflowArray, nil
}
