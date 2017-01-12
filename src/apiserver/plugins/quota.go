package plugins

import (
	"bytes"
	"encoding/json"
	"io/ioutil"

	"github.com/Dataman-Cloud/swan/src/manager/framework/store"
	rafttypes "github.com/Dataman-Cloud/swan/src/manager/raft/types"
	"github.com/Dataman-Cloud/swan/src/types"

	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
	"golang.org/x/net/context"
	"net/http"
)

const (
	HardLimitTierType    string = "hardLimit"
	OverCapacityTierType string = "overCapacity"
	FreeOfChargeTierType string = "freeOfCharge"
)

// QuotaControl introspects incoming requests and make decisions by evaluating the quota of the quotaGroup.
// return 403 if exceed the limit, otherwise update the quotaStatus.
func QuotaControl(store store.Store) restful.FilterFunction {
	return func(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
		// cache request body
		var reqBody []byte
		if req.Request.Body != nil {
			reqBody, _ = ioutil.ReadAll(req.Request.Body)
		}
		req.Request.Body = ioutil.NopCloser(bytes.NewBuffer(reqBody))

		// create app
		if req.Request.Method == "POST" {
			var version types.Version

			// TODO(xychu): error handling
			decoder := json.NewDecoder(bytes.NewReader(reqBody))
			decoder.UseNumber()
			decoder.Decode(&version)

			logrus.Debugf("Got request method: %s, URL: %s, body: %#v",
				req.Request.Method,
				req.Request.URL.RequestURI(),
				version,
			)

			quota, err := store.GetQuota(version.RunAs)
			if err != nil {
				logrus.Errorf("%s did not have a quote spec.", version.RunAs)
				resp.WriteErrorString(http.StatusForbidden,
					fmt.Sprintf("%s did not have a quote spec.", version.RunAs))
				return
				// FIXME(xychu): continue process for demo purpose
				//chain.ProcessFilter(req, resp)
				//return
			}

			logrus.Debugf("%#v", quota.Quotas)
			// priority 200+ HardLimit tier
			// 0-199 OverCapacity tier
			// -1 FreeToCharge
			if version.Priority >= 200 {
				logrus.Debugf("t1 quota: %#v", quota.Quotas[HardLimitTierType])
				tierQuota, ok := quota.Quotas[HardLimitTierType]
				if ok {
					logrus.Debugf("t1 quota spec: %#v", tierQuota.QuotaSpec)
					logrus.Debugf("t1 quota status: %#v", tierQuota.QuotaStatus)
					targetQuotaStatus := new(rafttypes.ResourceQuotaStatus)
					if tierQuota.QuotaStatus == nil {
						logrus.Debugf("status nil")
						targetQuotaStatus.Offered = map[string]int64{}
						logrus.Debugf("set offered")
						targetQuotaStatus.Offered["cpu"] = int64(version.CPUs * 1000) // times 1000 to trans 0.01 to int64
						logrus.Debugf("set cpu")
						targetQuotaStatus.Offered["mem"] = int64(version.Mem)
					} else {
						targetQuotaStatus = tierQuota.QuotaStatus
						targetQuotaStatus.Offered["cpu"] += int64(version.CPUs * 1000)
						targetQuotaStatus.Offered["mem"] += int64(version.Mem)
					}

					logrus.Debugf("t1 before evaluate")
					exceed, resource := evaluate(targetQuotaStatus.Offered, tierQuota.QuotaSpec.Limit)
					logrus.Debugf("t1 after evaluate")
					if exceed {
						resp.WriteErrorString(http.StatusForbidden,
							fmt.Sprintf("resource[%s] exceed quota limit[%d/%d]", resource,
								targetQuotaStatus.Offered[resource], tierQuota.QuotaSpec.Limit[resource]))
						return
					} else {
						logrus.Debugf("t1 before update")
						tierQuota.QuotaStatus = targetQuotaStatus
						quota.Quotas[HardLimitTierType] = tierQuota
						// TODO(xychu): atomic
						_ = store.UpdateQuota(context.TODO(), quota, nil)
						logrus.Debugf("t1 after update")
					}
				} else {
					logrus.Errorf("%s did not have a quote spec.", version.RunAs)
					resp.WriteErrorString(http.StatusForbidden,
						fmt.Sprintf("%s did not have a quote spec.", version.RunAs))
					return
				}
			} else if version.Priority > 0 {
				logrus.Debugf("t2 quota: %#v", quota.Quotas[OverCapacityTierType])
				tierQuota, ok := quota.Quotas[OverCapacityTierType]
				if ok {
					logrus.Debugf("t2 quota spec: %#v", tierQuota.QuotaSpec)
					logrus.Debugf("t2 quota status: %#v", tierQuota.QuotaStatus)
					targetQuotaStatus := new(rafttypes.ResourceQuotaStatus)
					if tierQuota.QuotaStatus == nil {
						logrus.Debugf("status nil")
						targetQuotaStatus.Offered = map[string]int64{}
						logrus.Debugf("set offered")
						targetQuotaStatus.Offered["cpu"] = int64(version.CPUs * 1000) // times 1000 to trans 0.01 to int64
						logrus.Debugf("set cpu")
						targetQuotaStatus.Offered["mem"] = int64(version.Mem)
					} else {
						targetQuotaStatus = tierQuota.QuotaStatus
						targetQuotaStatus.Offered["cpu"] += int64(version.CPUs * 1000)
						targetQuotaStatus.Offered["mem"] += int64(version.Mem)
					}

					logrus.Debugf("t2 before evaluate")
					exceed, resource := evaluate(targetQuotaStatus.Offered, tierQuota.QuotaSpec.Limit)
					logrus.Debugf("t2 after evaluate")
					if exceed {
						resp.WriteErrorString(http.StatusForbidden,
							fmt.Sprintf("resource[%s] exceed quota limit[%d/%d]", resource,
								targetQuotaStatus.Offered[resource], tierQuota.QuotaSpec.Limit[resource]))
						return
					} else {
						logrus.Debugf("t2 before update")
						tierQuota.QuotaStatus = targetQuotaStatus
						quota.Quotas[OverCapacityTierType] = tierQuota
						// TODO(xychu): atomic
						_ = store.UpdateQuota(context.TODO(), quota, nil)
						logrus.Debugf("t2 after update")
					}
				} else {
					logrus.Errorf("%s did not have a quote spec.", version.RunAs)
					resp.WriteErrorString(http.StatusForbidden,
						fmt.Sprintf("%s did not have a quote spec.", version.RunAs))
					return
				}
			} else if version.Priority == -1 {
				logrus.Debugf("t3 quota: %#v", quota.Quotas[FreeOfChargeTierType])
				tierQuota, ok := quota.Quotas[FreeOfChargeTierType]
				if ok {
					logrus.Debugf("t3 quota status: %#v", tierQuota.QuotaStatus)
					targetQuotaStatus := new(rafttypes.ResourceQuotaStatus)
					if tierQuota.QuotaStatus == nil {
						logrus.Debugf("status nil")
						targetQuotaStatus.Offered = map[string]int64{}
						logrus.Debugf("set offered")
						targetQuotaStatus.Offered["cpu"] = int64(version.CPUs * 1000) // times 1000 to trans 0.01 to int64
						logrus.Debugf("set cpu")
						targetQuotaStatus.Offered["mem"] = int64(version.Mem)
					} else {
						targetQuotaStatus = tierQuota.QuotaStatus
						targetQuotaStatus.Offered["cpu"] += int64(version.CPUs * 1000)
						targetQuotaStatus.Offered["mem"] += int64(version.Mem)
					}

					tierQuota.QuotaStatus = targetQuotaStatus
					quota.Quotas[FreeOfChargeTierType] = tierQuota
					// TODO(xychu): atomic
					_ = store.UpdateQuota(context.TODO(), quota, nil)
				} else {
					tierQuota := new(rafttypes.TierResourceQuota)
					targetQuotaStatus := new(rafttypes.ResourceQuotaStatus)

					logrus.Debugf("status nil")
					targetQuotaStatus.Offered = map[string]int64{}
					logrus.Debugf("set offered")
					targetQuotaStatus.Offered["cpu"] = int64(version.CPUs * 1000) // times 1000 to trans 0.01 to int64
					logrus.Debugf("set cpu")
					targetQuotaStatus.Offered["mem"] = int64(version.Mem)

					tierQuota.QuotaStatus = targetQuotaStatus
					quota.Quotas[FreeOfChargeTierType] = tierQuota
					// TODO(xychu): atomic
					_ = store.UpdateQuota(context.TODO(), quota, nil)
				}
			} else {
				logrus.Errorf("Unsupported priority: %d", version.Priority)
			}

		}

		// TODO(xychu): handle scale-up app for slot count quota

		// TODO(xychu): handle update app with cpu or mem changed

		chain.ProcessFilter(req, resp)
	}
}

func evaluate(offered, limit map[string]int64) (bool, string) {
	for resource, value := range offered {
		if limit[resource] < value {
			return true, resource
		}
	}
	return false, ""
}
