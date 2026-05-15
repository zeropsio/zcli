//go:build !devel

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

func printMessageData(ctx context.Context, out io.Writer) error {
	latest, _ := GetLatest(ctx)
	latestURL, _ := GetLatestUrl(ctx)
	d := messageData{
		CurrentVersion: GetCurrent(),
		LatestVersion:  latest,
		LatestUrl:      latestURL,
	}
	return d.Output(out)
}

type messageData struct {
	CurrentVersion string
	LatestVersion  string
	LatestUrl      string
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
