package usecases

type Builder interface {
	SetSegment(segment Segment) Builder
	SetUser(user User) Builder
	Build() UseCases
}

type UseCases interface {
	Segment() Segment
	User() User
}

type usecases struct {
	SegmentUC Segment
	UserUC    User
}

func New() Builder {
	return &usecases{}
}

func (uc *usecases) SetSegment(segment Segment) Builder {
	uc.SegmentUC = segment
	return uc
}

func (uc *usecases) SetUser(user User) Builder {
	uc.UserUC = user
	return uc
}

func (uc *usecases) Build() UseCases {
	return uc
}

func (uc *usecases) Segment() Segment {
	return uc.SegmentUC
}
func (uc *usecases) User() User {
	return uc.UserUC
}
