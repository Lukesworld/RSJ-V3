package audit

import "testing"

func TestBayesianInference(t *testing.T) {
	// Test case from Python version:
	// Prior = 0.15, Evidence = 0.95 (High confidence flag)
	
	evidence := 0.95
	prior := 0.15
	
	// Expected calculation:
	// p_ev_recoverable = 0.95
	// p_ev_not_recoverable = 0.05
	// p_evidence = (0.95 * 0.15) + (0.05 * 0.85) = 0.1425 + 0.0425 = 0.185
	// posterior = (0.95 * 0.15) / 0.185 = 0.1425 / 0.185 = 0.77027...
	
	result := BayesianInference(evidence, prior)
	
	if result < 0.77 || result > 0.78 {
		t.Errorf("Expected result ~0.77, got %f", result)
	}
}

func TestRunSimulation(t *testing.T) {
	records := []Record{
		{ID: "TEST-1", Amt: 100.0, Flag: 0.99}, // Should trigger (conf > 0.90)
		{ID: "TEST-2", Amt: 50.0, Flag: 0.1},   // Should not trigger
	}
	
	// With prior 0.50, 0.99 flag yields high confidence > 0.90
	
	res := RunSimulation(records, 0.50)
	
	if res.RecoveredTotal != 100.0 {
		t.Errorf("Expected recovered total 100.0, got %f", res.RecoveredTotal)
	}
}
