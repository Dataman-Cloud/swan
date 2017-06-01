package nameserver

import (
	"fmt"
)

const (
	NONE uint8 = iota << 1
	A
	SRV
)

type RecordChangeEvent struct {
	Change string
	Record
}

type Record struct {
	Cluster string
	RunAs   string
	AppName string
	InsName string
	SlotID  string
	Ip      string
	Port    string
	Type    uint8
	IsProxy bool
}

func (rce *RecordChangeEvent) record() *Record {
	return &rce.Record
}

func (record *Record) Key() string {
	if record.Port != "" {
		return fmt.Sprintf("%s-%s-%s-%s-%s-%s-%s", record.SlotID, record.AppName, record.InsName, record.RunAs, record.Cluster, record.Ip, record.Port)
	} else {
		return fmt.Sprintf("%s-%s-%s-%s-%s-%s", record.SlotID, record.AppName, record.InsName, record.RunAs, record.Cluster, record.Ip)
	}

	return ""
}

func (record *Record) WithSlotDomain() string {
	return fmt.Sprintf("%s.%s.%s.%s.%s", record.SlotID, record.AppName, record.InsName, record.RunAs, record.Cluster)
}

func (record *Record) WithoutSlotDomain() string {
	return fmt.Sprintf("%s.%s.%s.%s", record.AppName, record.InsName, record.RunAs, record.Cluster)
}

func (record *Record) IsSRV() bool {
	return record.Type&SRV != 0
}

func (record *Record) IsA() bool {
	return record.Type&A != 0
}

func (record *Record) IsAAndSRV() bool {
	return record.IsSRV() && record.IsA()
}

func (record *Record) Typ() string {
	switch {
	case record.IsSRV():
		return "SRV"
	case record.IsA():
		return "A"
	case record.IsAAndSRV():
		return "A-SRV"
	}
	return "UNKN"
}
