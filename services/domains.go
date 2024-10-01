package services

import "fuk-funding/go/database"

type Domains struct {
	db database.SqlDatabase
}

func NewDomainsService(db database.SqlDatabase) *Domains {
	return &Domains{
		db,
	}
}
