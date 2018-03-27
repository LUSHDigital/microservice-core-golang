package migrate

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestCockroach_HasMigrationsInPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			"passing",
			"{CALCULATED}",
			true,
		},
		{
			"failing - no path",
			"",
			false,
		},
		{
			"failing - invalid path",
			"/etc/ihope/i/dont/exist	",
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Path setup for passing tests
			if tt.path == "{CALCULATED}" {
				var (
					dir  = os.TempDir()
					file = fmt.Sprintf("%s/temp.sql", dir)
				)
				defer os.RemoveAll(dir)

				err := ioutil.WriteFile(file, nil, 0644)
				assert.Nil(t, err)

				// Update path
				tt.path = dir
			}

			assert.Equal(t, tt.expected, migrationsInPath(tt.path))
		})
	}
}

func TestCockroachOptions_ConnectionString(t *testing.T) {
	type fields struct {
		ConnectionOptions *ConnectionOptions
		MigrationTable    string
		Secure            bool
		SSL               *CockroachSSL
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			"Passing - insecure options",
			fields{
				&ConnectionOptions{
					User:     "test-user",
					Port:     9001,
					Host:     "test-host",
					Database: "test-database",
				},
				"migration_table",
				false,
				nil,
			},
			"cockroach://test-user@test-host:9001/test-database?sslmode=disable&x-migrations-table=migration_table",
			false,
		},
		{
			"Passing - defaults",
			fields{},
			"cockroach://cockroach@localhost:26257/service-db?sslmode=disable&x-migrations-table=schema_migrations",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			co := &CockroachOptions{
				ConnectionOptions: tt.fields.ConnectionOptions,
				MigrationTable:    tt.fields.MigrationTable,
				Secure:            tt.fields.Secure,
				SSL:               tt.fields.SSL,
			}
			got, err := co.ConnectionString()
			if (err != nil) != tt.wantErr {
				t.Errorf("CockroachOptions.ConnectionString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CockroachOptions.ConnectionString() = %v, want %v", got, tt.want)
			}
		})
	}
}
