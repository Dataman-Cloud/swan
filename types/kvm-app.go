package types

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/Dataman-Cloud/swan/mesosproto"
	"github.com/Dataman-Cloud/swan/utils"
	"github.com/gogo/protobuf/proto"
)

// KvmApp is a db kvm app
type KvmApp struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	RunAs       string         `json:"runAs"`
	Cluster     string         `json:"cluster"`
	Desc        string         `json:"desc"`
	TaskCount   int            `json:"taskCount"`       // display, auto setup in store
	TasksStatus map[string]int `json:"tasksStatus"`     // display, auto setup in store
	Config      *KvmConfig     `json:"config"`          // deploy settings
	OpStatus    string         `json:"operationStatus"` // op status
	ErrMsg      string         `json:"errmsg"`          // op errmsg
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
}

func (app *KvmApp) Valid() error {
	if app.Name == "" {
		return errors.New("kvm app name required")
	}
	if app.RunAs == "" {
		return errors.New("kvm app runas required")
	}
	if app.Cluster == "" {
		return errors.New("kvm app cluster required")
	}
	if app.Config == nil {
		return errors.New("kvm app configs required")
	}
	return app.Config.Valid()
}

// KvmTask is a db kvm task
type KvmTask struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	AgentId   string    `json:"agentId"`    // mesos agent id located at, set with offer before launch
	DomUuid   string    `json:"domainUuid"` // libvirt domain uuid, set by mesos update event
	DomName   string    `json:"domainName"` // libvirt domain name, set by mesos update event
	OpStatus  string    `json:"opstatus"`   // op status
	Status    string    `json:"status"`     // mesos task status, set by mesos update event
	ErrMsg    string    `json:"errmsg"`     // mesos task errmsg, set by mesos update event or scheduler according by ops
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	IP string `json:"ip"` // TODO
}

type KvmConfig struct {
	Count       int           `json:"count"`
	Cpus        int           `json:"cpus"`
	Mems        int           `json:"mems"`
	Disks       int           `json:"disks"` // by GiB
	Executor    *Executor     `json:"executor"`
	Image       *KvmImage     `json:"image"`
	Vnc         *Vnc          `json:"vnc"`
	KillPolicy  *KillPolicy   `json:"killPolicy"`
	Constraints []*Constraint `json:"constraints"`
}

func (cfg *KvmConfig) Valid() error {
	if cfg.Count <= 0 {
		return errors.New("count must be positive")
	}

	if cfg.Cpus <= 0 {
		return errors.New("cpus must be positive")
	}

	if cfg.Mems <= 0 {
		return errors.New("mems must be positive")
	}

	if cfg.Disks <= 0 {
		return errors.New("disks must be positive")
	}

	if exec := cfg.Executor; exec == nil {
		return errors.New("kvm executor required")
	} else {
		if err := exec.Valid(); err != nil {
			return err
		}
	}

	if img := cfg.Image; img == nil {
		return errors.New("kvm image config required")
	} else {
		if err := img.Valid(); err != nil {
			return err
		}
	}

	if vnc := cfg.Vnc; vnc == nil {
		return errors.New("kvm vnc config required")
	} else {
		if err := vnc.Valid(); err != nil {
			return err
		}
	}

	if plcy := cfg.KillPolicy; plcy != nil {
		if err := plcy.Valid(); err != nil {
			return err
		}
	}

	for _, cons := range cfg.Constraints {
		if err := cons.validate(); err != nil {
			return err
		}
	}

	return nil
}

func (cfg *KvmConfig) ResourcesRequired() ResourcesRequired {
	return ResourcesRequired{
		CPUs: float64(cfg.Cpus),
		Mem:  float64(cfg.Mems),
	}
}

func (cfg *KvmConfig) BuildResources() []*mesosproto.Resource {
	return []*mesosproto.Resource{
		&mesosproto.Resource{
			Name: proto.String("cpus"),
			Type: mesosproto.Value_SCALAR.Enum(),
			Scalar: &mesosproto.Value_Scalar{
				Value: proto.Float64(float64(cfg.Cpus)),
			},
		},
		&mesosproto.Resource{
			Name: proto.String("mem"),
			Type: mesosproto.Value_SCALAR.Enum(),
			Scalar: &mesosproto.Value_Scalar{
				Value: proto.Float64(float64(cfg.Mems)),
			},
		},
	}
}

func (cfg *KvmConfig) BuildKvmExecutor() *mesosproto.ExecutorInfo {
	return &mesosproto.ExecutorInfo{
		ExecutorId: &mesosproto.ExecutorID{
			Value: proto.String(utils.RandomString(32)),
		},
		Type: mesosproto.ExecutorInfo_CUSTOM.Enum(),
		Command: &mesosproto.CommandInfo{
			Uris: []*mesosproto.CommandInfo_URI{
				&mesosproto.CommandInfo_URI{
					Value:      proto.String(cfg.Executor.URL), // executor URI, eg: http://xxxx/swan-executor
					Executable: proto.Bool(true),
				},
			},
			Shell: proto.Bool(true),
			Value: proto.String(cfg.Executor.Command), // eg: swan-executor kvm
		},
	}
}

func (cfg *KvmConfig) BuildLabels(id, name string) *mesosproto.Labels {
	var labels = make([]*mesosproto.Label, 0)

	kvmLables := map[string]string{
		"SWAN_KVM_APP_ID":       strings.SplitN(name, ".", 2)[1],
		"SWAN_KVM_TASK_ID":      id,
		"SWAN_KVM_TASK_NAME":    name,
		"SWAN_KVM_CPUS":         strconv.Itoa(cfg.Cpus),
		"SWAN_KVM_MEMS":         strconv.Itoa(cfg.Mems),
		"SWAN_KVM_DISKS":        strconv.Itoa(cfg.Disks),
		"SWAN_KVM_IMAGE_TYPE":   cfg.Image.Type,
		"SWAN_KVM_IMAGE_URI":    cfg.Image.URI,
		"SWAN_KVM_VNC_ENABLED":  strconv.FormatBool(cfg.Vnc.Enabled),
		"SWAN_KVM_VNC_PASSWORD": cfg.Vnc.Password,
	}

	for k, v := range kvmLables {
		label := &mesosproto.Label{
			Key:   proto.String(k),
			Value: proto.String(v),
		}

		labels = append(labels, label)
	}

	return &mesosproto.Labels{
		Labels: labels,
	}
}

type Executor struct {
	URL     string `json:"url"`
	Command string `json:"command"`
}

func (exec *Executor) Valid() error {
	if exec.URL == "" {
		return errors.New("executor url required")
	}
	if exec.Command == "" {
		return errors.New("executor command required")
	}
	return nil
}

type KvmImage struct {
	Type string `json:"type"` // iso, qcow2
	URI  string `json:"uri"`  // url of image
}

func (img *KvmImage) Valid() error {
	switch typ := img.Type; typ {
	case "iso", "qcow2":
	default:
		return errors.New("invalid image type")
	}
	if img.URI == "" {
		return errors.New("image url required")
	}
	return nil
}

type Vnc struct {
	Enabled  bool   `json:"enabled"`
	Password string `json:"password"`
}

func (vnc *Vnc) Valid() error {
	if vnc.Enabled && vnc.Password == "" {
		return errors.New("vnc password required")
	}
	if len(vnc.Password) < 6 {
		return errors.New("vnc password length should >= 6")
	}
	return nil
}
