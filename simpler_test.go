package simpler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsingQuery(t *testing.T) {
	q := NewQuery("users")

	err := q.readMetadata("-- name: delete-user")
	assert.Nil(t, err)
	assert.Equal(t, q.Name, "users/delete-user")

	err = q.readSql("SELECT * FROM")
	assert.Nil(t, err)

	err = q.readSql("users WHERE id = ?")
	assert.Nil(t, err)

	assert.Equal(t, q.Sql, " SELECT * FROM users WHERE id = ?")
}

func TestReadFile(t *testing.T) {
	r := NewRegistry()

	err := r.readFile("fixtures/sql/users.sql")
	assert.Nil(t, err)

	assert.Len(t, r.registry, 2)
	assert.NotNil(t, r.queryByName("users/select-user"))
	assert.NotNil(t, r.queryByName("users/delete-user"))
}

func TestReadDir(t *testing.T) {
	r := NewRegistry()

	err := r.readDirectory("fixtures/sql")
	assert.Nil(t, err)

	assert.Len(t, r.registry, 4)
	assert.NotNil(t, r.queryByName("users/select-user"))
	assert.NotNil(t, r.queryByName("users/delete-user"))
	assert.NotNil(t, r.queryByName("posts/select-post"))
	assert.NotNil(t, r.queryByName("posts/delete-post"))
}
