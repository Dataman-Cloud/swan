package middleware

func Metrics() HandlerFunc {
	return func(s *Session) *Session {
		return s
	}
}
