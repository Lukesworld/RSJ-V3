package audit

// BayesianInference calculates the posterior probability of a claim being recoverable.
//
// evidenceStrength is the probability of the evidence given the claim is recoverable.
// priorProbability is the prior probability of the claim being recoverable.
func BayesianInference(evidenceStrength float64, priorProbability float64) float64 {
	pEvidenceGivenRecoverable := evidenceStrength
	pEvidenceGivenNotRecoverable := 0.05
	pEvidence := (pEvidenceGivenRecoverable * priorProbability) +
		(pEvidenceGivenNotRecoverable * (1 - priorProbability))
	
	if pEvidence == 0 {
		return 0 // Avoid division by zero
	}

	return (pEvidenceGivenRecoverable * priorProbability) / pEvidence
}

type Record struct {
	ID   string
	Desc string
	Amt  float64
	Flag float64
}

type SimulationResult struct {
	RecoveredTotal float64
	RecoverableFee float64
}

func RunSimulation(records []Record, prior float64) SimulationResult {
	var totalRecovered float64

	for _, rec := range records {
		prob := BayesianInference(rec.Flag, prior)
		if prob > 0.90 {
			totalRecovered += rec.Amt
		}
	}

	return SimulationResult{
		RecoveredTotal: totalRecovered,
		RecoverableFee: totalRecovered * 0.30,
	}
}
