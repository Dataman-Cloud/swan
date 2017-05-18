package store

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Dataman-Cloud/swan/src/utils/dfs"
	"github.com/aanand/compose-file/types"
)

type Application struct {
	ID              string        `json:"id,omitempty"`
	Name            string        `json:"name,omitempty"`
	Version         *Version      `json:"version,omitempty"`
	ProposedVersion *Version      `json:"proposedVersion,omitempty"`
	ClusterID       string        `json:"clusterId,omitempty"`
	StateMachine    *StateMachine `json:"stateMachine,omitempty"`
	CreatedAt       int64         `json:"createdAt,omitempty"`
	UpdatedAt       int64         `json:"updatedAt,omitempty"`
	State           string        `json:"State,omitempty"`
}

func (app *Application) Bytes() []byte {
	var buf bytes.Buffer
	dec := gob.NewEncoder(&buf)
	dec.Encode(app)
	return buf.Bytes()
}

func (app *Application) FromBytes(buf []byte) *Application {
	dec := gob.NewDecoder(bytes.NewBuffer(buf))
	dec.Decode(app)

	return app
}

type Version struct {
	ID           string            `json:"id,omitempty"`
	Command      string            `json:"command,omitempty"`
	Cpus         float64           `json:"cpus,omitempty"`
	Mem          float64           `json:"mem,omitempty"`
	Disk         float64           `json:"disk,omitempty"`
	Instances    int32             `json:"instances,omitempty"`
	RunAs        string            `json:"runAs,omitempty"`
	Container    *Container        `json:"container,omitempty"`
	Labels       map[string]string `protobuf_val:"bytes,2,opt,name=value,proto3"`
	HealthCheck  *HealthCheck      `json:"healthCheck,omitempty"`
	Env          map[string]string `protobuf_val:"bytes,2,opt,name=value,proto3"`
	KillPolicy   *KillPolicy       `json:"killPolicy,omitempty"`
	UpdatePolicy *UpdatePolicy     `json:"updatePolicy,omitempty"`
	Gateway      *Gateway          `json:"gateway,omitempty"`
	Constraints  string            `json:"constraints,omitempty"`
	Uris         []string          `json:"uris,omitempty"`
	Ip           []string          `json:"ip,omitempty"`
	Mode         string            `json:"mode,omitempty"`
	AppName      string            `json:"appName,omitempty"`
	AppID        string            `json:"appID,omitempty"`
	Priority     int32             `json:"priority,omitempty"`
	AppVersion   string            `json:"appVersion,omitempty"`
}

func (version *Version) Bytes() []byte {
	var buf bytes.Buffer
	dec := gob.NewEncoder(&buf)
	dec.Encode(version)
	return buf.Bytes()
}

func (version *Version) FromBytes(buf []byte) *Version {
	dec := gob.NewDecoder(bytes.NewBuffer(buf))
	dec.Decode(version)

	return version
}

type Container struct {
	Type    string    `json:"type,omitempty"`
	Docker  *Docker   `json:"docker,omitempty"`
	Volumes []*Volume `json:"volumes,omitempty"`
}

type Docker struct {
	ForcePullImage bool           `json:"forcePullImage,omitempty"`
	Image          string         `json:"image,omitempty"`
	Network        string         `json:"network,omitempty"`
	Parameters     []*Parameter   `json:"parameters,omitempty"`
	PortMappings   []*PortMapping `json:"portMappings,omitempty"`
	Privileged     bool           `json:"privileged,omitempty"`
}

type Parameter struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

type PortMapping struct {
	ContainerPort int32  `json:"containerPort,omitempty"`
	HostPort      int32  `json:"hostPort,omitempty"`
	Name          string `json:"name,omitempty"`
	Protocol      string `json:"protocol,omitempty"`
}

type Volume struct {
	ContainerPath string `json:"containerPath,omitempty"`
	HostPath      string `json:"hostPath,omitempty"`
	Mode          string `json:"mode,omitempty"`
}

type KillPolicy struct {
	Duration int64 `json:"duration,omitempty"`
}

