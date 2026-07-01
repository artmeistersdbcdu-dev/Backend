package utlis

import (
	"database/sql"
	"errors"

	"github.com/Blue-Onion/ArtmeisterBackend/config"
)

var conf *config.Config = config.GetConfig()
var Db string = conf.DbUrl

func IsNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

