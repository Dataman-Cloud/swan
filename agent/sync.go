package agent

import (
	"encoding/json"
	"net/http"
	"strings"

	log "github.com/Sirupsen/logrus"

	"github.com/Dataman-Cloud/swan/types"
)

func (agent *Agent) syncFull(addr string) error {
	if !strings.HasPrefix(addr, "http://") {
		addr = "http://" + addr
	}
	addr += "/v1/fullsync"

	resp, err := http.Get(addr)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var full []*types.CombinedEvents
	if err := json.NewDecoder(resp.Body).Decode(&full); err != nil {
		return err
	}

	log.Printf("full syncing %d dns & proxy records ...", len(full))

	for _, cmb := range full {
		var (
			proxy = cmb.Proxy
			dns   = cmb.DNS
		)

		if err := agent.resolver.Upsert(dns); err != nil {
			log.Errorln("upsert dns record error:", err)
			return err
		}

		if err := agent.janitor.UpsertBackend(proxy); err != nil {
			log.Errorln("upsert proxy record error:", err)
			return err
		}
	}

	log.Println("full sync dns & proxy records succeed")
	return nil
}
