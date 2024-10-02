package services

import (
	"context"
	"fuk-funding/go/database/dbtypes"
)

type Domains struct {
	db dbtypes.Sql
}

func (d Domains) CreateTable(ctx context.Context) error {
	_, err := d.db.ExecContext(ctx, `
		create table domains (
		  domain     text      primary key,
		  created_at timestamp not null default now()
		);
	`)
	return err
}

func (d *Domains) UpsertNewDomain(ctx context.Context, domain string) error {
	_, err := d.db.ExecContext(ctx, `insert into domains (domain) values ($1)`, domain)
	return err
}

func NewDomainsService(db dbtypes.Sql) *Domains {
	return &Domains{db}
}
