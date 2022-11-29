package slack

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/jatalocks/opsilon/internal/concurrency"
	"github.com/jatalocks/opsilon/internal/get"
	"github.com/jatalocks/opsilon/internal/internaltypes"
	"github.com/jatalocks/opsilon/internal/logger"
	"github.com/jatalocks/opsilon/pkg/run"
	"github.com/shomali11/slacker"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

var listDefinition slacker.CommandDefinition = slacker.CommandDefinition{
	Description: "List Available Workflows",
	Examples:    []string{"list", "list myteam", "list examples,myteam"},
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
				inputs := ""
				for _, i := range v.Input {
					inputs += i.Name + " "
				}
				attachments = append(attachments, slack.Attachment{
					Color:      "#65C5A6",
					AuthorName: v.Repo,
					Title:      v.ID,
					Text:       v.Description,
					Footer:     inputs,
				})
			}

			response.Reply("", slacker.WithAttachments(attachments))
		}

	},
}

const (
	selectWorkflowAction = "select-workflow"
)

var runDefinition slacker.CommandDefinition = slacker.CommandDefinition{
	Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
		// happyBtn := slack.NewButtonBlockElement("happy", "true", slack.NewTextBlockObject("plain_text", "Happy 🙂", true, false))
		// happyBtn.Style = "primary"
		// sadBtn := slack.NewButtonBlockElement("sad", "false", slack.NewTextBlockObject("plain_text", "Sad ☹️", true, false))
		// sadBtn.Style = "danger"
		// priority

		// headerText := slack.NewTextBlockObject("mrkdwn", "Please choose a workflow", false, false)
		// headerSection := slack.NewSectionBlock(headerText, nil, nil)

		// firstNameText := slack.NewTextBlockObject("plain_text", "First Name", false, false)
		// firstNameHint := slack.NewTextBlockObject("plain_text", "First Name Hint", false, false)
		// firstNamePlaceholder := slack.NewTextBlockObject("plain_text", "Enter your first name", false, false)
		// firstNameElement := slack.NewPlainTextInputBlockElement(firstNamePlaceholder, "firstName")
		// // Notice that blockID is a unique identifier for a block
		// firstName := slack.NewInputBlock("First Name", firstNameText, firstNameHint, firstNameElement)

		// lastNameText := slack.NewTextBlockObject("plain_text", "Last Name", false, false)
		// lastNameHint := slack.NewTextBlockObject("plain_text", "Last Name Hint", false, false)
		// lastNamePlaceholder := slack.NewTextBlockObject("plain_text", "Enter your first name", false, false)
		// lastNameElement := slack.NewPlainTextInputBlockElement(lastNamePlaceholder, "lastName")
		// lastName := slack.NewInputBlock("Last Name", lastNameText, lastNameHint, lastNameElement)
		// wForm := slack.NewOptionsSelectBlockElement("workflow_select", slack.NewTextBlockObject("plain_text", "workflow", false, false), "workflow", &slack.OptionBlockObject{
		// 	Text: &slack.TextBlockObject{
		// 		Type:     "plain_text",
		// 		Text:     "hi",
		// 		Emoji:    false,
		// 		Verbatim: false,
		// 	},
		// 	Value: "",
		// 	Description: &slack.TextBlockObject{
		// 		Type:     "plain_text",
		// 		Text:     "desc",
		// 		Emoji:    false,
		// 		Verbatim: false,
		// 	},
		// 	URL: "",
		// })

		text := slack.NewTextBlockObject(slack.MarkdownType, "Please select a *workflow*.", false, false)
		textSection := slack.NewSectionBlock(text, nil, nil)

		w, err := get.GetWorkflowsForRepo([]string{})
		if err != nil {
			panic(err)
		}
		options := make([]*slack.OptionBlockObject, 0, len(w))
		for _, v := range w {
			optionText := slack.NewTextBlockObject(slack.PlainTextType, v.ID, false, false)
			optionDesc := slack.NewTextBlockObject(slack.PlainTextType, v.Repo, false, false)
			options = append(options, slack.NewOptionBlockObject(v.ID, optionText, optionDesc))
		}

		placeholder := slack.NewTextBlockObject(slack.PlainTextType, "Select workflow", false, false)
		selectMenu := slack.NewOptionsSelectBlockElement(slack.OptTypeStatic, placeholder, "workflow", options...)

		actionBlock := slack.NewActionBlock(selectWorkflowAction, selectMenu)

		err = response.Reply("", slacker.WithBlocks([]slack.Block{
			// slack.NewSectionBlock(slack.NewTextBlockObject(slack.PlainTextType, "What is your mood today?", true, false), nil, nil),
			// slack.NewActionBlock("mood-block", happyBtn, sadBtn),
			textSection,
			actionBlock,
			// headerSection,
			// firstName,
		}))
		if err != nil {
			panic(err)
		}
	},
}

