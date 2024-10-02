package services

import (
	"context"
	"fuk-funding/go/database"
)

type Domains struct {
	db database.Sql
}

func (d *Domains) UpsertNewDomain(ctx context.Context, domain string) error {
	_, err := d.db.ExecContext(ctx, `insert into domains (domain) values ($1)`, domain)
	return err
}

func NewDomainsService(db database.Sql) *Domains {
	return &Domains{db}
}
