package repository

import (
	"database/sql"

	"github.com/9ssi7/txn"
	"github.com/9ssi7/txn/txnsql"
)

type txnSqlRepo struct {
	adapter txnsql.SqlAdapter
}

func (r *txnSqlRepo) GetTxnAdapter() txn.Adapter {
	return r.adapter
}

func newTxnSqlRepo(db *sql.DB) txnSqlRepo {
	return txnSqlRepo{
		adapter: txnsql.New(db),
	}
}