type UpdatePolicy struct {
	UpdateDelay  int32  `json:"updateDelay,omitempty"`
	MaxRetries   int32  `json:"maxRetries,omitempty"`
	MaxFailovers int32  `json:"maxFailovers,omitempty"`
	Action       string `json:"action,omitempty"`
}

type Gateway struct {
	Enabled bool    `json:"enabled,omitempty"`
	Weight  float64 `json:"weight,omitempty"`
}

type HealthCheck struct {
	ID                  string  `json:"id,omitempty"`
	Address             string  `json:"address,omitempty"`
	Protocol            string  `json:"protocol,omitempty"`
	Port                int32   `json:"port,omitempty"`
	PortIndex           int32   `json:"portIndex,omitempty"`
	PortName            string  `json:"portName,omitempty"`
	Value               string  `json:"value,omitempty"`
	Path                string  `json:"path,omitempty"`
	ConsecutiveFailures uint32  `json:"consecutiveFailures,omitempty"`
	GracePeriodSeconds  float64 `json:"gracePeriodSeconds,omitempty"`
	IntervalSeconds     float64 `json:"intervalSeconds,omitempty"`
	TimeoutSeconds      float64 `json:"timeoutSeconds,omitempty"`
	DelaySeconds        float64 `json:"delaySeconds,omitempty"`
}

type Slot struct {
	Index                int32          `json:"index,omitempty"`
	ID                   string         `json:"id,omitempty"`
	AppID                string         `json:"appId,omitempty"`
	VersionID            string         `json:"versionId,omitempty"`
	State                string         `json:"state,omitempty"`
	MarkForDeletion      bool           `json:"markForDeletion,omitempty"`
	MarkForRollingUpdate bool           `json:"markForRollingUpdate,omitempty"`
	Healthy              bool           `json:"healthy,omitempty"`
	CurrentTask          *Task          `json:"CurrentTask,omitempty"`
	TaskHistory          []*Task        `json:"TaskHistory,omitempty"`
	RestartPolicy        *RestartPolicy `json:"restartPolicy,omitempty"`
	Weight               float64        `json:"weight,omitempty"`
}

func (slot *Slot) Bytes() []byte {
	var buf bytes.Buffer
	dec := gob.NewEncoder(&buf)
	dec.Encode(slot)
	return buf.Bytes()
}

func (slot *Slot) FromBytes(buf []byte) *Slot {
	dec := gob.NewDecoder(bytes.NewBuffer(buf))
	dec.Decode(slot)

	return slot
}

type RestartPolicy struct {
}

type Task struct {
	ID            string   `json:"id,omitempty"`
	AppID         string   `json:"appId,omitempty"`
	VersionID     string   `json:"versionId,omitempty"`
	SlotID        string   `json:"slotId,omitempty"`
	State         string   `json:"state,omitempty"`
	Stdout        string   `json:"stdout,omitempty"`
	Stderr        string   `json:"stderr,omitempty"`
	HostPorts     []uint64 `json:"hostPorts,omitempty"`
	OfferID       string   `json:"offerId,omitempty"`
	AgentID       string   `json:"agentId,omitempty"`
	Ip            string   `json:"ip,omitempty"`
	AgentHostName string   `json:"agentHostName,omitempty"`
	Reason        string   `json:"reason,omitempty"`
	Message       string   `json:"message,omitempty"`
	CreatedAt     int64    `json:"createdAt,omitempty"`
	ArchivedAt    int64    `json:"archivedAt,omitempty"`
	ContainerId   string   `json:"containerId,omitempty"`
	ContainerName string   `json:"containerName,omitempty"`
	Weight        float64  `json:"weight,omitempty"`
}

type OfferAllocatorItem struct {
	SlotID   string `json:"slotId,omitempty"`
	OfferID  string `json:"offerId,omitempty"`
	Hostname string `json:"hostname,omitempty"`
	AgentID  string `json:"agentId,omitempty"`
}

type StateMachine struct {
	State *State `json:"state,omitempty"`
}

