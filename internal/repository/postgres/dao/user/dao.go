package user

import "github.com/google/uuid"

type ID uuid.UUID

func (id *ID) ToUUID() uuid.UUID {
	return uuid.UUID(*id)
}
