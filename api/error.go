package api

type httpError struct {
	errmsg     string
	statuscode int
}

func (e httpError) Error() string {
	return e.errmsg
}

func (e httpError) StatusCode() int {
	return e.statuscode
}
