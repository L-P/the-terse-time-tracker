package tt

import (
	"database/sql"
	"fmt"
)

func migrate(db *sql.DB) (err error) {
	cur := getCurrentMigrationVersion(db)
	if cur < 0 {
		return DatabaseError(fmt.Sprintf("invalid database version: %d", cur))
	}

	switch cur {
	case 0:
		err = doInitialMigration(db)
		fallthrough
	case 1:
		break // current version
	default:
		return DatabaseError(fmt.Sprintf("database is at version %d which is not compatible with your local tt version", cur))
	}

	return err
}

func getCurrentMigrationVersion(db *sql.DB) int {
	var version int
	if err := db.QueryRow(
		`SELECT Value FROM Config WHERE Key = 'MigrationVersion' LIMIT 1`,
	).Scan(&version); err != nil {
		return 0
	}

	return version
}

func doInitialMigration(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE "Config" (
            "Key" text NOT NULL,
            "Value" text COLLATE 'BINARY' NOT NULL,
            PRIMARY KEY ("Key")
        );`,

		`CREATE TABLE "Task" (
            "ID" integer NOT NULL,
            "Description" text NOT NULL,
            "StartedAt" integer NOT NULL,
            "StoppedAt" integer NULL,
            "Tags" text COLLATE 'BINARY' NOT NULL,
            PRIMARY KEY ("ID")
        );`,

		`INSERT INTO "Config" ("Key", "Value") VALUES (
            'MigrationVersion', 1
        )`,
	}

	for k := range queries {
		_, err := db.Exec(queries[k])
		if err != nil {
			return BadQueryError{err, queries[k], nil}
		}
	}

	return nil
}
