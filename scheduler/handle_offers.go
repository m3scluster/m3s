package scheduler

import (
	logrus "github.com/AVENTER-UG/mesos-m3s/logger"

	mesosproto "github.com/AVENTER-UG/mesos-m3s/proto"
	cfg "github.com/AVENTER-UG/mesos-m3s/types"
)

func (e *Scheduler) getAllOfferIDs(offers *mesosproto.Event_Offers) []mesosproto.OfferID {
	var offerIds []mesosproto.OfferID
	for _, offer := range offers.Offers {
		offerIds = append(offerIds, offer.ID)
	}

	return offerIds
}

// getOffer get out the offer for the mesos task
func (e *Scheduler) getOffer(offers *mesosproto.Event_Offers, cmd cfg.Command) (mesosproto.Offer, []mesosproto.OfferID) {
	var offerret mesosproto.Offer

	offerIds := e.getAllOfferIDs(offers)

	// if the constraints does not match, return an empty offer
	for n, offer := range offers.Offers {
		logrus.WithField("func", "scheduler.getOffer").Debug("Got Offer From:", offer.GetHostname(), " with offer ID:", offer.GetID())

		// Check Constraints of server, agent and datastore
		if cmd.TaskName == e.Framework.FrameworkName+":server" {
			if e.Config.K3SServerConstraintHostname != "" {
				if e.Config.K3SServerConstraintHostname == offer.GetHostname() {
					logrus.WithField("func", "scheduler.getOffer").Debug("Set Server Constraint to:", offer.GetHostname())
				} else {
					logrus.WithField("func", "scheduler.getOffer").Debug("Could not find hostname, get next offer")
					continue
				}
			}
		}
		if cmd.TaskName == e.Framework.FrameworkName+":agent" {
			if e.Config.K3SAgentConstraintHostname != "" {
				if e.Config.K3SAgentConstraintHostname == offer.GetHostname() {
					logrus.WithField("func", "scheduler.getOffer").Debug("Set Agent Constraint to:", offer.GetHostname())
				} else {
					logrus.WithField("func", "scheduler.getOffer").Debug("Could not find hostname, get next offer")
					continue
				}
			}
		}
		if cmd.TaskName == e.Framework.FrameworkName+":datastore" {
			if e.Config.DSConstraintHostname != "" {
				if e.Config.DSConstraintHostname == offer.GetHostname() {
					logrus.WithField("func", "scheduler.getOffer").Debug("Set Datastore Constraint to:", offer.GetHostname())
				} else {
					logrus.WithField("func", "scheduler.getOffer").Debug("Could not find hostname, get next offer")
					continue
				}
				offerret = offers.Offers[n]
			}
		}

		// if the ressources of this offer does not matched what the command need, the skip
		if !e.Mesos.IsRessourceMatched(offer.Resources, cmd) {
			logrus.WithField("func", "scheduler.getOffer").Debug("Could not found any matched resources, get next offer")
			continue
		}

		// offer match out needs, exit the loop
		offerret = offer
		break
	}

	// remove the offer we took
	if offerret.GetHostname() != "" {
		offerIds = e.removeOffer(offerIds, offerret.ID.Value)
	}
	e.Mesos.Call(e.Mesos.DeclineOffer(offerIds))
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
		// if the offer does not have id's, we skip it and restore the chan.
		if takeOffer.ID.Value == "" {
			logrus.WithField("func", "schueduler.HandleOffers").Error("OfferIds are empty.")
			e.Framework.CommandChan <- cmd
			return nil
		}
		logrus.WithField("func", "scheduler.HandleOffers").Info("Take Offer from " + takeOffer.GetHostname() + " for task " + cmd.TaskID + " (" + cmd.TaskName + ")")

		var taskInfo []mesosproto.TaskInfo
		RefuseSeconds := 5.0

		// #bugfix k3s 1.28.2 - if task is the server, add tls-san with the agents hostname
		if cmd.TaskName == e.Framework.FrameworkName+":server" {
			cmd.Command = cmd.Command + " --tls-san=" + takeOffer.GetHostname() + "'"
		}

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

		logrus.WithField("func", "scheduler.HandleOffers").Debug("Offer Accept: ", takeOffer.GetID(), " On Node: ", takeOffer.GetHostname())
		err := e.Mesos.Call(accept)
		if err != nil {
			logrus.WithField("func", "scheduler.HandleOffers").Error(err.Error())
			return err
		}
		// decline unneeded offer
		// logrus.WithField("func", "scheduler.HandleOffers").Debug("Offer Decline: ", offerIds)
		return e.Mesos.Call(e.Mesos.DeclineOffer(offerIds))

	default:
		offerIds = e.getAllOfferIDs(offers)
	}

	e.Mesos.Call(e.Mesos.DeclineOffer(offerIds))
	return nil
}
