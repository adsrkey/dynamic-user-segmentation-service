package usecase

import (
	"github.com/adsrkey/dynamic-user-segmentation-service/internal/usecase/segment"
	"github.com/adsrkey/dynamic-user-segmentation-service/internal/usecase/user"
)

type Builder interface {
	SetSegment(segment *segment.UseCase) Builder
	SetUser(user *user.UseCase) Builder
	Build() UseCases
}

func New() Builder {
	return &usecases{}
}

type UseCases interface {
	Segment() *segment.UseCase
	User() *user.UseCase
}

type usecases struct {
	SegmentUC *segment.UseCase
	UserUC    *user.UseCase
}

func (uc *usecases) SetSegment(segment *segment.UseCase) Builder {
	uc.SegmentUC = segment
	return uc
}

func (uc *usecases) SetUser(user *user.UseCase) Builder {
	uc.UserUC = user
	return uc
}

func (uc *usecases) Build() UseCases {
	return uc
}

func (uc *usecases) Segment() *segment.UseCase {
	return uc.SegmentUC
}
func (uc *usecases) User() *user.UseCase {
	return uc.UserUC
}
