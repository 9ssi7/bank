package configtest

import (
	"github.com/9ssi7/bank/config"
)

func Load() (*config.App, error) {
	var cnf config.App
	err := config.Bind(&cnf)
	return &cnf, err
}
