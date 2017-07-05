package agent

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"

	"github.com/Dataman-Cloud/swan/agent/janitor"
	"github.com/Dataman-Cloud/swan/agent/janitor/upstream"
	"github.com/Dataman-Cloud/swan/agent/nameserver"
	"github.com/Dataman-Cloud/swan/config"
	"github.com/Dataman-Cloud/swan/mole"
	"github.com/Dataman-Cloud/swan/types"
	"github.com/Dataman-Cloud/swan/utils"
)

type Agent struct {
	config      config.AgentConfig
	resolver    *nameserver.Resolver
	janitor     *janitor.JanitorServer
	clusterNode *mole.Agent
	eventCh     chan []byte
}

func New(cfg config.AgentConfig) *Agent {
	masterURL, _ := url.Parse("http://0.0.0.0:10000") // TODO

	agent := &Agent{
		config:   cfg,
		resolver: nameserver.NewResolver(&cfg.DNS),
		janitor:  janitor.NewJanitorServer(&cfg.Janitor),
		eventCh:  make(chan []byte, 1024),
		clusterNode: mole.NewAgent(&mole.Config{
			Role:   mole.RoleAgent,
			Master: masterURL,
		}),
	}
	return agent
}

func (agent *Agent) StartAndJoin() error {
	go agent.resolver.Start()
	go agent.janitor.Start()

	var (
		delayMin = time.Second      // min retry delay 1s
		delayMax = time.Second * 60 // max retry delay 60s
		delay    = delayMin         // retry delay
	)
	for {
		err := agent.Join()
		if err != nil {
			log.Errorln("agent Join() error:", err)
			delay *= 2
			if delay > delayMax {
				delay = delayMax // reset delay to max
			}
			log.Warnln("agent ReJoin in", delay.String())
			time.Sleep(delay)
			continue
		}

		l := agent.NewListener()

		go func(l net.Listener) {
			err := agent.ServeProtocol()
			if err != nil {
				log.Errorln("agent ServeProtocol() error:", err)
				l.Close() // close the listener -> the ServeApi() return with error -> Rejoin triggered.
			}
		}(l)

		log.Println("agent Joined succeed, ready ...")
		delay = delayMin // reset dealy to min
		err = agent.ServeApi(l)
		if err != nil {
			log.Errorln("agent ServeApi() error:", err)
		}
	}

	return nil
}

func (agent *Agent) Join() error {
	return agent.clusterNode.Join()
}

func (agent *Agent) ServeProtocol() error {
	return agent.clusterNode.ServeProtocol()
}

func (agent *Agent) NewListener() net.Listener {
	return agent.clusterNode.NewListener()
}

func (agent *Agent) ServeApi(l net.Listener) error {
	log.Println("agent api in serving ...")
	mux := gin.Default()
	mux.NoRoute(agent.serveProxy)
	agent.janitor.ApiServe(mux.Group("/proxy")) // TODO
	agent.resolver.ApiServe(mux.Group("/dns"))  // TODO
	mux.GET("/sysinfo", agent.sysinfo)

	httpd := &http.Server{
		Handler: mux,
	}

	return httpd.Serve(l)
}

func (agent *Agent) sysinfo(ctx *gin.Context) {
	info, err := Gather()
	if err != nil {
		http.Error(ctx.Writer, err.Error(), http.StatusInternalServerError)
		return
	}

	ctx.JSON(200, info)
}

func (agent *Agent) serveProxy(ctx *gin.Context) {
	var (
		r = ctx.Request
		w = ctx.Writer
	)

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		w.WriteHeader(500)
		return
	}

	connMaster, _, err := hijacker.Hijack()
	if err != nil {
		w.WriteHeader(500)
		return
	}
	defer connMaster.Close()

	connBackend, err := agent.dialBackend()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer connBackend.Close()

	go func() {
		r.Write(connBackend)
	}()

	io.Copy(connMaster, connBackend)
}

func (agent *Agent) dialBackend() (net.Conn, error) {
	return net.Dial("unix", "/var/run/docker.sock")
}

//
// clean up followings laster
//

func (agent *Agent) detectManagerAddr() (string, error) {
	for _, addr := range agent.config.JoinAddrs {
		resp, err := http.Get("http://" + addr + "/ping")
		if err != nil {
			log.Warnf("detect swan manager %s error %v", addr, err)
			continue
		}
		resp.Body.Close() // prevent fd leak

		log.Infof("detect swan manager %s succeed", addr)
		return addr, nil
	}

	return "", errors.New("all of swan manager unavailable")
}

