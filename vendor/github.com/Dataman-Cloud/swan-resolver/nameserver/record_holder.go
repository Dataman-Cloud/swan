package nameserver

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/Sirupsen/logrus"
)

type RecordHolder struct {
	Domain     string
	recordsMap map[string]*Record

	mu sync.RWMutex
}

func NewRecordHolder(domain string) *RecordHolder {
	if !strings.HasSuffix(domain, ".") {
		domain = domain + "."
	}

	return &RecordHolder{
		Domain:     domain,
		recordsMap: make(map[string]*Record),
	}
}

func (rh *RecordHolder) Add(record *Record) {
	logrus.Debugf("add %s record %s@%s:%s", record.Type, record.Key(), record.Ip, record.Port)

	rh.mu.Lock()
	defer rh.mu.Unlock()

	rh.recordsMap[record.Key()] = record
}

func (rh *RecordHolder) Del(record *Record) {
	logrus.Debugf("del %s record %s@%s:%s", record.Type, record.Key(), record.Ip, record.Port)
	rh.mu.Lock()
	defer rh.mu.Unlock()

	delete(rh.recordsMap, record.Key())
}

func (rh *RecordHolder) GetA(name string) []*Record {
	rh.mu.RLock()
	defer rh.mu.RUnlock()

	results := make([]*Record, 0)

	gatewayHost := fmt.Sprintf("%s.%s", RESERVED_API_GATEWAY_DOMAIN, rh.Domain)

	// special case, is dns request for gateway
	if strings.HasSuffix(name, gatewayHost) {
		for _, record := range rh.recordsMap {
			if record.IsProxy && record.IsA() {
				results = append(results, record)
			}
		}

		return results
	}

	nameWithoutDomain := strings.Replace(name, "."+rh.Domain, "", -1)

	isDigitPrefix := regexp.MustCompile("^\\d\\..*")
	// A 0.nginx.xcm.cluster.swan.com
	if isDigitPrefix.MatchString(nameWithoutDomain) {
		for _, record := range rh.recordsMap {
			if (record.WithSlotDomain() == nameWithoutDomain) && record.IsA() {
				results = append(results, record)
			}
		}
		// A nginx.xcm.cluster.swan.com
	} else {
		for _, record := range rh.recordsMap {
			if (record.WithoutSlotDomain() == nameWithoutDomain) && record.IsA() {
				results = append(results, record)
			}
		}
	}

	return results
}

func (rh *RecordHolder) GetSRV(name string) []*Record {
	rh.mu.RLock()
	defer rh.mu.RUnlock()
	nameWithoutDomain := strings.Replace(name, "."+rh.Domain, "", -1)

	results := make([]*Record, 0)

	isDigitPrefix := regexp.MustCompile("^\\d\\..*")
	// SRV 0.nginx.xcm.cluster.swan.com
	if isDigitPrefix.MatchString(nameWithoutDomain) {
		for _, record := range rh.recordsMap {
			if (record.WithSlotDomain() == nameWithoutDomain) && record.IsAAndSRV() {
				results = append(results, record)
			}
		}
		// SRV nginx.xcm.cluster.swan.com
	} else {
		for _, record := range rh.recordsMap {
			if (record.WithoutSlotDomain() == nameWithoutDomain) && record.IsAAndSRV() {
				results = append(results, record)
			}
		}
	}

	return results
}
