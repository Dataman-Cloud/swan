package api

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Dataman-Cloud/swan/api/mock"
)

var (
	baseUrl string
	server  *httptest.Server
)

func TestMain(m *testing.M) {
	server = startHttpServer()
	baseUrl = server.URL
	ret := m.Run()
	server.Close()
	os.Exit(ret)
}

func startHttpServer() *httptest.Server {
	srv := NewServer(&mock.Backend{})
	return httptest.NewServer(srv.createMux())
}

func TestBuildApplication(t *testing.T) {
	req, _ := http.NewRequest("POST", baseUrl+"/v1/apps", nil)

	resp, _ := http.DefaultClient.Do(req)
	assert.Equal(t, resp.StatusCode, 500)

	header := map[string][]string{
		"Content-Type": {"application/json"},
	}

	req.Header = header
	resp, _ = http.DefaultClient.Do(req)
	assert.Equal(t, resp.StatusCode, 500)

	var jsonStr = []byte(`{
  				"id": "nginx0002",
  				"cmd": null,
  				"cpus": 0.01,
  				"mem": 6,
  				"disk": 0,
  				"instances": 1,
  				"container": {
  				  "docker": {
  				    "image": "nginx:1.10",
  				    "network": "BRIDGE",
  				    "forcePullImage": false,
  				    "privileged": true,
  				    "parameters": [
  				      {
  				          "key": "label",
  				          "value": "APP_ID=nginx"
  				      }
  				    ],
      				    "portMappings": [
      				        {
      				            "containerPort": 80,
      				            "protocol": "tcp",
      				            "name": "web"
      				        }
      				    ]
    				},
    				"type": "DOCKER",
    				"volumes": [
    				  {
    				    "hostPath": "/home",
    				    "containerPath": "/data",
    				    "mode": "RW"
    				  }
    					]
  				},
  				"env": {
  				  "DB": "mysql"
  				},
  				"label": {
  				  "USER_ID": "1"
  				},
  				"killPolicy": {
  				  "duration": 5
  				},
			        "healthChecks": [
			          {
			            "protocol": "TCP",
			            "path": "/",
			            "portIndex": 0,
			            "gracePeriodSeconds": 5,
			            "intervalSeconds": 3,
			            "timeoutSeconds": 3,
			            "maxConsecutiveFailures": 3
			          }
			       ],
                               "updatePolicy": {
                                       "updateDelay": 30,
                                       "maxRetries": 3,
                                       "maxFailovers": 3,
                                       "action": "rollback"
                               }
				}
			}`)

	req2, _ := http.NewRequest("POST", baseUrl+"/v1/apps", bytes.NewBuffer(jsonStr))
	req2.Header.Set("Content-Type", "application/json")

	resp, _ = http.DefaultClient.Do(req2)
	assert.Equal(t, resp.StatusCode, 200)
}

func TestListApplication(t *testing.T) {
	req, _ := http.NewRequest("GET", baseUrl+"/v1/apps", nil)

	resp, _ := http.DefaultClient.Do(req)
	assert.Equal(t, resp.StatusCode, 200)
}

func TestFetchApplication(t *testing.T) {
	req, _ := http.NewRequest("GET", baseUrl+"/v1/apps/11111", nil)

	resp, _ := http.DefaultClient.Do(req)
	assert.Equal(t, resp.StatusCode, 200)
}

func TestDeleteApplication(t *testing.T) {
	req, _ := http.NewRequest("DELETE", baseUrl+"/v1/apps/11111", nil)

	resp, _ := http.DefaultClient.Do(req)
	assert.Equal(t, resp.StatusCode, 200)
}

func TestListApplicationTasks(t *testing.T) {
	req, _ := http.NewRequest("GET", baseUrl+"/v1/apps/11111/tasks", nil)
	resp, _ := http.DefaultClient.Do(req)
	assert.Equal(t, resp.StatusCode, 200)
}

