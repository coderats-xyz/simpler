package simpler

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	dbx "github.com/go-ozzo/ozzo-dbx"
)

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
func (q *Query) processMeta(meta *MetaData) error {
	switch meta.Key {
	case "name":
		q.Name = fmt.Sprintf("%s/%s", q.Prefix, meta.Value)
	default:
		return fmt.Errorf(`Unknown metadata key %s with value %s`, meta.Key, meta.Value)
	}

	return nil
}

func (q *Query) readSQL(line string) error {
	q.SQL = strings.Trim(fmt.Sprintf("%s %s", q.SQL, line), "\n")
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

// QueryString returns raw SQL string
// string will be empty if registry has no such query
func (r *Registry) QueryString(name string) string {
	if !r.HasQuery(name) {
		return ""
	}

	return r.queryByName(name).SQL
}

// LoadDirectory reads all *.sql files in a given directory
// All nested folders will be processed also
// Nested folder names are preserved and used as a prefix in query name
// so query `delete-user` from a file `sql/user-queries/users` will have name
// `user-queries/users/delete-user`
func (r *Registry) LoadDirectory(dir string) error {
	return r.readDirectory(dir)
}

func (r *Registry) saveQuery(query *Query) error {
	if len(query.Name) == 0 {
		return fmt.Errorf("Found query without a name: %#v", query)
	}

	r.registry[query.Name] = query

	return nil
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

		meta, isMeta, _ := parseMeta(line)

		if isMeta && meta.Key == "name" {
			if query != nil {
				qerr = r.saveQuery(query)
				if qerr != nil {
					return qerr
				}
			}

			query = NewQuery(prefix)
		}

		if isMeta {
			qerr = query.processMeta(meta)
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
				qerr = r.saveQuery(query)
				if qerr != nil {
					return qerr
				}
			}
			break
		}
	}

	return nil
}

func (r *Registry) queryByName(name string) *Query {
	return r.registry[name]
}
