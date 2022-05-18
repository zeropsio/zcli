package startStopDelete

import (
	"fmt"
	"strings"

	"github.com/zerops-io/zcli/src/constants"
	"github.com/zerops-io/zcli/src/i18n"
)

func askForConfirmation(parent constants.ParentCmd) bool {
	if parent == constants.Project {
		fmt.Print(i18n.DeleteProjectConfirm)
	} else {
		fmt.Print(i18n.DeleteServiceConfirm)
	}

	var response string

	_, err := fmt.Scan(&response)
	if err != nil {
		fmt.Println(err)
		return false
	}

	resp := strings.ToLower(response)
	if resp == "y" || resp == "yes" {
		return true
	} else if resp == "n" || resp == "no" {
		return false
	} else {
		return askForConfirmation(parent)
	}
}
