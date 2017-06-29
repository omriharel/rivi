package autoassign

import (
	"github.com/bivas/rivi/bot"
	"github.com/bivas/rivi/util"
	"github.com/mitchellh/mapstructure"
	"math/rand"
)

type action struct {
	rule *rule
}

func findAssignedRoles(assignees []string, config bot.Configuration) []string {
	var assignedRoles []string
	if len(assignees) > 0 {
		assignedRolesSet := util.StringSet{}
		for _, role := range config.GetRoles() {
			for _, member := range config.GetRoleMembers(role) {
				for _, assignee := range assignees {
					if member == assignee {
						assignedRolesSet.Add(role)
					}
				}
			}
		}
		assignedRoles = assignedRolesSet.Values()
		util.Logger.Debug("There are %d assignees from %d roles", len(assignees), len(assignedRoles))
	}
	return assignedRoles
}

func (a *action) findLookupRoles(config bot.Configuration, assignedRoles []string) []string {
	lookupRoles := config.GetRoles()
	if len(a.rule.FromRoles) > 0 {
		lookupRoles = a.rule.FromRoles
	}
	return lookupRoles
}

func (a *action) Apply(config bot.Configuration, meta bot.EventData) {
	assignees := meta.GetAssignees()
	if len(assignees) > 0 && a.rule.IfNoAssignees {
		util.Logger.Debug("Skipping as there are assignees and no more are allowed")
		return
	}
	if len(assignees) >= a.rule.Require {
		util.Logger.Debug("Skipping as there are matched required assignees")
		return
	}
	assignedRoles := findAssignedRoles(assignees, config)
	lookupRoles := a.findLookupRoles(config, assignedRoles)

	winners := a.randomUsers(config, meta, lookupRoles)
	meta.AddAssignees(winners...)
}

func (a *action) randomUsers(config bot.Configuration, meta bot.EventData, lookupRoles []string) []string {
	possibleSet := util.StringSet{}
	possibleSet.AddAll(config.GetRoleMembers(lookupRoles...)).Remove(meta.GetOrigin())
	for _, assignee := range meta.GetAssignees() {
		possibleSet.Remove(assignee)
	}
	possible := possibleSet.Values()
	util.Logger.Debug("There are %d possible assignees from %d roles", len(possible), len(lookupRoles))
	remainingRequired := a.rule.Require - len(meta.GetAssignees())
	if remainingRequired < 0 {
		remainingRequired = 0
	}
	winners := make([]string, remainingRequired)
	for i := 0; i < remainingRequired; i++ {
		index := rand.Intn(len(possible))
		if possible[index] == "" {
			i--
		} else {
			winners[i] = possible[index]
			possible[index] = ""
		}
	}
	util.Logger.Debug("Selecting users %s as assignees", winners)
	return winners
}

type factory struct {
}

func (*factory) BuildAction(config map[string]interface{}) bot.Action {
	item := rule{}
	if e := mapstructure.Decode(config, &item); e != nil {
		panic(e)
	}
	item.Defaults()
	return &action{rule: &item}
}

func init() {
	bot.RegisterAction("autoassign", &factory{})
}
