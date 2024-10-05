package services

import (
	"context"
	"fuk-funding/go/database/dbtypes"
	sq "github.com/Masterminds/squirrel"
	"go.uber.org/zap"
)

type CrunchDataModel struct {
	ID int

	Name                  string
	NumberEmployees       string
	TotalFundingAmountUSD int
	LastFundingAmountUSD  int
	CrunchbaseURL         string
	LinkedInURL           string
	FoundedDate           string

	WebsiteURL  string
	WebsiteHost string
}

type CrunchDataService struct {
	db  dbtypes.Sql
	log *zap.SugaredLogger
}

func (c *CrunchDataService) Insert(ctx context.Context, data *CrunchDataModel) error {
	c.log.Debugf("Inserting crunch data: %v", data)

	query := sq.Insert("crunch_data").
		Columns("name", "number_of_employees", "total_investment_amount_usd", "last_investment_amount_usd", "crunchbase_url", "linkedin_url", "website_url", "founded_at", "website_host").
		Values(data.Name, data.NumberEmployees, data.TotalFundingAmountUSD, data.LastFundingAmountUSD, data.CrunchbaseURL, data.LinkedInURL, data.WebsiteURL, data.FoundedDate, data.WebsiteHost).
		Suffix("RETURNING id")

	rows, err := c.db.QueryRowContext(ctx, query)
	if err != nil {
		return err
	}

	return rows.Scan(&data.ID)
}

func NewCrunchDataService(log *zap.SugaredLogger, db dbtypes.Sql) *CrunchDataService {
	return &CrunchDataService{
		db:  db,
		log: log,
	}
}
