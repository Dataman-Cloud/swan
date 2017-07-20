package agent

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"

	"github.com/Dataman-Cloud/swan/agent/ipam"
	"github.com/Dataman-Cloud/swan/agent/janitor"
	"github.com/Dataman-Cloud/swan/agent/resolver"
	"github.com/Dataman-Cloud/swan/config"
	"github.com/Dataman-Cloud/swan/mole"
)

type Agent struct {
	config      *config.AgentConfig
	resolver    *resolver.Resolver
	janitor     *janitor.JanitorServer
	ipam        *ipam.IPAM
	clusterNode *mole.Agent
}

func New(cfg *config.AgentConfig) *Agent {
	agent := &Agent{
		config:   cfg,
		resolver: resolver.NewResolver(cfg.DNS, cfg.Janitor.AdvertiseIP),
		janitor:  janitor.NewJanitorServer(cfg.Janitor),
		ipam:     ipam.New(cfg.IPAM),
	}
	return agent
}

// IPAMSetIPPool called via CLI
func (agent *Agent) IPAMSetIPPool(start, end string) error {
	if agent.ipam == nil {
		return errors.New("ipam not initilized yet")
	}

	if err := agent.ipam.StoreSetup(); err != nil {
		return err
	}

	pool := &ipam.IPPoolRange{
		IPStart: start,
		IPEnd:   end,
	}

	if err := pool.Valid(); err != nil {
		return err
	}

	if err := agent.ipam.SetIPPool(pool); err != nil {
		return err
	}

	os.Stdout.Write([]byte("OK\r\n"))
	return nil
}

func (agent *Agent) StartAndJoin() error {
	// detect healhty leader firstly
	addr, err := agent.detectLeaderAddr()
	if err != nil {
		return err
	}

	// sync all of dns & proxy records on start up
	if err := agent.syncFull(addr); err != nil {
		return fmt.Errorf("full sync manager's records error: %v", err)
	}

	// startup pong & resolver & janitor
	go func() {
		http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`pong`))
		})

		if err := http.ListenAndServe(agent.config.Listen, nil); err != nil {
			log.Fatalln("httpd pong occured fatal error:", err)
		}
	}()

	go func() {
		if err := agent.resolver.Start(); err != nil {
			log.Fatalln("resolver occured fatal error:", err)
		}
	}()

	go func() {
		if err := agent.janitor.Start(); err != nil {
			log.Fatalln("janitor occured fatal error:", err)
		}
	}()

	go func() {
		if err := agent.ipam.Serve(); err != nil {
			log.Fatalln("ipam occured fatal error:", err)
		}
	}()

	// serving protocol & Api with underlying mole
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

		log.Warnln("agent Rejoin ...")
		time.Sleep(time.Second)
	}

	return nil
}

func (agent *Agent) Join() error {
	// detect healthy leader
	addr, err := agent.detectLeaderAddr()
	if err != nil {
		return err
	}
	masterURL, err := url.Parse(addr)
	if err != nil {
		return err
	}

	// setup & join
	agent.clusterNode = mole.NewAgent(&mole.Config{
		Role:   mole.RoleAgent,
		Master: masterURL,
	})

	return agent.clusterNode.Join()
}

func (agent *Agent) NewListener() net.Listener {
	return agent.clusterNode.NewListener()
}

func (agent *Agent) ServeProtocol() error {
	return agent.clusterNode.ServeProtocol()
}

func (agent *Agent) ServeApi(l net.Listener) error {
	log.Println("agent api in serving ...")

	httpd := &http.Server{
		Handler: agent.NewHTTPMux(),
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

func (agent *Agent) detectLeaderAddr() (string, error) {
	for _, addr := range agent.config.JoinAddrs {
		resp, err := http.Get("http://" + addr + "/v1/leader")
		if err != nil {
			log.Warnf("detect swan manager leader %s error %v", addr, err)
			continue
		}
		defer resp.Body.Close() // prevent fd leak

		var info struct {
			Leader string `json:"leader"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
			log.Warnf("detect swan manager leader %s error %v", addr, err)
			continue
		}

		log.Infof("detect swan manager leader %s succeed", info.Leader)
		return "http://" + info.Leader, nil
	}

	return "", errors.New("all of swan manager unavailable")
}
