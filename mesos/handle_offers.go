package mesos

import (
	"github.com/sirupsen/logrus"

	mesosproto "mesos-k3s/proto"
	cfg "mesos-k3s/types"
)

func defaultResources(cmd cfg.Command) []mesosproto.Resource {
	CPU := "cpus"
	MEM := "mem"
	cpu := config.ResCPU
	mem := config.ResMEM
	PORT := "ports"

	var portBegin, portEnd uint64
	portBegin = 31210 + uint64(cmd.TaskID)
	portEnd = 31210 + uint64(cmd.TaskID)

	return []mesosproto.Resource{
		{
			Name:   CPU,
			Type:   mesosproto.SCALAR.Enum(),
			Scalar: &mesosproto.Value_Scalar{Value: cpu},
		},
		{
			Name:   MEM,
			Type:   mesosproto.SCALAR.Enum(),
			Scalar: &mesosproto.Value_Scalar{Value: mem},
		},
		{
			Name: PORT,
			Type: mesosproto.RANGES.Enum(),
			Ranges: &mesosproto.Value_Ranges{Range: []mesosproto.Value_Range{{
				Begin: portBegin,
				End:   portEnd,
			}}},
		},
	}
}

// HandleOffers will handle the offers event of mesos
func HandleOffers(offers *mesosproto.Event_Offers) error {
	offerIds := []mesosproto.OfferID{}
	var count int
	for a, offer := range offers.Offers {
		offerIds = append(offerIds, offer.ID)
		count = a
		logrus.Debug("Got Offer From:", offer.GetHostname())
	}

	select {
	case cmd := <-config.CommandChan:

		takeOffer := offers.Offers[count]

		var taskInfo []mesosproto.TaskInfo
		RefuseSeconds := 5.0

		switch cmd.ContainerType {
		case "MESOS":
			taskInfo, _ = prepareTaskInfoExecuteContainer(takeOffer.AgentID, cmd)
		case "DOCKER":
			taskInfo, _ = prepareTaskInfoExecuteContainer(takeOffer.AgentID, cmd)
		}

		logrus.Debug("HandleOffers cmd: ", taskInfo)

		accept := &mesosproto.Call{
			Type: mesosproto.Call_ACCEPT,
			Accept: &mesosproto.Call_Accept{
				OfferIDs: []mesosproto.OfferID{{
					Value: takeOffer.ID.Value,
				}},
				Filters: &mesosproto.Filters{
					RefuseSeconds: &RefuseSeconds,
				},
				Operations: []mesosproto.Offer_Operation{{
					Type: mesosproto.Offer_Operation_LAUNCH,
					Launch: &mesosproto.Offer_Operation_Launch{
						TaskInfos: taskInfo,
					}}}}}

		logrus.Info("Offer Accept: ", takeOffer.GetID(), " On Node: ", takeOffer.GetHostname())
		Call(accept)

		// decline unneeded offer
		logrus.Info("Offer Decline: ", offerIds)
		decline := &mesosproto.Call{
			Type:    mesosproto.Call_DECLINE,
			Decline: &mesosproto.Call_Decline{OfferIDs: offerIds},
		}
		return Call(decline)
	default:
		// decline unneeded offer
		logrus.Info("Offer Decline: ", offerIds)
		decline := &mesosproto.Call{
			Type:    mesosproto.Call_DECLINE,
			Decline: &mesosproto.Call_Decline{OfferIDs: offerIds},
		}
		Call(decline)

		// tell mesos he dont have to offer again until we ask
		logrus.Info("Framework Suppress: ", offerIds)
		suppress := &mesosproto.Call{
			Type: mesosproto.Call_SUPPRESS,
		}
		return Call(suppress)
	}
}
