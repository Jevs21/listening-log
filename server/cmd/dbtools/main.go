package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"

	"listening-log/server/dbtools"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		fmt.Fprintln(os.Stderr, "error: DATABASE_URL environment variable is required")
		os.Exit(1)
	}

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		fmt.Fprintf(os.Stderr, "error: database not reachable: %v\n", err)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "export":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: dbtools export <path>")
			os.Exit(1)
		}
		if err := dbtools.Export(db, os.Args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	case "import":
		importCmd := flag.NewFlagSet("import", flag.ExitOnError)
		mode := importCmd.String("mode", "", "Import mode: overwrite or merge")
		importCmd.Parse(os.Args[2:])

		if *mode != "overwrite" && *mode != "merge" {
			fmt.Fprintln(os.Stderr, "error: --mode must be 'overwrite' or 'merge'")
			os.Exit(1)
		}
		args := importCmd.Args()
		if len(args) < 1 {
			fmt.Fprintln(os.Stderr, "usage: dbtools import --mode=<overwrite|merge> <path>")
			os.Exit(1)
		}
		if err := dbtools.Import(db, args[0], *mode); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	default:
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: dbtools <command> [options]")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "commands:")
	fmt.Fprintln(os.Stderr, "  export <path>                          Export database to tar.gz archive")
	fmt.Fprintln(os.Stderr, "  import --mode=<overwrite|merge> <path> Import from tar.gz archive")
}
