package services

import (
	"context"
	"errors"
	"fuk-funding/go/ctx"
	"fuk-funding/go/database/dbtypes"
	sq "github.com/Masterminds/squirrel"
	"github.com/fatih/color"
	"github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

var Green = color.New(color.FgGreen).SprintFunc()
var Gray = color.New(color.FgHiBlack).SprintFunc()
var Red = color.New(color.FgRed).SprintFunc()

type Domains struct {
	db  dbtypes.Sql
	log *zap.SugaredLogger
}

var insertedText = Green("inserted")
var skippedText = Gray("skipped")
var failedText = Red("failed")

func (d *Domains) UpsertNewHost(ctx context.Context, domain string) error {
	action := insertedText
	defer d.log.Debugf(`%s processing domain %s`, action, domain)

	_, err := d.db.ExecContext(
		ctx,
		sq.Insert(`domains`).
			Columns(`host`).
			Values(domain).
			Suffix(`on conflict(host) do nothing`),
	)

	var sqlite3Err sqlite3.Error
	if errors.As(err, &sqlite3Err) {
		switch {
		case errors.Is(sqlite3Err.ExtendedCode, sqlite3.ErrConstraintUnique):
			action = skippedText
			return nil
		}
	}

	if err != nil {
		action = failedText
		return err
	}
	return nil
}

func NewDomainsService(ctx *ctx.Context, db dbtypes.Sql) *Domains {
	return &Domains{db, ctx.Logger.Named(
		"[Domain Service]",
	)}
}
