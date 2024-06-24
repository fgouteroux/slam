package webserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"text/template"

	"github.com/slack-go/slack"

	log "github.com/sirupsen/logrus"

	alertmanagerTmpl "github.com/fgouteroux/slam/template"
)

var (
	defaultMsgTmpl = `
*Description:* {{ .CommonAnnotations.description }}

*Details:*
{{ range .CommonLabels.SortedPairs }} • *{{ .Name }}:*` + " `{{ .Value }}`\n{{ end }}"
)

func (ws *webserver) sendSlackMessage(channel, title, titleLink, text, footer, color, timestamp string, update bool) (string, string, error) {
	maxCharacters := 1000
	textLength := len(text)

	if textLength > maxCharacters {
		text = text[:maxCharacters] + fmt.Sprintf(" ...truncated to %d chars. was %d.", maxCharacters, textLength)
	}
	msgoptions := []slack.MsgOption{}

	if timestamp != "" && !update {
		msgoptions = append(msgoptions, slack.MsgOptionTS(timestamp))
	}

	msgAttachments := []slack.Attachment{}

	attachment := &slack.Attachment{
		Color:     color,
		Text:      text,
		Title:     title,
		TitleLink: titleLink,
		Footer:    footer,
	}

	msgAttachments = append(msgAttachments, *attachment)

	msgoptions = append(msgoptions, slack.MsgOptionAttachments(msgAttachments...))

	var channelID string
	var err error
	if update {
		_, channelID, _, err = ws.Slack.UpdateMessage(channel, timestamp, msgoptions...)
	} else {
		channelID, timestamp, err = ws.Slack.PostMessage(channel, msgoptions...)
	}
	if err != nil {
		if debug {
			msgJSON, _ := json.Marshal(attachment)
			return "", "", fmt.Errorf("error sending to channel %s with message %s: %w", channel, string(msgJSON), err)
		} else {
			return "", "", fmt.Errorf("error sending to channel %s: %w", channel, err)
		}
	}

	log.Infof("message successfully sent to channel %s", channel)

	return channelID, timestamp, nil
}

func renderTemplate(msg *alertmanagerTmpl.Data, templateName string) (string, error) {
	var err error
	var rendered bytes.Buffer

	if templateName == "" {
		var tmpl *template.Template
		tmpl, _ = template.New("default").Parse(defaultMsgTmpl)
		err = tmpl.Execute(&rendered, msg)
	} else {
		err = msgTmpl.ExecuteTemplate(&rendered, templateName, msg)
	}
	if err != nil {
		return "", err
	}

	return rendered.String(), nil
}
