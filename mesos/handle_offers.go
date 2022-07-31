package mesos

import (
	"github.com/sirupsen/logrus"

	mesosutil "github.com/AVENTER-UG/mesos-util"

	mesosproto "github.com/AVENTER-UG/mesos-util/proto"
)

// getOffer get out the offer for the mesos task
func (e *Scheduler) getOffer(offers *mesosproto.Event_Offers, cmd mesosutil.Command) (mesosproto.Offer, []mesosproto.OfferID) {
	var offerIds []mesosproto.OfferID
	var offerret mesosproto.Offer
	// if the constraints does not match, return an empty offer
	logrus.Debug("Get Offer for: ", cmd.TaskName)
	for n, offer := range offers.Offers {
		logrus.Debug("Got Offer From:", offer.GetHostname())
		offerIds = append(offerIds, offer.ID)

		// if the ressources of this offer does not matched what the command need, the skip
		if !e.isRessourceMatched(offer.Resources, cmd) {
			logrus.Debug("Could not found any matched ressources, get next offer")
			mesosutil.Call(mesosutil.DeclineOffer(offerIds))
			continue
		}
		// Check Constraints of server, agent and datastore
		if cmd.TaskName == e.Framework.FrameworkName+":server" {
			if e.Config.K3SServerConstraintHostname == "" {
				offerret = offers.Offers[n]
			} else if e.Config.K3SServerConstraintHostname == offer.GetHostname() {
				logrus.Debug("Set Server Constraint to:", offer.GetHostname())
				offerret = offers.Offers[n]
			}
		}
		if cmd.TaskName == e.Framework.FrameworkName+":agent" {
			if e.Config.K3SAgentConstraintHostname == "" {
				offerret = offers.Offers[n]
			} else if e.Config.K3SAgentConstraintHostname == offer.GetHostname() {
				logrus.Debug("Set Agent Constraint to:", offer.GetHostname())
				offerret = offers.Offers[n]
			}
		}
		if cmd.TaskName == e.Framework.FrameworkName+":datastore" {
			if e.Config.DSConstraintHostname == "" {
				offerret = offers.Offers[n]
			} else if e.Config.DSConstraintHostname == offer.GetHostname() {
				logrus.Debug("Set Datastore Constraint to:", offer.GetHostname())
				offerret = offers.Offers[n]
			}
		}
	}
	return offerret, offerIds
}

// HandleOffers will handle the offers events of mesos
func (e *Scheduler) HandleOffers(offers *mesosproto.Event_Offers) error {
	var offerIds []mesosproto.OfferID
	select {
	case cmd := <-e.Framework.CommandChan:
		// if no taskid or taskname is given, it's a wrong task.
		if cmd.TaskID == "" || cmd.TaskName == "" {
			return nil
		}
		var takeOffer mesosproto.Offer
		// if the offer the take does not have a hostname, we skip it and restore the chan.
		takeOffer, offerIds = e.getOffer(offers, cmd)
		if takeOffer.GetHostname() == "" {
			e.Framework.CommandChan <- cmd
			return nil
		}
		logrus.Debug("Take Offer From:", takeOffer.GetHostname())
		// if the offer does not have id's, we skip it and restore the chan.
		if offerIds == nil {
			e.Framework.CommandChan <- cmd
			return nil
		}

		var taskInfo []mesosproto.TaskInfo
		RefuseSeconds := 5.0

		// build the mesos task info object with the current offer
		taskInfo = e.prepareTaskInfoExecuteContainer(takeOffer.AgentID, cmd)

		// remember information for the boostrap server to reach it later
		if cmd.TaskName == e.Framework.FrameworkName+":server" {
			e.Config.M3SBootstrapServerHostname = takeOffer.GetHostname()
			e.Config.M3SBootstrapServerPort = int(cmd.DockerPortMappings[0].HostPort)
			e.Config.K3SServerPort = int(cmd.DockerPortMappings[1].HostPort)
		}

		// build mesos call object
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
		_, offerIds = e.getOffer(offers, empty)
		logrus.Info("Decline unneeded offer: ", offerIds)
		return mesosutil.Call(mesosutil.DeclineOffer(offerIds))
	}
}

// check if the ressources of the offer are matching the needs of the cmd
func (e *Scheduler) isRessourceMatched(ressource []mesosproto.Resource, cmd mesosutil.Command) bool {
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
