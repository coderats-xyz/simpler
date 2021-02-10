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

	err = q.readSQL("SELECT * FROM")
	assert.Nil(t, err)

	err = q.readSQL("users WHERE id = ?")
	assert.Nil(t, err)

	assert.Equal(t, q.SQL, " SELECT * FROM users WHERE id = ?")
}

func TestReadFile(t *testing.T) {
	r, err := NewRegistry()
	assert.Nil(t, err)

	err = r.readFile("fixtures/sql", "fixtures/sql/users.sql")
	assert.Nil(t, err)

	assert.Len(t, r.registry, 2)
	assert.NotNil(t, r.queryByName("users/select-user"))
	assert.NotNil(t, r.queryByName("users/delete-user"))

	assert.Equal(t, " SELECT * FROM users WHERE id = ? ", r.QueryString("users/select-user"))
}

func TestReadDir(t *testing.T) {
	r, err := NewRegistry("fixtures/sql")
	assert.Nil(t, err)

	assert.Len(t, r.registry, 4)
	assert.NotNil(t, r.queryByName("users/select-user"))
	assert.NotNil(t, r.queryByName("users/delete-user"))
	assert.NotNil(t, r.queryByName("content/posts/select-post"))
	assert.NotNil(t, r.queryByName("content/posts/delete-post"))
}
