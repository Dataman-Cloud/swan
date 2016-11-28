package ns

type rrs map[string]map[string]struct{}

func (r rrs) add(name, host string) bool {
	if host == "" {
		return false
	}
	v, ok := r[name]
	if !ok {
		v = make(map[string]struct{})
		r[name] = v
	} else {
		// don't overwrite existing values
		_, ok = v[host]
		if ok {
			return false
		}
	}
	v[host] = struct{}{}
	return true
}

func (r rrs) First(name string) (string, bool) {
	for host := range r[name] {
		return host, true
	}
	return "", false
}

type rrsKind string

const (
	// A record types
	A rrsKind = "A"
	// SRV record types
	SRV = "SRV"
)

func (kind rrsKind) rrs(rg *RecordGenerator) rrs {
	switch kind {
	case A:
		return rg.As
	case SRV:
		return rg.SRVs
	default:
		return nil
	}
}

// RecordGenerator contains DNS records and methods to access and manipulate
// them. TODO(kozyraki): Refactor when discovery id is available.
type RecordGenerator struct {
	As       rrs
	SRVs     rrs
	SlaveIPs map[string]string
}
