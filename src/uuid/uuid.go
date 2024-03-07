package uuid

import (
	"encoding/base64"

	"github.com/google/uuid"
)

const encodeUUID = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789AB"

var encoding = base64.NewEncoding(encodeUUID)

func GetShort() string {
	x := uuid.New()
	return encoding.EncodeToString(x[:])[0:22]
}
