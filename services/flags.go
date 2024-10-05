package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fuk-funding/go/database/dbtypes"
	sq "github.com/Masterminds/squirrel"
	"go.uber.org/zap"
)

type Flags struct {
	db  dbtypes.Sql
	log *zap.SugaredLogger
}

type FlagModel struct {
	Flag           string
	AdditionalData map[string]interface{}
}

func (f *Flags) HasFlag(ctx context.Context, domainID int, flag string) (bool, error) {
	query := sq.Select("true").
		From("flags").
		Where(sq.Eq{"domain_id": domainID, "flag": flag})

	var found bool
	rows, err := f.db.QueryRowContext(ctx, query)
	if err != nil {
		return false, err
	}

	err = rows.Scan(&found)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (f *Flags) UpsertFlag(ctx context.Context, domainID int, flag string, additionalData interface{}) error {
	action := insertedText
	defer f.log.Debugf(`%s processing flag %s for domain ID %d`, action, flag, domainID)

	jsonData, err := json.Marshal(additionalData)
	if err != nil {
		action = failedText
		return err
	}

	query := sq.Insert("flags").
		Columns("domain_id", "flag", "additional_data").
		Values(domainID, flag, string(jsonData)).
		Suffix("ON CONFLICT(domain_id, flag) DO UPDATE SET additional_data = excluded.additional_data")

	_, err = f.db.ExecContext(ctx, query)

	if err != nil {
		action = failedText
		if errors.Is(err, sql.ErrNoRows) {
			action = skippedText
		}
		return err
	}

	return nil
}

func (f *Flags) GetFlagsForDomain(ctx context.Context, domainID int) ([]FlagModel, error) {
	query := sq.Select("flag", "additional_data").
		From("flags").
		Where(sq.Eq{"domain_id": domainID})

	var flags []FlagModel
	err := f.db.IterateRows(ctx, query, func(rows *sql.Rows) error {
		var flag string
		var additionalDataStr string
		if err := rows.Scan(&flag, &additionalDataStr); err != nil {
			return err
		}

		var additionalData map[string]interface{}
		if err := json.Unmarshal([]byte(additionalDataStr), &additionalData); err != nil {
			return err
		}

		flags = append(flags, FlagModel{
			Flag:           flag,
			AdditionalData: additionalData,
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	return flags, nil
}

func NewFlagsService(log *zap.SugaredLogger, db dbtypes.Sql) *Flags {
	return &Flags{db, log.Named("svc[flags]")}
}
