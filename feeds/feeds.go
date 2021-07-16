package feeds

import (
	"context"
	"encoding/json"
	"time"
)

type (
	Report struct {
		Message  string    `json:"message"`
		Reporter string    `json:"reporter"`
		Listener string    `json:"listener"`
		Source   string    `json:"source"`
		Time     time.Time `json:"time"`
		Status   uint8     `json:"status"`
	}

	ReportList []*Report

	Locstat struct {
		System    string    `json:"system"`
		Time      time.Time `json:"time"`
		Character string    `json:"character"`
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

func (r *Report) String() string {

	//  TODO Handle this error

	b, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}

	return string(b)
}

func (rl ReportList) Len() int {
	return len(rl)
}

func (rl ReportList) Less(i, j int) bool {
	return rl[i].Time.Before(rl[j].Time)
}

func (rl ReportList) Swap(i, j int) {
	rl[i], rl[j] = rl[j], rl[i]
}
