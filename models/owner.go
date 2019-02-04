package models

import (
	"github.com/krise3k/armada-stats/utils"
	"regexp"
	"strings"
)

type OwnerMapping struct {
	Owner string
	Regex regexp.Regexp
}
type OwnerMappings []OwnerMapping

var ownerMappings OwnerMappings
var defaultOwner string

func GetOwner(containerName string) string {
	for _, ownerMapping := range getMapping() {
		if ownerMapping.Regex.MatchString(containerName) {
			return ownerMapping.Owner
		}
	}
	return defaultOwner
}

func getMapping() OwnerMappings {
	if ownerMappings == nil {
		raw, _ := utils.Config.List("ownership_mapping")
		owner, _ := utils.Config.String("default_owner")
		setMapping(raw, owner)
	}

	return ownerMappings
}

func setMapping(mapping []interface{}, owner string) {
	ownerMappings = generateOwnerMapping(mapping)
	defaultOwner = owner
}

func generateOwnerMapping(ownershipMapping []interface{}) OwnerMappings {
	var ownerMappings OwnerMappings

	for _, item := range ownershipMapping {
		item_map := item.(map[string]interface{})

		regex_str := item_map["pattern"].(string)
		owner := item_map["owner"].(string)
		if !strings.HasPrefix(regex_str, "^") {
			regex_str = "^" + regex_str
		}
		if !strings.HasSuffix(regex_str, "$") {
			regex_str = regex_str + "$"
		}
		regex, err := regexp.Compile(regex_str)
		if err != nil {
			utils.GetLogger().WithError(err).Panic("Can't create owner to service regex mapping")
		}
		ownerMapping := OwnerMapping{Regex: *regex, Owner: owner}
		ownerMappings = append(ownerMappings, ownerMapping)
	}
	return ownerMappings
}
