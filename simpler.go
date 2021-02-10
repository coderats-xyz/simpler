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

type Query struct {
	Prefix string
	Name   string
	Sql    string
}

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

func (q *Query) readSql(line string) error {
	q.Sql = fmt.Sprintf("%s %s", q.Sql, line)
	return nil
}

type Registry struct {
	registry map[string]*Query
	DB       *dbx.DB
}

func NewRegistry() *Registry {
	return &Registry{
		registry: map[string]*Query{},
	}
}

func (r *Registry) MustConnect(adapter string, dbUrl string) {
	r.DB = dbx.MustOpen(adapter, dbUrl)
}

func (r *Registry) Query(name string) *dbx.Query {
	if r.DB == nil {
		panic("Must connect first before creating a query")
	}

	q := r.queryByName(name)
	if q == nil {
		panic(fmt.Sprintf("Query not found with name %s", name))
	}

	return r.DB.NewQuery(q.Sql)
}

func (r *Registry) LoadDirectory(dir string) error {
	return r.readDirectory(dir)
}

func (r *Registry) readDirectory(dir string) error {
	return filepath.Walk(dir, func(file string, info os.FileInfo, err error) error {
		if path.Ext(file) == ".sql" {
			err := r.readFile(file)
			if err != nil {
				panic(err)
				return err
			}
		}

		return nil
	})
}

func (r *Registry) readFile(file string) error {
	println("loading queries from")
	println(file)
	prefix := strings.Replace(path.Base(file), path.Ext(file), "", 1)

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
			qerr = query.readSql(line)
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
