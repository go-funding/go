package services

import (
	"context"
	"database/sql"
	"fuk-funding/go/database/dbtypes"
	sq "github.com/Masterminds/squirrel"
	"github.com/fatih/color"
	"go.uber.org/zap"
)

type Domains struct {
	db  dbtypes.Sql
	log *zap.SugaredLogger
}

var updatedText = color.YellowString("updated")
var insertedText = color.GreenString("inserted")
var skippedText = color.WhiteString("skipped")
var failedText = color.RedString("failed")
var upsertedText = color.YellowString("upserted")

type DomainModel struct {
	ID   int
	Host string
	TLD  string
	SLD  string
	TRD  string
}

const domainColumnNames = "id, host, tld, trd, sld"

func domainColumnScanner(rows *sql.Rows, model *DomainModel) error {
	var tld sql.NullString
	var sld sql.NullString
	var trd sql.NullString

	defer func() {
		model.TLD = tld.String
		model.SLD = sld.String
		model.TRD = trd.String
	}()

	return rows.Scan(&model.ID, &model.Host, &tld, &sld, &trd)
}

func (d *Domains) UpdateDomain(ctx context.Context, model *DomainModel) error {
	defer d.log.Debugf(`%s domain %s`, updatedText, model.Host)

	query := sq.Update("domains").
		Set("tld", model.TLD).
		Set("sld", model.SLD).
		Set("trd", model.TRD).
		Where(sq.Eq{"id": model.ID})

	_, err := d.db.ExecContext(ctx, query)
	return err
}

func (d *Domains) GetDomainsNoLevels(ctx context.Context) ([]DomainModel, error) {
	query := sq.Select(domainColumnNames).
		From("domains").
		Where(sq.And{
			sq.Eq{"tld": nil},
			sq.Eq{"sld": nil},
			sq.Eq{"trd": nil},
		})
	return d.GetDomainsQ(ctx, query)
}

func (d *Domains) GetDomains(ctx context.Context) ([]DomainModel, error) {
	query := sq.Select(domainColumnNames).
		From("domains")
	return d.GetDomainsQ(ctx, query)
}

func (d *Domains) GetDomainsQ(ctx context.Context, sq sq.Sqlizer) ([]DomainModel, error) {
	var domains []DomainModel
	err := d.db.IterateRows(ctx, sq, func(rows *sql.Rows) error {
		var domain DomainModel
		if err := domainColumnScanner(rows, &domain); err != nil {
			return err
		}

		domains = append(domains, domain)
		return nil
	})

	if err != nil {
		return nil, err
	}
	return domains, nil
}

func (d *Domains) GetDomainID(ctx context.Context, domain string) (int, error) {
	query := sq.Select("id").
		From("domains").
		Where(sq.Eq{"host": domain}).
		Limit(1)

	var id int

	row, err := d.db.QueryRowContext(ctx, query)
	if err != nil {
		return 0, err
	}

	err = row.Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (d *Domains) UpsertNewHost(ctx context.Context, domain string) error {
	action := upsertedText
	defer d.log.Debugf(`%s processing domain %s`, action, domain)

	_, err := d.db.ExecContext(
		ctx,
		sq.Insert(`domains`).
			Columns(`host`).
			Values(domain).
			Suffix(`on conflict(host) do nothing`),
	)

	if err != nil {
		action = failedText
		return err
	}

	return nil
}

func NewDomainsService(log *zap.SugaredLogger, db dbtypes.Sql) *Domains {
	return &Domains{db, log.Named("svc[domains]")}
}
