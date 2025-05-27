package concretes 

import (
	"testing" 
)

func TestCalculateOutcomeMetrics(t *testing.T) {
	
	testCases := []struct {
		name           string
		goalsFor       int
		goalsAgainst   int
		expectedPoints int
		expectedWins   int
		expectedDraws  int
		expectedLosses int
	}{
		{
			name:           "Ev Sahibi Net Galibiyet",
			goalsFor:       3,
			goalsAgainst:   1,
			expectedPoints: 3,
			expectedWins:   1,
			expectedDraws:  0,
			expectedLosses: 0,
		},
		{
			name:           "Deplasman Net Galibiyet (Takım Kaybetti)",
			goalsFor:       0,
			goalsAgainst:   2,
			expectedPoints: 0,
			expectedWins:   0,
			expectedDraws:  0,
			expectedLosses: 1,
		},
		{
			name:           "Gollü Beraberlik",
			goalsFor:       2,
			goalsAgainst:   2,
			expectedPoints: 1,
			expectedWins:   0,
			expectedDraws:  1,
			expectedLosses: 0,
		},
		{
			name:           "Golsüz Beraberlik",
			goalsFor:       0,
			goalsAgainst:   0,
			expectedPoints: 1,
			expectedWins:   0,
			expectedDraws:  1,
			expectedLosses: 0,
		},
		{
			name:           "Tek Farkla Galibiyet",
			goalsFor:       1,
			goalsAgainst:   0,
			expectedPoints: 3,
			expectedWins:   1,
			expectedDraws:  0,
			expectedLosses: 0,
		},
		{
			name:           "Tek Farkla Mağlubiyet",
			goalsFor:       2,
			goalsAgainst:   3,
			expectedPoints: 0,
			expectedWins:   0,
			expectedDraws:  0,
			expectedLosses: 1,
		},
	}

	// Run each test case
	for _, tc := range testCases {
		//  t.run runs each scenario as a seperate subset
		t.Run(tc.name, func(t *testing.T) {
			points, wins, draws, losses := calculateOutcomeMetrics(tc.goalsFor, tc.goalsAgainst)

			if points != tc.expectedPoints {
				t.Errorf("Puan Hatalı: Beklenen %d, Alınan %d (Senaryo: %s, Skor: %d-%d)",
					tc.expectedPoints, points, tc.name, tc.goalsFor, tc.goalsAgainst)
			}
			if wins != tc.expectedWins {
				t.Errorf("Galibiyet Hatalı: Beklenen %d, Alınan %d (Senaryo: %s, Skor: %d-%d)",
					tc.expectedWins, wins, tc.name, tc.goalsFor, tc.goalsAgainst)
			}
			if draws != tc.expectedDraws {
				t.Errorf("Beraberlik Hatalı: Beklenen %d, Alınan %d (Senaryo: %s, Skor: %d-%d)",
					tc.expectedDraws, draws, tc.name, tc.goalsFor, tc.goalsAgainst)
			}
			if losses != tc.expectedLosses {
				t.Errorf("Mağlubiyet Hatalı: Beklenen %d, Alınan %d (Senaryo: %s, Skor: %d-%d)",
					tc.expectedLosses, losses, tc.name, tc.goalsFor, tc.goalsAgainst)
			}
		})
	}
}
