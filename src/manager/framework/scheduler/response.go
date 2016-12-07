package scheduler

import (
	"github.com/Dataman-Cloud/swan/src/mesosproto/sched"
)

type Response struct {
	Calls []*sched.Call
}

func NewResponse() *Response {
	return &Response{
		Calls: make([]*sched.Call, 0),
	}
}
