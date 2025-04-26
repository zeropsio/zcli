package version

import (
	"context"
	_ "embed"
	"io"
	"text/template"

	"github.com/pkg/errors"
)

//go:embed assets/message.txt
var messageTemplate string

func printMessageData(out io.Writer) error {
	var d messageData
	return d.Output(out)
}

type messageData struct {
}

func (_ messageData) CurrentVersion() string {
	return GetCurrent()
}

func (_ messageData) LatestVersion() string {
	latest, _ := GetLatest(context.Background())
	return latest
}

func (_ messageData) LatestUrl() string {
	latest, _ := GetLatestUrl(context.Background())
	return latest
}

func (d messageData) Output(out io.Writer) error {
	t, err := template.New("").Parse(messageTemplate)
	if err != nil {
		return errors.Wrap(err, "Failed to parse message template")
	}
	if err := t.Execute(out, d); err != nil {
		return errors.Wrap(err, "Failed to execute message template")
	}
	return nil
}
