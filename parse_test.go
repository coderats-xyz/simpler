package simpler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsingMetaLineCorrect(t *testing.T) {
	md, ok, err := parseMeta("-- name: delete_user")

	assert.Nil(t, err)
	assert.True(t, ok)
	assert.Equal(t, "name", md.Key)
	assert.Equal(t, "delete_user", md.Value)

}
