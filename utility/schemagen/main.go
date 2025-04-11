package main

import (
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"go/format"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	gopluralize "github.com/gertd/go-pluralize"
	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/v2"
	_ "github.com/scylladb/gocqlx/v2/table"
	"github.com/thoas/go-funk"
)

var (
	cmd            = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flagCluster    = cmd.String("cluster", "127.0.0.1", "a comma-separated list of host:port tuples")
	flagKeyspace   = cmd.String("keyspace", "", "keyspace to inspect")
	flagPkgname    = cmd.String("pkgname", "models", "the name you wish to assign to your generated package")
	flagOutputFile = cmd.String("outputfile", "models", "the name of the folder to output to")
	flagOutputDir  = cmd.String("outputdir", "models", "the name of the folder to output to")
	flagUser       = cmd.String("user", "", "user for password authentication")
	flagPassword   = cmd.String("password", "", "password for password authentication")
	flagTables     = cmd.String("tables", "", "a comma-separated list of table names")
	pluralize      = gopluralize.NewClient()
)

var (
	//go:embed keyspace.tmpl
	keyspaceTmpl string
)

func main() {
	err := cmd.Parse(os.Args[1:])
	if err != nil {
		log.Fatalln("can't parse flags")
	}

	if *flagKeyspace == "" {
		log.Fatalln("missing required flag: keyspace")
	}

	if err := schemagen(); err != nil {
		log.Fatalf("failed to generate schema: %s", err)
	}
}

func schemagen() error {
	log.Printf("tables=%s", *flagTables)
	log.Printf("pkgname=%s", *flagPkgname)
	tables := make([]string, 0)
	if len(*flagTables) > 0 {
		tables = strings.Split(*flagTables, ",")
	}

	log.Printf("flagOutputFile=%s", *flagOutputFile)
	log.Printf("flagOutputDir=%s", *flagOutputDir)

	if err := os.MkdirAll(*flagOutputDir, os.ModePerm); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	session, err := createSession()
	if err != nil {
		return fmt.Errorf("open output file: %w", err)
	}
	metadata, err := session.KeyspaceMetadata(*flagKeyspace)
	if err != nil {
		return fmt.Errorf("fetch keyspace metadata: %w", err)
	}
	b, err := renderTemplate(metadata, tables)
	if err != nil {
		return fmt.Errorf("render template: %w", err)
	}
	outputPath := path.Join(*flagOutputDir, *flagOutputFile+".go")

	return ioutil.WriteFile(outputPath, b, os.ModePerm)
}

func renderTemplate(md *gocql.KeyspaceMetadata, tables []string) ([]byte, error) {
	t, err := template.
		New("keyspace.tmpl").
		Funcs(template.FuncMap{"camelize": camelize}).
		Funcs(template.FuncMap{"tableToStructname": tableToStructname}).
		Funcs(template.FuncMap{"mapScyllaToGoType": mapScyllaToGoType}).
		Funcs(template.FuncMap{"typeToString": typeToString}).
		Parse(keyspaceTmpl)

	if err != nil {
		log.Fatalln("unable to parse models template:", err)
	}

	imports := make([]string, 0)
	filteredTables := make([]*gocql.TableMetadata, 0)

	noSpecifiedTalbes := len(tables) == 0
	for _, t := range md.Tables {
		if noSpecifiedTalbes || funk.Contains(tables, t.Name) {
			filteredTables = append(filteredTables, t)
		}
	}

	for _, t := range filteredTables {
		for _, c := range t.Columns {
			if (c.Validator == "timestamp" || c.Validator == "date" || c.Validator == "duration" || c.Validator == "time") && !existsInSlice(imports, "time") {
				imports = append(imports, "time")
			}
			if c.Validator == "decimal" && !existsInSlice(imports, "gopkg.in/inf.v0") {
				imports = append(imports, "gopkg.in/inf.v0")
			}
			if c.Validator == "duration" && !existsInSlice(imports, "github.com/gocql/gocql") {
				imports = append(imports, "github.com/gocql/gocql")
			}
		}
	}

	buf := &bytes.Buffer{}
	data := map[string]interface{}{
		"PackageName": *flagPkgname,
		"Tables":      filteredTables,
		"UserTypes":   md.UserTypes,
		"Imports":     imports,
	}

	if err = t.Execute(buf, data); err != nil {
		return nil, fmt.Errorf("template: %w", err)
	}
	return format.Source(buf.Bytes())
}

func createSession() (gocqlx.Session, error) {
	cluster := gocql.NewCluster(clusterHosts()...)
	if *flagUser != "" {
		cluster.Authenticator = gocql.PasswordAuthenticator{
			Username: *flagUser,
			Password: *flagPassword,
		}
	}
	return gocqlx.WrapSession(cluster.CreateSession())
}

func clusterHosts() []string {
	return strings.Split(*flagCluster, ",")
}

func existsInSlice(s []string, v string) bool {
	for _, i := range s {
		if v == i {
			return true
		}
	}

	return false
}

func tableToStructname(tablename string) string {
	structName := pluralize.Singular(camelize(tablename))
	return structName
}
