package scheduler

import (
	"github.com/sirupsen/logrus"

	mesosproto "github.com/AVENTER-UG/mesos-m3s/proto"
	cfg "github.com/AVENTER-UG/mesos-m3s/types"
)

// getOffer get out the offer for the mesos task
func (e *Scheduler) getOffer(offers *mesosproto.Event_Offers, cmd cfg.Command) (mesosproto.Offer, []mesosproto.OfferID) {
	var offerIds []mesosproto.OfferID
	var offerret mesosproto.Offer
	if cmd.TaskName != "" {
		// if the constraints does not match, return an empty offer
		logrus.Debug("Get Offer for: ", cmd.TaskName)
		for n, offer := range offers.Offers {
			logrus.Debug("Got Offer From:", offer.GetHostname())
			offerIds = append(offerIds, offer.ID)

			// if the ressources of this offer does not matched what the command need, the skip
			if !e.Mesos.IsRessourceMatched(offer.Resources, cmd) {
				logrus.Debug("Could not found any matched ressources, get next offer")
				e.Mesos.Call(e.Mesos.DeclineOffer(offerIds))
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
	}

	// remove the offer we took
	offerIds = e.removeOffer(offerIds, offerret.ID.Value)
	return offerret, offerIds
}

// remove the offer we took from the list
func (e *Scheduler) removeOffer(offers []mesosproto.OfferID, clean string) []mesosproto.OfferID {
	var offerIds []mesosproto.OfferID
	for _, offer := range offers {
		if offer.Value != clean {
			offerIds = append(offerIds, offer)
		}
	}
	return offerIds
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
		if takeOffer.ID.Value == "" {
			logrus.WithField("func", "HandleOffers").Error("OfferIds are empty.")
			e.Framework.CommandChan <- cmd
			return nil
		}

		var taskInfo []mesosproto.TaskInfo
		RefuseSeconds := 5.0

		// build the mesos task info object with the current offer
		taskInfo = e.prepareTaskInfoExecuteContainer(takeOffer.AgentID, cmd)

		// remember information for the boostrap server to reach it later
		if cmd.TaskName == e.Framework.FrameworkName+":server" {
			e.Config.K3SServerHostname = takeOffer.GetHostname()
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
		err := e.Mesos.Call(accept)
		if err != nil {
			logrus.Error("Handle Offers: ", err)
			return err
		}
		// decline unneeded offer
		logrus.Info("Offer Decline: ", offerIds)
		return e.Mesos.Call(e.Mesos.DeclineOffer(offerIds))

	default:
		// decline unneeded offer
		_, offerIds := e.Mesos.GetOffer(offers, cfg.Command{})
		logrus.Info("Decline unneeded offer: ", offerIds)
		return e.Mesos.Call(e.Mesos.DeclineOffer(offerIds))
	}
}
