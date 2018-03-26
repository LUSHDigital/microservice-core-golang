package migrate

type Migrator interface {
	Migrate() error
}

type ConnectionOptions struct {
	Host     string
	Port     int
	User     string
	Pass     string
	Database string
}
