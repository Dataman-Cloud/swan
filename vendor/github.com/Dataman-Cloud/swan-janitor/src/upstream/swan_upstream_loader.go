package upstream

import (
	"fmt"
	"strings"
	"sync"

	"github.com/Dataman-Cloud/swan-janitor/src/loadbalance"

	log "github.com/Sirupsen/logrus"
)

const (
	SWAN_UPSTREAM_LOADER_KEY = "SwanUpstreamLoader"
)

type TargetChangeEvent struct {
	Change       string
	TargetName   string
	TargetIP     string
	TargetPort   string
	FrontendPort string
}

type SwanUpstreamLoader struct {
	UpstreamLoader
	Upstreams    []*Upstream
	changeNotify chan bool
	sync.Mutex
	swanEventChan     chan *TargetChangeEvent
	DefaultUpstreamIp string
	Port              string
	Proto             string
}

func InitSwanUpstreamLoader(defaultUpstreamIp string, defaultPort string, defaultProto string) (*SwanUpstreamLoader, error) {
	swanUpstreamLoader := &SwanUpstreamLoader{}
	swanUpstreamLoader.changeNotify = make(chan bool, 64)
	swanUpstreamLoader.Upstreams = make([]*Upstream, 0)
	swanUpstreamLoader.DefaultUpstreamIp = defaultUpstreamIp
	swanUpstreamLoader.Port = defaultPort
	swanUpstreamLoader.Proto = defaultProto
	swanUpstreamLoader.swanEventChan = make(chan *TargetChangeEvent, 1)
	go swanUpstreamLoader.Poll()
	return swanUpstreamLoader, nil
}

func (swanUpstreamLoader *SwanUpstreamLoader) Poll() {
	log.Infof("SwanUpstreamLoader starts listening app event...")
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("SwanUpstreamLoader poll got error: %s", err)
			swanUpstreamLoader.Poll() // execute poll again
		}
	}()

	for {
		targetChangeEvent := <-swanUpstreamLoader.swanEventChan
		log.Infof("SwanUpstreamLoader receive one app event:%s", targetChangeEvent)

		swanUpstreamLoader.Lock()
		switch strings.ToLower(targetChangeEvent.Change) {
		case "add":
			upstream := buildSwanUpstream(targetChangeEvent, swanUpstreamLoader.DefaultUpstreamIp, swanUpstreamLoader.Port, swanUpstreamLoader.Proto)
			target := buildSwanTarget(targetChangeEvent)
			u := swanUpstreamLoader.Get(upstream.ServiceName)
			if u != nil {
				if u.FieldsEqual(upstream) {
					//case 2: upstream exists, add target
					t := u.GetTarget(target.ServiceID)
					if t != nil {
						//case 3: target exists, update target
						if !t.Equal(target) {
							target.Upstream = u
							u.Remove(t)
							u.Targets = append(u.Targets, target)
							log.Infof("Target [%s] updated in upstream [%s]", target.ServiceID, u.ServiceName)
						}
					} else {
						//case 4: add new target
						target.Upstream = u
						u.Targets = append(u.Targets, target)
						log.Infof("Target [%s] added in upstream [%s]", target.ServiceID, u.ServiceName)
					}
				} else {
					//case 5: update upstream
					//update target
					target.Upstream = upstream
					u.Remove(target)
					upstream.Targets = u.Targets
					upstream.Targets = append(upstream.Targets, target)
					log.Infof("Target [%s] added in upstream [%s]", target.ServiceID, upstream.ServiceName)

					//add loadbalance when upstream created
					loadBalance := loadbalance.NewRoundRobinLoadBalancer()
					loadBalance.Seed()
					upstream.LoadBalance = loadBalance

					//update upstream
					swanUpstreamLoader.Remove(u)
					swanUpstreamLoader.Upstreams = append(swanUpstreamLoader.Upstreams, upstream)
					log.Infof("Upstream [%s] updated in swanUpstreamLoader", upstream.ServiceName)
				}
			} else {
				//case 1: add new upstream
				//set state
				upstream.SetState(STATE_NEW)
				//add target
				target.Upstream = upstream
				upstream.Targets = append(upstream.Targets, target)
				log.Infof("Target [%s] added in upstream [%s]", target.ServiceID, upstream.ServiceName)
				//add loadbalance when upstream created
				loadBalance := loadbalance.NewRoundRobinLoadBalancer()
				loadBalance.Seed()
				upstream.LoadBalance = loadBalance
				swanUpstreamLoader.Upstreams = append(swanUpstreamLoader.Upstreams, upstream)
				log.Infof("Upstream [%s] created", upstream.ServiceName)
			}
		case "del":
			upstream := buildSwanUpstream(targetChangeEvent, swanUpstreamLoader.DefaultUpstreamIp, swanUpstreamLoader.Port, swanUpstreamLoader.Proto)
			target := buildSwanTarget(targetChangeEvent)
			u := swanUpstreamLoader.Get(upstream.ServiceName)
			if u != nil {
				t := u.GetTarget(target.ServiceID)
				if t != nil {
					u.Remove(target)
					log.Infof("Target [%s] removed from upstream [%s]", target.ServiceID, upstream.ServiceName)
					if len(u.Targets) == 0 {
						u.StaleMark = true
						swanUpstreamLoader.Remove(u)
						log.Infof("Upstream [%s] removed", u.ServiceName)
					}
				}
			}
		}
		swanUpstreamLoader.changeNotify <- true
		swanUpstreamLoader.Unlock()
	}
}