func TestDeleteApplicationTasks(t *testing.T) {
	req, _ := http.NewRequest("DELETE", baseUrl+"/v1/apps/11111/tasks", nil)
	resp, _ := http.DefaultClient.Do(req)
	assert.Equal(t, resp.StatusCode, 200)
}

func TestDeleteApplicationTask(t *testing.T) {
	req, _ := http.NewRequest("DELETE", baseUrl+"/v1/apps/11111/tasks/2222", nil)
	resp, _ := http.DefaultClient.Do(req)
	assert.Equal(t, resp.StatusCode, 200)
}

func TestListApplicationVersions(t *testing.T) {
	req, _ := http.NewRequest("GET", baseUrl+"/v1/apps/11111/versions", nil)
	resp, _ := http.DefaultClient.Do(req)
	assert.Equal(t, resp.StatusCode, 200)
}

func TestFetchApplicationVersion(t *testing.T) {
	req, _ := http.NewRequest("GET", baseUrl+"/v1/apps/11111/versions/222222", nil)
	resp, _ := http.DefaultClient.Do(req)
	assert.Equal(t, resp.StatusCode, 200)
}

func TestUpdateApplication(t *testing.T) {
	req, _ := http.NewRequest("POST", baseUrl+"/v1/apps/11111/update?instances=1", nil)

	resp, _ := http.DefaultClient.Do(req)
	assert.Equal(t, resp.StatusCode, 500)

	header := map[string][]string{
		"Content-Type": {"application/json"},
	}

	req.Header = header
	resp, _ = http.DefaultClient.Do(req)
	assert.Equal(t, resp.StatusCode, 500)

	var jsonStr = []byte(`{
  				"id": "nginx0002",
  				"cmd": null,
  				"cpus": 0.01,
  				"mem": 6,
  				"disk": 0,
  				"instances": 1,
  				"container": {
  				  "docker": {
  				    "image": "nginx:1.10",
  				    "network": "BRIDGE",
  				    "forcePullImage": false,
  				    "privileged": true,
  				    "parameters": [
  				      {
  				          "key": "label",
  				          "value": "APP_ID=nginx"
  				      }
  				    ],
      				    "portMappings": [
      				        {
      				            "containerPort": 80,
      				            "protocol": "tcp",
      				            "name": "web"
      				        }
      				    ]
    				},
    				"type": "DOCKER",
    				"volumes": [
    				  {
    				    "hostPath": "/home",
    				    "containerPath": "/data",
    				    "mode": "RW"
    				  }
    					]
  				},
  				"env": {
  				  "DB": "mysql"
  				},
  				"label": {
  				  "USER_ID": "1"
  				},
  				"killPolicy": {
  				  "duration": 5
  				},
			        "healthChecks": [
			          {
			            "protocol": "TCP",
			            "path": "/",
			            "portIndex": 0,
			            "gracePeriodSeconds": 5,
			            "intervalSeconds": 3,
			            "timeoutSeconds": 3,
			            "maxConsecutiveFailures": 3
			          }
			       ],
                               "updatePolicy": {
                                       "updateDelay": 30,
                                       "maxRetries": 3,
                                       "maxFailovers": 3,
                                       "action": "rollback"
                               }
				}
			}`)

	req2, _ := http.NewRequest("POST", baseUrl+"/v1/apps/1111/update?instances=5", bytes.NewBuffer(jsonStr))
	req2.Header.Set("Content-Type", "application/json")

	resp, _ = http.DefaultClient.Do(req2)
	assert.Equal(t, resp.StatusCode, 200)
}

func TestScaleApplication(t *testing.T) {
	req, _ := http.NewRequest("POST", baseUrl+"/v1/apps/1111/scale?instances=5", nil)

	resp, _ := http.DefaultClient.Do(req)
	assert.Equal(t, resp.StatusCode, 200)
}

func TestRollbackApplication(t *testing.T) {
	req, _ := http.NewRequest("POST", baseUrl+"/v1/apps/1111/rollback", nil)

	resp, _ := http.DefaultClient.Do(req)
	assert.Equal(t, resp.StatusCode, 200)
}
