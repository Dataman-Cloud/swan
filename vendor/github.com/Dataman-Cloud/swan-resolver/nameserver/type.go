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
	Change  string
	Cluster string
	RunAs   string
	AppName string
	SlotID  string
	Ip      string
	Port    string
	Type    uint8
	IsProxy bool // if record  reserved for proxy
}

type Record struct {
	Cluster string
	RunAs   string
	AppName string
	SlotID  string
	Ip      string
	Port    string
	Type    uint8
	IsProxy bool
}

func RecordFromRecordChangeEvent(e *RecordChangeEvent) *Record {
	return &Record{
		Cluster: e.Cluster,
		RunAs:   e.RunAs,
		AppName: e.AppName,
		SlotID:  e.SlotID,
		Ip:      e.Ip,
		Port:    e.Port,
		Type:    e.Type,
		IsProxy: e.IsProxy,
	}
}

func (record *Record) Key() string {
	if record.Port != "" {
		return fmt.Sprintf("%s-%s-%s-%s-%s-%s", record.SlotID, record.AppName, record.RunAs, record.Cluster, record.Ip, record.Port)
	} else {
		return fmt.Sprintf("%s-%s-%s-%s-%s", record.SlotID, record.AppName, record.RunAs, record.Cluster, record.Ip)
	}

	return ""
}

func (record *Record) WithSlotDomain() string {
	return fmt.Sprintf("%s.%s.%s.%s", record.SlotID, record.AppName, record.RunAs, record.Cluster)
}

func (record *Record) WithoutSlotDomain() string {
	return fmt.Sprintf("%s.%s.%s", record.AppName, record.RunAs, record.Cluster)
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
