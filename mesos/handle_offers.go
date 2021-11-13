package mesos

import (
	"github.com/sirupsen/logrus"

	mesosutil "github.com/AVENTER-UG/mesos-util"

	mesosproto "github.com/AVENTER-UG/mesos-util/proto"
)

// getOffer get out the offer for the mesos task
func getOffer(offers *mesosproto.Event_Offers, cmd mesosutil.Command) (mesosproto.Offer, []mesosproto.OfferID, bool) {
	offerIds := []mesosproto.OfferID{}
	var empty mesosproto.Offer
	count := 0
	for n, offer := range offers.Offers {
		logrus.Debug("Got Offer From:", offer.GetHostname())
		offerIds = append(offerIds, offer.ID)
		if cmd.TaskName == "k3sserver" {
			if config.K3SServerConstraintHostname != "" && config.K3SServerConstraintHostname == offer.GetHostname() {
				logrus.Debug("Set Server Constraint to:", offer.GetHostname())
				return offers.Offers[n], offerIds, true
			}
		}
		if cmd.TaskName == "k3sagent" {
			if config.K3SAgentConstraintHostname != "" && config.K3SAgentConstraintHostname == offer.GetHostname() {
				logrus.Debug("Set Agent Constraint to:", offer.GetHostname())
				return offers.Offers[n], offerIds, true
			}
		}
	}

	if (cmd.TaskName == "k3sserver") && (config.K3SServerConstraintHostname != "") || (cmd.TaskName == "k3sagent") && (config.K3SAgentConstraintHostname != "") {
		return empty, nil, false
	}
	return offers.Offers[count], offerIds, true
}

// HandleOffers will handle the offers event of mesos
func HandleOffers(offers *mesosproto.Event_Offers) error {
	_, offerIds, found := getOffer(offers, mesosutil.Command{})
	select {
	case cmd := <-framework.CommandChan:
		if cmd.TaskID == "" {
			return nil
		}
		var takeOffer mesosproto.Offer
		takeOffer, offerIds, found = getOffer(offers, cmd)
		logrus.Debug("Take Offer From:", takeOffer.GetHostname())

		if !found {
			return nil
		}

		var taskInfo []mesosproto.TaskInfo
		RefuseSeconds := 5.0

		taskInfo, _ = prepareTaskInfoExecuteContainer(takeOffer.AgentID, cmd)

		if cmd.TaskName == "k3sserver" {
			config.M3SBootstrapServerHostname = takeOffer.GetHostname()
			config.M3SBootstrapServerPort = int(cmd.DockerPortMappings[0].HostPort)
			config.K3SServerPort = int(cmd.DockerPortMappings[1].HostPort)
		}

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
		err := mesosutil.Call(accept)
		if err != nil {
			logrus.Error("Handle Offers: ", err)
		}

		// decline unneeded offer
		logrus.Info("Offer Decline: ", offerIds)
		return mesosutil.Call(mesosutil.DeclineOffer(offerIds))
	default:
		// decline unneeded offer
		logrus.Info("Decline unneeded offer: ", offerIds)
		return mesosutil.Call(mesosutil.DeclineOffer(offerIds))
	}
}
