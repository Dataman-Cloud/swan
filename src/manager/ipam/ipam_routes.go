package ipam

type Router struct {
	ipam *IPAM
}

// NewRouter initializes a new ipam router.
func NewRouter(manager *IPAM) *Router {
	r := &Router{
		ipam: manager,
	}

	return r
}

//func (r *Router) initRoutes() {
//r.routes = []*router.Route{
//router.NewRoute("GET", "/v1/ipam/allocate_randomly", r.AllocateNextAvailable),
//router.NewRoute("GET", "/v1/ipam/allocated_ips", r.ListAllocatedIps),
//router.NewRoute("GET", "/v1/ipam/available_ips", r.ListAvailableIps),
//router.NewRoute("POST", "/v1/ipam/release", r.ReleaseIP),
//router.NewRoute("GET", "/v1/ipam/allocate", r.AllocateIP),
//router.NewRoute("POST", "/v1/ipam/ips", r.RefillIPs),
//router.NewRoute("GET", "/v1/ipam/ips", r.ListIPs),
//}
//}