func (swanUpstreamLoader *SwanUpstreamLoader) List() []*Upstream {
	swanUpstreamLoader.Lock()
	defer swanUpstreamLoader.Unlock()
	return swanUpstreamLoader.Upstreams
}

func (swanUpstreamLoader *SwanUpstreamLoader) SwanEventChan() chan<- *TargetChangeEvent {
	return swanUpstreamLoader.swanEventChan
}

func (swanUpstreamLoader *SwanUpstreamLoader) ServiceEntries() []string {
	entryList := make([]string, 0)
	for _, u := range swanUpstreamLoader.Upstreams {
		entry := fmt.Sprintf("%s://%s:%s", u.Key().Proto, u.Key().Ip, u.Key().Port)
		entryList = append(entryList, entry)
	}

	return entryList
}

func (swanUpstreamLoader *SwanUpstreamLoader) Get(serviceName string) *Upstream {
	for _, u := range swanUpstreamLoader.Upstreams {
		if u.ServiceName == serviceName {
			return u
			break
		}
	}
	return nil
}

func (swanUpstreamLoader *SwanUpstreamLoader) Remove(upstream *Upstream) {
	index := -1
	for k, v := range swanUpstreamLoader.Upstreams {
		if v == upstream {
			index = k
			break
		}
	}

	if index >= 0 {
		swanUpstreamLoader.Upstreams = append(swanUpstreamLoader.Upstreams[:index], swanUpstreamLoader.Upstreams[index+1:]...)
	}
}

func (swanUpstreamLoader *SwanUpstreamLoader) ChangeNotify() <-chan bool {
	return swanUpstreamLoader.changeNotify
}

func buildSwanTarget(targetChangeEvent *TargetChangeEvent) *Target {
	// create a new target
	var target Target
	taskNamespaces := strings.Split(targetChangeEvent.TargetName, ".")
	taskNum := taskNamespaces[0]
	appName := strings.Join(taskNamespaces[1:], ".")
	target.Address = targetChangeEvent.TargetIP
	target.ServiceName = appName
	target.ServiceID = taskNum
	target.ServiceAddress = targetChangeEvent.TargetIP
	target.ServicePort = targetChangeEvent.TargetPort
	return &target
}

func buildSwanUpstream(targetChangeEvent *TargetChangeEvent, defaultUpstreamIp string, defaultPort string, defaultProto string) *Upstream {
	// create a new upstream
	var upstream Upstream
	taskNamespaces := strings.Split(targetChangeEvent.TargetName, ".")
	appName := strings.Join(taskNamespaces[1:], ".")
	upstream.ServiceName = appName
	upstream.FrontendProto = defaultProto
	upstream.FrontendIp = defaultUpstreamIp
	upstream.FrontendPort = defaultPort
	if targetChangeEvent.FrontendPort != "" {
		upstream.FrontendPort = targetChangeEvent.FrontendPort
	}
	upstream.Targets = make([]*Target, 0)
	upstream.StaleMark = false
	return &upstream
}
