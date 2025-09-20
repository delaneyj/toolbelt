package toolbelt

type MusicalRatioName string

const (
	MusicalRatioMinorSecond     MusicalRatioName = "Minor Second"
	MusicalRatioMajorSecond     MusicalRatioName = "Major Second"
	MusicalRatioMinorThird      MusicalRatioName = "Minor Third"
	MusicalRatioMajorThird      MusicalRatioName = "Major Third"
	MusicalRatioPerfectFourth   MusicalRatioName = "Perfect Fourth"
	MusicalRatioAugmentedFourth MusicalRatioName = "Augmented Fourth"
	MusicalRatioPerfectFifth    MusicalRatioName = "Perfect Fifth"
	MusicalRatioGoldenRatio     MusicalRatioName = "Golden Ratio"
	MusicalRatioMinorSixth      MusicalRatioName = "Minor Sixth"
	MusicalRatioMajorSixth      MusicalRatioName = "Major Sixth"
	MusicalRatioMajorSeventh    MusicalRatioName = "Major Seventh"
	MusicalRatioOctave          MusicalRatioName = "Octave"
)

var musicalRatios = map[MusicalRatioName]float64{
	MusicalRatioMinorSecond:     1.067,
	MusicalRatioMajorSecond:     1.125,
	MusicalRatioMinorThird:      1.200,
	MusicalRatioMajorThird:      1.250,
	MusicalRatioPerfectFourth:   1.333,
	MusicalRatioAugmentedFourth: 1.414,
	MusicalRatioPerfectFifth:    1.500,
	MusicalRatioGoldenRatio:     1.618,
	MusicalRatioMinorSixth:      1.667,
	MusicalRatioMajorSixth:      1.778,
	MusicalRatioMajorSeventh:    1.875,
	MusicalRatioOctave:          2.000,
}

func MusicalRatio(name MusicalRatioName) float64 {
	return musicalRatios[name]
}
