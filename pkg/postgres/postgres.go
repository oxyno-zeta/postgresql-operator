package postgres

import (
	"database/sql"
	"fmt"

	"github.com/easymile/postgresql-operator/pkg/apis/postgresql/v1alpha1"
	"github.com/go-logr/logr"
)

type PG interface {
	CreateDB(dbname, username string) error
	CreateSchema(db, role, schema string, logger logr.Logger) error
	CreateExtension(db, extension string, logger logr.Logger) error
	CreateGroupRole(role string) error
	CreateUserRole(role, password string) (string, error)
	UpdatePassword(role, password string) error
	GrantRole(role, grantee string) error
	SetSchemaPrivileges(db, creator, role, schema, privs string, logger logr.Logger) error
	RevokeRole(role, revoked string) error
	AlterDefaultLoginRole(role, setRole string) error
	DropDatabase(db string, logger logr.Logger) error
	DropRole(role, newOwner, database string, logger logr.Logger) error
	GetUser() string
	GetDefaultDatabase() string
	Ping() error
}

type pg struct {
	db               *sql.DB
	log              logr.Logger
	host             string
	user             string
	pass             string
	args             string
	default_database string
}

func NewPG(host, user, password, uri_args, default_database string, cloud_type v1alpha1.ProviderType, logger logr.Logger) PG {
	postgres := &pg{
		log:              logger,
		host:             host,
		user:             user,
		pass:             password,
		args:             uri_args,
		default_database: default_database,
	}

	switch cloud_type {
	case v1alpha1.AWSProvider:
		return newAWSPG(postgres)
	case v1alpha1.AzureProvider:
		return newAzurePG(postgres)
	default:
		return postgres
	}
}

func (c *pg) GetUser() string {
	return c.user
}

func (c *pg) GetDefaultDatabase() string {
	return c.default_database
}

func (c *pg) connect(database string) error {
	db, err := sql.Open("postgres", fmt.Sprintf("postgresql://%s:%s@%s/%s?%s", c.user, c.pass, c.host, database, c.args))
	if err != nil {
		return err
	}
	c.log.Info("connected to postgres server")
	c.db = db
	return nil
}

func (c *pg) Ping() error {
	err := c.connect(c.default_database)
	if err != nil {
		return err
	}
	err = c.db.Ping()
	if err != nil {
		return err
	}
	return c.close()
}

func (c *pg) close() error {
	return c.db.Close()
}
