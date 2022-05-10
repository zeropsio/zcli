package uuid

import (
	"encoding/base64"

	uuid "github.com/satori/go.uuid"
)

const encodeUUID = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789AB"

var encoding = base64.NewEncoding(encodeUUID)

func GetShort() string {
	x := uuid.NewV4()
	return encoding.EncodeToString(x[:])[0:22]
}
