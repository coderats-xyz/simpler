package simpler

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	dbx "github.com/go-ozzo/ozzo-dbx"
)

var metaReg = regexp.MustCompile(`--\s([a-zA-Z-]+):\s([a-zA-Z-]+)`)

// Query represents data loaded from a SQL file
type Query struct {
	Prefix string
	Name   string
	SQL    string
}

// NewQuery returns new empty query with a given prefix
func NewQuery(prefix string) *Query {
	return &Query{Prefix: prefix}
}

func (q *Query) readMetadata(line string) error {
	matches := metaReg.FindStringSubmatch(line)

	if len(matches) < 3 {
		return fmt.Errorf(`Could not parse metadata line: "%s"`, line)
	}
	k := matches[1]
	v := matches[2]

	switch k {
	case "name":
		q.Name = fmt.Sprintf("%s/%s", q.Prefix, v)
	default:
		return fmt.Errorf(`Unknown metadata key %s with value %s`, k, v)
	}

	return nil
}

func (q *Query) readSQL(line string) error {
	q.SQL = fmt.Sprintf("%s %s", q.SQL, line)
	return nil
}

// Registry holds queries and *dbx.DB object
// is capable of loading queries from a folder
type Registry struct {
	registry map[string]*Query
	db       *dbx.DB
}

// NewRegistry creates a new registry
// and loads queries from a list of directiories
func NewRegistry(dirs ...string) (*Registry, error) {
	r := &Registry{
		registry: map[string]*Query{},
	}

	for _, dir := range dirs {
		err := r.readDirectory(dir)
		if err != nil {
			return nil, err
		}
	}

	return r, nil
}

// DB returns pointer to *dbx.DB instance
func (r *Registry) DB() *dbx.DB {
	return r.db
}

// Connect registry to a dabatase using adapter and url
func (r *Registry) Connect(adapter string, dbURL string) error {
	db, err := dbx.MustOpen(adapter, dbURL)
	if err != nil {
		return err
	}

	r.db = db

	return nil
}

// HasQuery returns true if registry has loaded query named name
func (r *Registry) HasQuery(name string) bool {
	_, ok := r.registry[name]
	return ok
}

// Query returns *dbx.Query created from a query found in *.sql files with name `name`
func (r *Registry) Query(name string) *dbx.Query {
	if r.db == nil {
		panic("Must connect first before creating a query")
	}

	if !r.HasQuery(name) {
		panic(fmt.Sprintf("Query not found with name %s", name))
	}

	q := r.queryByName(name)
	return r.db.NewQuery(q.SQL)
}

// LoadDirectory reads all *.sql files in a given directory
// All nested folders will be processed also
// Nested folder names are preserved and used as a prefix in query name
// so query `delete-user` from a file `sql/user-queries/users` will have name
// `user-queries/users/delete-user`
func (r *Registry) LoadDirectory(dir string) error {
	return r.readDirectory(dir)
}

func (r *Registry) readDirectory(dir string) error {
	return filepath.Walk(dir, func(file string, info os.FileInfo, err error) error {
		if path.Ext(file) == ".sql" {
			err := r.readFile(dir, file)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *Registry) readFile(dir string, file string) error {
	fileNoDir := strings.Replace(file, dir, "", 1)
	prefix := strings.Replace(fileNoDir, path.Ext(fileNoDir), "", 1)
	if strings.HasPrefix(prefix, "/") {
		prefix = strings.Replace(prefix, "/", "", 1)
	}

	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	reader := bufio.NewReader(f)

	var qerr error
	var line string
	var query *Query

	for {
		line, err = reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return err
		}

		if strings.HasPrefix(line, "-- name:") {
			if query != nil {
				if len(query.Name) == 0 {
					return fmt.Errorf("Found query without a name: %#v", query)
				}

				r.registry[query.Name] = query
			}

			query = NewQuery(prefix)
		}

		if strings.HasPrefix(line, "--") {
			qerr = query.readMetadata(line)
			if qerr != nil {
				return qerr
			}
		} else {
			qerr = query.readSQL(line)
			if qerr != nil {
				return qerr
			}
		}

		if err == io.EOF {
			if query != nil {
				if len(query.Name) == 0 {
					return fmt.Errorf("Found query without a name: %#v", query)
				}

				r.registry[query.Name] = query
			}
			break
		}
	}

	return nil
}

func (r *Registry) queryByName(name string) *Query {
	return r.registry[name]
}
