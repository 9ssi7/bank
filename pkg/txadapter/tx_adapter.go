package txadapter

import "github.com/9ssi7/txn"

type Repo interface {
	GetTxnAdapter() txn.Adapter
}