type State struct {
	Name                string `json:"name,omitempty"`
	CurrentSlotIndex    int64  `json:"currentSlotIndex,omitempty"`
	TargetSlotIndex     int64  `json:"targetSlotIndex,omitempty"`
	SlotCountNeedUpdate int64  `json:"slotCountNeedUpdate,omitempty"`
}

// compose instance
type Instance struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Desc      string    `json:"desc"`
	Status    string    `json:"status"` // op status
	ErrMsg    string    `json:"errmsg"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// request settings
	ServiceGroup ServiceGroup          `json:"service_group"`
	YAMLRaw      string                `json:"yaml_raw"`
	YAMLEnv      map[string]string     `json:"yaml_env"`
	YAMLExtra    map[string]*YamlExtra `json:"yaml_extra"`
}

func (ins *Instance) RequireConvert() bool {
	return len(ins.ServiceGroup) == 0 && ins.YAMLRaw != ""
}

func (ins *Instance) Valid() error {
	reg := regexp.MustCompile(`^[a-zA-Z0-9]{1,32}$`)
	if !reg.MatchString(ins.Name) {
		return errors.New("instance name should be regexp matched by: " + reg.String())
	}
	if ins.Name == "default" {
		return errors.New("instance name reserved")
	}
	if sg := ins.ServiceGroup; len(sg) > 0 {
		return sg.Valid()
	}
	if ins.YAMLRaw == "" {
		return errors.New("at least one of ServiceGroup or YamlRaw required")
	}
	return nil
}

type YamlExtra struct {
	Priority    uint      `json:"priority"`
	WaitDelay   uint      `json:"wait_delay"` // by second
	PullAlways  bool      `json:"pull_always"`
	Resource    *Resource `json:"resource"`
	Constraints string    `json:"constraints"`
	RunAs       string    `json:"runas"`
	URIs        []string  `json:"uris"`
	IPs         []string  `json:"ips"`
}

type Resource struct {
	CPU   float64  `json:"cpu"`
	Mem   float64  `json:"mem"`
	Disk  float64  `json:"disk"`
	Ports []uint64 `json:"ports"`
}

type ServiceGroup map[string]*DockerService

func (sg ServiceGroup) Valid() error {
	if len(sg) == 0 {
		return errors.New("serviceGroup empty")
	}
	for name, svr := range sg {
		if name == "" {
			return errors.New("service name required")
		}
		if strings.ContainsRune(name, '-') {
			return errors.New(`char '-' not allowed for service name`)
		}
		if name != svr.Name {
			return errors.New("service name mismatched")
		}
		if err := svr.Valid(); err != nil {
			return fmt.Errorf("validate service %s error: %v", name, err)
		}
	}
	return sg.circled()
}

func (sg ServiceGroup) PrioritySort() ([]string, error) {
	m, err := sg.dependMap()
	if err != nil {
		return nil, err
	}
	o := dfs.NewDfsOrder(m)
	return o.PostOrder(), nil
}

func (sg ServiceGroup) circled() error {
	m, err := sg.dependMap()
	if err != nil {
		return err
	}
	c := dfs.NewDirectedCycle(m)
	if cs := c.Cycle(); len(cs) > 0 {
		return fmt.Errorf("dependency circled: %v", cs)
	}
	return nil
}

func (sg ServiceGroup) dependMap() (map[string][]string, error) {
	ret := make(map[string][]string)
	for name, svr := range sg {
		// ensure exists
		for _, d := range svr.Service.DependsOn {
			if _, ok := sg[d]; !ok {
				return nil, fmt.Errorf("missing dependency: %s -> %s", name, d)
			}
		}
		ret[name] = svr.Service.DependsOn
	}
	return ret, nil
}

type DockerService struct {
	Name    string               `json:"name"`
	Service *types.ServiceConfig `json:"service"`
	Network *types.NetworkConfig `json:"network"`
	Volume  *types.VolumeConfig  `json:"volume"`
	Extra   *YamlExtra           `json:"extra"`
}

func (s *DockerService) Valid() error {
	return nil
}
