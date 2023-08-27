package user

import (
	dao "github.com/adsrkey/dynamic-user-segmentation-service/internal/repository/postgres/dao/user"
	"github.com/google/uuid"
)

type ID uuid.UUID

func (uID ID) ToDAO() dao.ID {
	return dao.ID(uID)
}

func (uID *ID) FromDAO(id dao.ID) {
	s := ID(id)
	uID = &s
}
