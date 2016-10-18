package types

type Application struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Instances       int    `json:"instances"`
	InstanceUpdated int    `json:"instance_updated"`
	UserId          string `json:"user_id"`
	ClusterId       string `json:"cluster_id"`
	Status          string `json:"status"`
	Created         int64  `json:"created"`
	Updated         int64  `json:"updated"`
}
