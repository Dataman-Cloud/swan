package types

import (
	"errors"
	"time"
)

type KvmApp struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	RunAs       string         `json:"runAs"`
	Cluster     string         `json:"cluster"`
	OpStatus    string         `json:"operationStatus"`
	TaskCount   int            `json:"taskCount"`
	TasksStatus map[string]int `json:"tasksStatus"`
	ErrMsg      string         `json:"errmsg"`
	Config      *KvmConfig     `json:"config"`
	CreatedAt   time.Time      `json:"created"`
	UpdatedAt   time.Time      `json:"updated"`
}

type KvmTask struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	IP       string    `json:"ip"`
	AgentId  string    `json:"agentId"`
	ErrMsg   string    `json:"errmsg"`
	OpStatus string    `json:"opstatus"`
	Status   string    `json:"status"`     // mesos task status
	DomUuid  string    `json:"domainUuid"` // libvirt domain uuid
	DomName  string    `json:"domainName"` // libvirt domain name
	Created  time.Time `json:"created"`
	Updated  time.Time `json:"updated"`
}

type KvmConfig struct {
	Count int       `json:"count"`
	Cpus  int       `json:"cpus"`
	Mems  int       `json:"mems"`
	Image *KvmImage `json:"iso"`
	Vnc   *Vnc      `json:"vnc"`
}

func (cfg *KvmConfig) Valid() error {
	// verify image
	if img := cfg.Image; img == nil {
		return errors.New("kvm image config required")
	} else {
		if err := img.Valid(); err != nil {
			return err
		}
	}

	// verify vnc
	if vnc := cfg.Vnc; vnc == nil {
		return errors.New("kvm vnc config required")
	} else {
		if err := vnc.Valid(); err != nil {
			return err
		}
	}

	return nil
}

type KvmImage struct {
	Type string `json:"type"` // iso, qcow2
	URI  string `json:"uri"`  // url of image
}

func (img *KvmImage) Valid() error {
	return nil
}

type Vnc struct {
	Enabled  bool   `json:"enabled"`
	Password string `json:"password"`
}

func (vnc *Vnc) Valid() error {
	return nil
}
