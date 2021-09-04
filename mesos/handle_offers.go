package mesos

import (
	"github.com/sirupsen/logrus"

	cfg "github.com/AVENTER-UG/mesos-m3s/types"

	mesosproto "github.com/AVENTER-UG/mesos-m3s/proto"
)

func defaultResources(cmd cfg.Command) []mesosproto.Resource {
	CPU := "cpus"
	MEM := "mem"
	cpu := cmd.CPU
	mem := cmd.Memory
	PORT := "ports"

	res := []mesosproto.Resource{
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
	}

	var portBegin, portEnd uint64

	if cmd.DockerPortMappings != nil {
		portBegin = uint64(cmd.DockerPortMappings[0].HostPort)
		portEnd = portBegin + 2

		res = []mesosproto.Resource{
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
				Ranges: &mesosproto.Value_Ranges{
					Range: []mesosproto.Value_Range{
						{
							Begin: portBegin,
							End:   portEnd,
						},
					},
				},
			},
		}
	}

	return res
}

// getOffer get out the offer for the mesos task
func getOffer(offers *mesosproto.Event_Offers, cmd cfg.Command) (mesosproto.Offer, []mesosproto.OfferID) {
	offerIds := []mesosproto.OfferID{}
	count := 0
	for n, offer := range offers.Offers {
		logrus.Debug("Got Offer From:", offer.GetHostname())
		offerIds = append(offerIds, offer.ID)
		if cmd.IsK3SServer {
			if config.K3SServerConstraintHostname != "" && config.K3SServerConstraintHostname == offer.GetHostname() {
				logrus.Debug("Set Server Constraint to:", offer.GetHostname())
				count = n
			}
		}
		if cmd.IsK3SAgent {
			if config.K3SAgentConstraintHostname != "" && config.K3SAgentConstraintHostname == offer.GetHostname() {
				logrus.Debug("Set Agent Constraint to:", offer.GetHostname())
				count = n
			}
		}
	}

	return offers.Offers[count], offerIds

}

// HandleOffers will handle the offers event of mesos
func HandleOffers(offers *mesosproto.Event_Offers) error {
	offerIds := []mesosproto.OfferID{}

	select {
	case cmd := <-config.CommandChan:

		takeOffer, offerIds := getOffer(offers, cmd)
		logrus.Debug("Take Offer From:", takeOffer.GetHostname())

		var taskInfo []mesosproto.TaskInfo
		RefuseSeconds := 5.0

		// if its the K3SServer, remember the mesos agents hostename and hostport
		if cmd.IsK3SServer {
			config.M3SBootstrapServerHostname = takeOffer.GetHostname()
			config.M3SBootstrapServerPort = int(cmd.DockerPortMappings[0].HostPort)
			config.K3SServerPort = int(cmd.DockerPortMappings[1].HostPort)
		}

		logrus.Debug("Schedule Command: ", cmd.Command)

		taskInfo, _ = prepareTaskInfoExecuteContainer(takeOffer.AgentID, cmd)

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
		err := Call(accept)
		if err != nil {
			logrus.Error("Handle Offers: ", err)
		}

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
		return Call(decline)

	}
}
