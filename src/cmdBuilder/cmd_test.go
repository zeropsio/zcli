package cmdBuilder

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const loremIpsum = "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nam efficitur ut purus sit amet pharetra. Praesent ipsum lacus, varius tincidunt accumsan et, tincidunt eu quam."

func TestWarpDesc(t *testing.T) {
	assert := require.New(t)
	desc := warpString(loremIpsum, 15)
	const expected = "Lorem ipsum dolor\n" + // 15 runes
		"sit amet, consectetur\n" + // 19 runes
		"adipiscing elit.\n" + // 15 runes
		"Nam efficitur ut purus\n" + // 19 runes
		"sit amet pharetra.\n" + // 16 runes
		"Praesent ipsum lacus,\n" + // 19 runes
		"varius tincidunt\n" + // 15 runes
		"accumsan et, tincidunt\n" + // 20 runes
		"eu quam." // 18 runes
	assert.Equal(
		expected,
		desc)
}
