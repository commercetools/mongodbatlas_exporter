package measurer

import "testing"

func TestProcessInfoLabels(t *testing.T) {
	b := Process{
		Base: Base{
			ProjectID: "9u1ji2k3jlj",
			RsName:    "rs1",
			UserAlias: "cluster-rs1:27017",
			TypeName:  "REPLICA_PRIMARY",
			Hostname:  "cluster-rs1",
			ID:        "kj1lk2jklji",
		},
		Version: "4.2.17",
	}

	//if more labels are needed make sure to include
	//a reason why here.
	allowedLabels := map[string]bool{
		"project_id": true, //0.0.2
		"rs_name":    true, //0.0.2
		"user_alias": true, //0.0.2
		"version":    true, //0.0.2
		"type":       true, //0.0.2
	}

	labels := b.PromInfoConstLabels()

	for k := range allowedLabels {
		if _, ok := labels[k]; !ok {
			t.Fatalf("missing label %s", k)
		}
	}

	for k := range labels {
		if _, ok := allowedLabels[k]; !ok {
			t.Fatalf("extra label %s", k)
		}
	}
}
