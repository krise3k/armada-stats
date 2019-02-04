package models

import "testing"

func TestGetOwner(t *testing.T) {
	mapping := []interface{}{}

	mapping = append(mapping, map[string]interface{}{"owner": "gd", "pattern": "gd_.*|.*_gd"})
	mapping = append(mapping, map[string]interface{}{"owner": "cerebro", "pattern": "crm-panel-v2|gauss-influx"})

	defaultOwner := "ops"

	testParams := []struct {
		service string
		owner   string
	}{
		{"gd_slots_tes", "gd"},
		{"slots_test_gd", "gd"},
		{"slots_gd_test", "ops"},
		{"gauss-influx", "cerebro"},
		{"ops_service", "ops"},
	}
	setMapping(mapping, defaultOwner)

	for _, service2owner := range testParams {
		owner := GetOwner(service2owner.service)
		if owner != service2owner.owner {
			t.Errorf("Not correct owner, service name: %s, got: %s, want: %s.", service2owner.service, owner, service2owner.owner)
		}
	}
}
