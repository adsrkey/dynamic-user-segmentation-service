package segment

import (
	dao "github.com/adsrkey/dynamic-user-segmentation-service/internal/repository/postgres/dao/segment"
	"github.com/google/uuid"
)

type Slug string

func (sl Slug) ToDAO() dao.Slug {
	return dao.Slug(sl)
}

func (sl Slug) FromDAO(slug dao.Slug) Slug {
	return Slug(slug)
}

type ID uuid.UUID

func (sID ID) ToDAO() dao.ID {
	return dao.ID(sID)
}

func (sID ID) FromDAO(id dao.ID) ID {
	return ID(id)
}
