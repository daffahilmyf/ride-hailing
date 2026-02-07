package domain

import "testing"

func TestRideTransitions(t *testing.T) {
	tests := []struct {
		name    string
		from    RideStatus
		next    RideStatus
		wantErr bool
	}{
		{"requested_to_matching", StatusRequested, StatusMatching, false},
		{"requested_to_cancelled", StatusRequested, StatusCancelled, false},
		{"matching_to_offered", StatusMatching, StatusOffered, false},
		{"offered_to_assigned", StatusOffered, StatusDriverAssigned, false},
		{"assigned_to_in_progress", StatusDriverAssigned, StatusInProgress, false},
		{"in_progress_to_completed", StatusInProgress, StatusCompleted, false},
		{"completed_to_cancelled", StatusCompleted, StatusCancelled, true},
		{"requested_to_completed", StatusRequested, StatusCompleted, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ride := Ride{Status: tt.from}
			_, err := ride.Transition(tt.next)
			if tt.wantErr && err == nil {
				t.Fatalf("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
