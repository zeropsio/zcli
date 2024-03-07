package uuid

import (
	"encoding/base64"
	"strings"

	"github.com/google/uuid"
)

func GetShort() string {
	x := uuid.New()
	return encode(x[:])
}

func encode(uuid []byte) string {
	b64 := base64.RawURLEncoding.EncodeToString(uuid)
	// TODO(tikinang): Fix for 1.22, improve.
	b64 = strings.ReplaceAll(b64, "-", "A")
	b64 = strings.ReplaceAll(b64, "_", "B")
	// Should already be 22 chars, just making sure if creators of base64 package change their mind.
	return b64[:22]
}
