package feeds

import (
	"context"
	"time"
)

type (
	Report struct {
		Message string
		Reporter string
		Listener string
		Source string
		time time.Time
	}

	Locstat struct {
		System    string
		Time      time.Time
		Character string
	}

	IntelFeeder interface {
		Feed(ctx context.Context, reps chan Report, errs chan error) (err error)
	}

	LocationFeeder interface {
		Feed(ctx context.Context, locs chan Locstat, errs chan error) (err error)
	}
)

func (r *Report) Hash() string {
	return r.Message
}