func (agent *Agent) watchEvents() {
	var (
		delayMin = time.Millisecond * 500 // min retry delay 0.5s
		delayMax = time.Second * 60       // max retry delay 60s
		delay    = delayMin               // retry delay
	)

	for {
		managerAddr, err := agent.detectManagerAddr()
		if err != nil {
			log.Errorln("agent Join error:", err)
			delay *= 2
			if delay > delayMax {
				delay = delayMax // reset to max
			}
			log.Printf("agent Rejoin in %s ...", delay)
			time.Sleep(delay)
			continue
		}

		log.Printf("agent Join to manager %s succeed, ready for events ...", managerAddr)
		delay = delayMin // reset to min

		err = agent.watchManagerEvents(managerAddr)
		if err != nil {
			log.Errorf("agent watch on manager events error: %v, retry ...", err)
			time.Sleep(time.Second)
		}
	}
}

func (agent *Agent) dispatchEvents() {
	// send local proxy dns record to dns
	var (
		proxyIP = agent.config.Janitor.AdvertiseIP
		ev      = nameserver.BuildRecordEvent("add", "local_proxy", "PROXY", proxyIP, "80", 0, true)
	)
	agent.resolver.EmitChange(ev)

	for evBody := range agent.eventCh {
		log.Printf("agent caught sse task event: %s", string(evBody))

		var taskEv types.TaskEvent
		if err := json.Unmarshal(evBody, &taskEv); err != nil {
			log.Errorf("agent unmarshal task event error: %v", err)
			continue
		}

		if taskEv.GatewayEnabled {
			ev := genJanitorBackendEvent(&taskEv)
			if ev != nil {
				agent.janitor.EmitEvent(ev)
			}
		}

		if taskEv.Type == types.EventTypeTaskHealthy || taskEv.Type == types.EventTypeTaskUnhealthy {
			ev := genDNSRecordEvent(&taskEv)
			if ev != nil {
				agent.resolver.EmitChange(ev)
			}
		}
	}
}

func (agent *Agent) watchManagerEvents(managerAddr string) error {
	eventsDoesMatter := []string{
		types.EventTypeTaskUnhealthy,
		types.EventTypeTaskHealthy,
		types.EventTypeTaskWeightChange,
	}

	eventsPath := fmt.Sprintf("http://%s/v1/events?catchUp=true", managerAddr)
	resp, err := http.Get(eventsPath)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var (
		ssePrefixData  = "data:"
		ssePrefixEvent = "event:"
		prefixDataLen  = len(ssePrefixData)
		prefixEventLen = len(ssePrefixEvent)
		reader         = bufio.NewReader(resp.Body)
	)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return err
		}

		// skip blank line
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, ssePrefixEvent) {
			eventType := strings.TrimSpace(line[prefixEventLen:])
			if !utils.SliceContains(eventsDoesMatter, eventType) {
				continue
			}

			// read next line of stream
			line, err := reader.ReadString('\n')
			if err != nil {
				return err
			}
			// if line is not data section
			if !strings.HasPrefix(line, ssePrefixData) {
				continue
			}

			agent.eventCh <- []byte(line[prefixDataLen:])
		}
	}
}

func genDNSRecordEvent(taskEv *types.TaskEvent) *nameserver.RecordEvent {
	var (
		act    string
		aid    = taskEv.AppID
		tid    = taskEv.TaskID
		ip     = taskEv.IP
		port   = fmt.Sprintf("%d", taskEv.Port)
		weight = taskEv.Weight
	)
	switch taskEv.Type {
	case types.EventTypeTaskHealthy:
		act = "add"
	case types.EventTypeTaskUnhealthy:
		act = "del"
	default:
		return nil
	}

	return nameserver.BuildRecordEvent(act, tid, aid, ip, port, weight, false)
}

func genJanitorBackendEvent(taskEv *types.TaskEvent) *upstream.BackendEvent {
	var (
		act string

		// upstream
		ups    = taskEv.AppID
		alias  = taskEv.AppAlias
		listen = "" // TODO

		// backend
		backend = taskEv.TaskID
		ip      = taskEv.IP
		port    = taskEv.Port
		weight  = taskEv.Weight
		version = taskEv.VersionID
	)

	switch taskEv.Type {
	case types.EventTypeTaskHealthy:
		act = "add"
	case types.EventTypeTaskUnhealthy:
		act = "del"
	case types.EventTypeTaskWeightChange:
		act = "change"
	default:
		return nil
	}

	return upstream.BuildBackendEvent(act, ups, alias, listen, backend, ip, version, port, weight)
}
