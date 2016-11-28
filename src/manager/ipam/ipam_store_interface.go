package ipam

type IPAMStore interface {
	// to store an ip into a persist store
	SaveIP(ip IP) error

	// retrive an ip from a store
	RetriveIP(k string) (ip IP, err error)

	// retrive all ip from a store
	ListAllIPs() (IPList, error)

	// update a ip address
	UpdateIP(ip IP) error

	// wipe out all ip and recreate the bucket
	EmptyPool() error
}
