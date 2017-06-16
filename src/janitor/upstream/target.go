package upstream

import (
	"errors"
	"fmt"
	"strings"
)

// Target
type Target struct {
	AppID      string  `json:"app_id"`     // uniq id (app,uniq)
	AppAlias   string  `json:"app_alias"`  // http visit hostname (app,uniq)
	AppListen  string  `json:"app_listen"` // listening port on proxy (app,uniq)
	VersionID  string  `json:"version_id"`
	AppVersion string  `json:"app_version"`
	TaskID     string  `json:"task_id"`
	TaskIP     string  `json:"task_ip"`
	TaskPort   uint32  `json:"task_port"`
	Scheme     string  `json:"scheme"` // http / https, auto detect & setup by httpProxy
	Weight     float64 `json:"weihgt"`
}

func (t *Target) Addr() string {
	return fmt.Sprintf("%s:%d", t.TaskIP, t.TaskPort)
}

func (t *Target) Valid() error {
	if t == nil {
		return errors.New("nil targte")
	}
	if t.AppID == "" || t.TaskID == "" {
		return errors.New("app_id or task_id required")
	}
	if t.TaskIP == "" || t.TaskPort == 0 {
		return errors.New("task_ip or task_port required")
	}
	if !strings.HasSuffix(t.TaskID, "-"+t.AppID) {
		return errors.New("invalid task_id, must be suffixed by app_id")
	}
	return nil
}

func (t *Target) Format() *Target {
	t.AppListen = t.tcpListen() // rewrite AppListen
	return t
}

func (t *Target) tcpListen() string {
	if t.AppListen == "" {
		return ""
	}

	ss := strings.Split(t.AppListen, ":")
	if port := ss[len(ss)-1]; port != "" {
		return ":" + port
	}

	return ""
}

// TargetChangeEvent
type TargetChangeEvent struct {
	Change string // add/del/update
	Target
}

func (ev TargetChangeEvent) String() string {
	return fmt.Sprintf("{%s app:%s task:%s ip:%s:%d weight:%f}",
		ev.Change, ev.AppID, ev.TaskID, ev.TaskIP, ev.TaskPort, ev.Weight)
}
