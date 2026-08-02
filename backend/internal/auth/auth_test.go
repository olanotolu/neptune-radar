package auth

import (
	"testing"

	"neptune-social-radar/backend/internal/store"
)

func TestScopeVisible(t *testing.T) {
	tests := []struct {
		role  store.Role
		scope string
		want  bool
	}{
		{store.RoleAdmin, "attorney_only", true},
		{store.RoleAttorney, "attorney_only", true},
		{store.RoleConcierge, "attorney_only", false},
		{store.RoleConcierge, "neptune_internal", true},
		{store.RoleConcierge, "shared_couple", true},
		{store.RoleConcierge, "unconfirmed_inference", true},
		{store.RoleConcierge, "", true},
	}
	for _, tt := range tests {
		t.Run(string(tt.role)+"_"+tt.scope, func(t *testing.T) {
			if got := ScopeVisible(tt.role, tt.scope); got != tt.want {
				t.Errorf("ScopeVisible(%s, %q) = %v, want %v", tt.role, tt.scope, got, tt.want)
			}
		})
	}
}
