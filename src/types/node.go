package types

type Node struct {
	ID            string            `json:"id"`
	ListenAddr    string            `json:"listenAddr"`
	AdvertiseAddr string            `json:"advertiseAddr"`
	Status        string            `json:"status"`
	Labels        map[string]string `json:"labels"`
	Role          NodeRole          `json:"role"`
}

type NodeRole string

const (
	RoleManager NodeRole = "manager"
	RoleAgent   NodeRole = "agent"
	RoleMix     NodeRole = "mix"
)
