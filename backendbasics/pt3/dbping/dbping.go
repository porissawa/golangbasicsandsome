package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strconv"
	"time"

	// embedded postgres server.
	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	_ "github.com/jackc/pgx/v5" // register the db driver
)

func main() {
	timeout := flag.Duration("timeout", 5*time.Second, "timeout for connecting to postgres")
	flag.Parse()

	cfg, err := pgConfigFromEnv()
	if err != nil {
		log.Fatalf("postgres configuration error: %v", err)
	}
	// setup embedded postgres server
	portN, err := strconv.Atoi(cfg.port)
	if err != nil {
		panic(err)
	}

	// we'll mirror the postgres config in the environment so that you can't actually get it 'wrong' when running
	// this example; you do need to set the environment variables, though.
	embeddedCfg := embeddedpostgres.DefaultConfig().
		Username(cfg.user).
		Password(cfg.password).
		Database(cfg.database).
		Port(uint32(portN)).
		Logger(io.Discard) // logs won't help in this example

	embeddedDB := embeddedpostgres.NewDatabase(embeddedCfg)
	if err := embeddedDB.Start(); err != nil {
		panic(err)
	}
	log.Printf("postgres is running on: %s\n", embeddedCfg.GetConnectionURL())
	defer embeddedDB.Stop() // stop the db when main exists so it doesn't keep on running and block the port

	// connect to postgres
	db, err := sql.Open(
		"postgres",
		cfg.String(),
	)
	if err != nil {
		panic(err)
	}
	defer db.Close() // remember to close the db connection too

	// always ping the database to ensure a connection is made.
	// any time you talk to a DB, use a context with a timeout, since DB connections
	// could be lost or delayed indefinitely
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		panic(err)
	}

	log.Println("ping successful")
}

type pgconfig struct {
	user, database, host, password, port string // required
	sslMode                              string // optional
}

func pgConfigFromEnv() (pgconfig, error) {
	var missing []string
	// small closures like this can help reduce code duplication and make intent clearer
	// you generally pay a small performance penalty for this, but for configuration, it's not a big deal
	// you can spare the nanoseconds.
	// I prefer little helper function like this to a complicated configuration framework like viper, cobra, vconfig, etc
	get := func(key string) string {
		val := os.Getenv(key)
		if val == "" {
			missing = append(missing, key)
		}
		return val
	}

	cfg := pgconfig{
		user:     get("PG_USER"),
		database: get("PG_DATABASE"),
		host:     get("PG_HOST"),
		password: get("PG_PASSWORD"),
		port:     get("PG_PORT"),
		sslMode:  os.Getenv("PG_SSLMODE"), // don't use the get closure. Since it's optional, we don't want to add it to missing
	}

	switch cfg.sslMode {
	case "", "disable", "allow", "require", "verify-ca", "verify-full": // valid sslmode
	default:
		return cfg, fmt.Errorf(`invalid sslmode "%s": expected on of: "", "disable", "allow", "require", "verify-ca" or "verify-full"`, cfg.sslMode)
	}

	if len(missing) > 0 {
		sort.Strings(missing) // so it's consistent
		return cfg, fmt.Errorf("missing required environment variables: %v", missing)
	}

	return cfg, nil
}

// String returns the connection string for the given pgconfig.
func (pg pgconfig) String() string {
	s := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", pg.user, pg.password, pg.host, pg.port, pg.database)
	if pg.sslMode != "" {
		s += "?sslmode=" + pg.sslMode
	}
	return s
}
