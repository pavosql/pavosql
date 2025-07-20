package db

import (
	"github.com/gkits/pavosql/internal/page"
	"github.com/gkits/pavosql/internal/tree"
)

type DB struct {
	pager *page.Pager
	data  *tree.Tree
}
