package types

type ResolverAcceptor struct {
	ID         string `json:"id"`
	RemoteAddr string `json:"remoteAddr"`
	Status     string `json:"status"`
}

type JanitorAcceptor struct {
	ID         string `json:"id"`
	RemoteAddr string `json:"remoteAddr"`
	Status     string `json:"status"`
}
