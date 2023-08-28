package dto

type Process struct {
	ErrDelCh chan struct{}
	ErrAddCh chan struct{}
	ErrAdd   error
	ErrDel   error
}
