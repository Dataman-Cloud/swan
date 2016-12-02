package engine

import (
	//"github.com/Dataman-Cloud/swan/src/mesosproto/mesos"
	//"github.com/Dataman-Cloud/swan/src/mesosproto/sched"

	"github.com/Sirupsen/logrus"
	//"github.com/golang/protobuf/proto"
)

func OfferHandler(h *Handler) *Handler {
	logrus.WithFields(logrus.Fields{"handler": "offer"}).Debugf("")

	//call := &sched.Call{
	//FrameworkId: s.framework.GetId(),
	//Type:        sched.Call_DECLINE.Enum(),
	//Decline: &sched.Call_Decline{
	//OfferIds: []*mesos.OfferID{
	//{
	//Value: offerId,
	//},
	//},
	//Filters: &mesos.Filters{
	//RefuseSeconds: proto.Float64(1),
	//},
	//},
	//}

	return h
}
