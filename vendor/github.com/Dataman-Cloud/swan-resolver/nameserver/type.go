package nameserver

type RecordGeneratorChangeEvent struct {
	Change       string
	DomainPrefix string
	Ip           string
	Port         string
	Type         string // "a" or "srv"
}
