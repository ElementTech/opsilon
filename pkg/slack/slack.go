package slack

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/jatalocks/opsilon/internal/get"
	"github.com/jatalocks/opsilon/internal/logger"
	"github.com/shomali11/slacker"
	"github.com/slack-go/slack"
)

var listDefinition slacker.CommandDefinition = slacker.CommandDefinition{
	Description: "List Available Workflows",
	Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
		repos := request.Param("repos")
		r := []string{}
		if repos != "" {
			r = strings.Split(repos, ",")
		}
		w, err := get.GetWorkflowsForRepo(r)

		if err != nil {
			response.ReportError(err)
		} else if len(w) == 0 {
			response.ReportError(errors.New(fmt.Sprint("Repositories ", r, " do not exist.")))
		} else {

			attachments := []slack.Attachment{}
			for _, v := range w {
				inputs := []slack.AttachmentField{}
				for _, i := range v.Input {
					inputs = append(inputs, slack.AttachmentField{Title: i.Name, Value: i.Default, Short: true})
				}
				attachments = append(attachments, slack.Attachment{
					Color:      "#65C5A6",
					AuthorName: v.Repo,
					Title:      v.ID,
					Text:       v.Description,
				})
			}

			response.Reply("", slacker.WithAttachments(attachments))
		}

	},
}

func App(botToken, appToken string) {
	logger.Operation(botToken, appToken)
	bot := slacker.NewClient(botToken, appToken)

	bot.Command("list {repos}", &listDefinition)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := bot.Listen(ctx)
	if err != nil {
		log.Fatal(err)
	}
}
