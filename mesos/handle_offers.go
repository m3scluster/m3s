package mesos

import (
	"github.com/sirupsen/logrus"

	mesosutil "github.com/AVENTER-UG/mesos-util"

	mesosproto "github.com/AVENTER-UG/mesos-util/proto"
)

// getOffer get out the offer for the mesos task
func getOffer(offers *mesosproto.Event_Offers, cmd mesosutil.Command) (mesosproto.Offer, []mesosproto.OfferID) {
	var offerIds []mesosproto.OfferID
	var offerret mesosproto.Offer
	// if the constraints does not match, return an empty offer
	logrus.Debug("Get Offer for: ", cmd.TaskName)
	for n, offer := range offers.Offers {
		logrus.Debug("Got Offer From:", offer.GetHostname())
		offerIds = append(offerIds, offer.ID)

		// if the ressources of this offer does not matched what the command need, the skip
		if !isRessourceMatched(offer.Resources, cmd) {
			logrus.Debug("Could not found any matched ressources, get next offer")
			mesosutil.Call(mesosutil.DeclineOffer(offerIds))
			continue
		}
		if cmd.TaskName == framework.FrameworkName+":server" {
			if config.K3SServerConstraintHostname == "" {
				offerret = offers.Offers[n]
			} else if config.K3SServerConstraintHostname == offer.GetHostname() {
				logrus.Debug("Set Server Constraint to:", offer.GetHostname())
				offerret = offers.Offers[n]
			}
		} else if cmd.TaskName == framework.FrameworkName+":agent" {
			if config.K3SAgentConstraintHostname == "" {
				offerret = offers.Offers[n]
			} else if config.K3SAgentConstraintHostname == offer.GetHostname() {
				logrus.Debug("Set Agent Constraint to:", offer.GetHostname())
				offerret = offers.Offers[n]
			}
		} else if cmd.TaskName == framework.FrameworkName+":etcd" {
			if config.ETCDConstraintHostname == "" {
				offerret = offers.Offers[n]
			} else if config.ETCDConstraintHostname == offer.GetHostname() {
				logrus.Debug("Set ETCD Constraint to:", offer.GetHostname())
				offerret = offers.Offers[n]
			}
		}
	}
	return offerret, offerIds
}

// HandleOffers will handle the offers event of mesos
func HandleOffers(offers *mesosproto.Event_Offers) error {
	var offerIds []mesosproto.OfferID
	select {
	case cmd := <-framework.CommandChan:
		// if no taskid or taskname is given, it's a wrong task.
		if cmd.TaskID == "" || cmd.TaskName == "" {
			return nil
		}
		var takeOffer mesosproto.Offer
		takeOffer, offerIds = getOffer(offers, cmd)
		if takeOffer.GetHostname() == "" {
			framework.CommandChan <- cmd
			return nil
		}
		logrus.Debug("Take Offer From:", takeOffer.GetHostname())

		if offerIds == nil {
			framework.CommandChan <- cmd
			return nil
		}

		var taskInfo []mesosproto.TaskInfo
		RefuseSeconds := 5.0

		taskInfo = prepareTaskInfoExecuteContainer(takeOffer.AgentID, cmd)

		if cmd.TaskName == framework.FrameworkName+":server" {
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
			return err
		}
		// decline unneeded offer
		logrus.Info("Offer Decline: ", offerIds)
		return mesosutil.Call(mesosutil.DeclineOffer(offerIds))

	default:
		// decline unneeded offer
		var empty mesosutil.Command
		_, offerIds = getOffer(offers, empty)
		logrus.Info("Decline unneeded offer: ", offerIds)
		return mesosutil.Call(mesosutil.DeclineOffer(offerIds))
	}
}

// check if the ressources of the offer are matching the needs of the cmd
func isRessourceMatched(ressource []mesosproto.Resource, cmd mesosutil.Command) bool {
	mem := false
	cpu := false

	for _, v := range ressource {
		if v.GetName() == "cpus" && v.Scalar.GetValue() >= cmd.CPU {
			logrus.Debug("Matched Offer CPU")
			cpu = true
		}
		if v.GetName() == "mem" && v.Scalar.GetValue() >= cmd.Memory {
			logrus.Debug("Matched Offer Memory")
			mem = true
		}
	}

	return mem && cpu
}
