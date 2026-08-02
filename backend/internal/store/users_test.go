package store

import "testing"

func TestCanAccessScope(t *testing.T) {
	tests := []struct {
		role  Role
		scope string
		want  bool
	}{
		{RoleAdmin, "attorney_only", true},
		{RoleAttorney, "attorney_only", true},
		{RoleConcierge, "attorney_only", false},
		{RoleAdmin, "shared_couple", true},
		{RoleConcierge, "shared_couple", true},
		{RoleAttorney, "neptune_internal", true},
		{RoleConcierge, "unconfirmed_inference", true},
		{RoleConcierge, "", true},
	}
	for _, tt := range tests {
		t.Run(string(tt.role)+"_"+tt.scope, func(t *testing.T) {
			if got := CanAccessScope(tt.role, tt.scope); got != tt.want {
				t.Errorf("CanAccessScope(%s, %q) = %v, want %v", tt.role, tt.scope, got, tt.want)
			}
		})
	}
}

func TestHashAPIKeyDeterministic(t *testing.T) {
	key := "npt_test123"
	h1 := HashAPIKey(key)
	h2 := HashAPIKey(key)
	if h1 != h2 {
		t.Error("HashAPIKey should be deterministic")
	}
	if h1 == key {
		t.Error("hash should not equal the plaintext")
	}
	if HashAPIKey("different") == h1 {
		t.Error("different keys should produce different hashes")
	}
}