var workflowDialog = func(obj slack.OptionBlockObject) slack.Dialog {
	name := obj.Value
	elements := []slack.DialogElement{
		// textInput,
		// textareaInput,
		// selectInput,
	}
	w, _ := get.GetWorkflowsForRepo([]string{})
	for _, v := range w {
		if v.ID == name {
			for _, input := range v.Input {
				elements = append(elements, slack.NewTextInput(input.Name, input.Name, input.Default))
			}
		}
	}

	dialog := slack.Dialog{
		CallbackID:  name + "&" + obj.Description.Text,
		Title:       name,
		SubmitLabel: "Run",
		Elements:    elements,
	}
	return dialog
}

var interactive = func(s *slacker.Slacker, event *socketmode.Event, callback *slack.InteractionCallback) {
	fmt.Println(callback.Type)
	switch callback.Type {
	case slack.InteractionTypeDialogSubmission:

		u := new(internaltypes.WorkflowArgument)
		u.Args = callback.Submission
		u.Workflow = strings.Split(callback.CallbackID, "&")[0]
		u.Repo = strings.Split(callback.CallbackID, "&")[1]
		missing, chosenAct := run.ValidateWorkflowArgs(u.Repo, u.Workflow, u.Args)

		if len(missing) > 0 {
			_, _, _ = s.Client().PostMessage(callback.Channel.ID, slack.MsgOptionText(fmt.Sprint("You have a problem in the following fields:", missing), false),
				slack.MsgOptionReplaceOriginal(callback.ResponseURL))
		} else {
			_, _, _ = s.Client().PostMessage(callback.Channel.ID, slack.MsgOptionText("Running "+u.Workflow, false),
				slack.MsgOptionReplaceOriginal(callback.ResponseURL))
		}

		concurrency.ToGraph(chosenAct, nil)

	case slack.InteractionTypeBlockActions:
		if len(callback.ActionCallback.BlockActions) != 1 {
			return
		}
		action := callback.ActionCallback.BlockActions[0]
		s.Client().OpenDialog(callback.TriggerID, workflowDialog(action.SelectedOption))
	}

	// if action.BlockID != "mood-block" {
	// 	return
	// }
	// var text string
	// switch action.ActionID {
	// case "workflow":
	// 	text = action.SelectedOption.Value
	// default:
	// 	text = "I don't understand"
	// }
	// Make new dialog components and open a dialog.
	// Component-Text
	// textInput := slack.NewTextInput("TextSample", "Sample label - Text", "Default value")

	// // Component-TextArea
	// textareaInput := slack.NewTextAreaInput("TexaAreaSample", "Sample label - TextArea", "Default value")

	// // Component-Select menu
	// option1 := slack.DialogSelectOption{
	// 	Label: action.SelectedOption.Value,
	// 	Value: action.SelectedOption.Value,
	// }
	// option2 := slack.DialogSelectOption{
	// 	Label: "Display name 2",
	// 	Value: "Inner value 2",
	// }
	// options := []slack.DialogSelectOption{option1, option2}
	// selectInput := slack.NewStaticSelectDialogInput("SelectSample", "Sample label - Select", options)

	// Open a dialog

	// _, _, _ = s.Client().PostMessage(callback.Channel.ID, slack.MsgOptionText(fmt.Sprint(elements), false),
	// 	slack.MsgOptionReplaceOriginal(callback.ResponseURL))

	s.SocketMode().Ack(*event.Request)
}

func App(botToken, appToken string) {
	logger.Operation(botToken, appToken)
	bot := slacker.NewClient(botToken, appToken)

	bot.Command("list {repos}", &listDefinition)
	bot.Interactive(interactive)
	bot.Command("run", &runDefinition)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := bot.Listen(ctx)
	if err != nil {
		log.Fatal(err)
	}
}
