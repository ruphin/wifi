package wifi

import (
    "fmt"
    "sort"
    "math"
)

type fingerprint struct {
    signals Signals
    location *Location
}

type fingerprints []fingerprint

func (fingerprints fingerprints) locations() []*Location {
    locations := make([]*Location, len(fingerprints))
    for i, fingerprint := range fingerprints {
        locations[i] = fingerprint.location
    }
    return locations
}

func (fingerprints fingerprints) signals() []Signals {
    signals := make([]Signals, len(fingerprints))
    for i, fingerprint := range fingerprints {
        signals[i] = fingerprint.signals
    }
    return signals
}

type food struct {
    location *Location
    signals Signals
}

type fingerprinting struct {
    fingerprintMap map[Key]fingerprints
    food []food
    bestMatches int
    enhanced bool
    learning bool
    smart bool
}

func NewFingerprinting() algorithm {
    return algorithm(&fingerprinting{make(map[Key]fingerprints), nil, 4, false, false, false})
}

func NewEnhancedFingerprinting() algorithm {
    return algorithm(&fingerprinting{make(map[Key]fingerprints), nil, 4, true, false, false})
}

func NewLearningFingerprinting() algorithm {
    return algorithm(&fingerprinting{make(map[Key]fingerprints), nil, 4, false, true, false})
}

func NewEnhancedLearningFingerprinting() algorithm {
    return algorithm(&fingerprinting{make(map[Key]fingerprints), nil, 4, true, true, false})
}

func NewSmartLearningFingerprinting() algorithm {
    return algorithm(&fingerprinting{make(map[Key]fingerprints), nil, 4, false, true, true})
}

func (f *fingerprinting) feed(signals Signals, location *Location) {
    if f.enhanced && len(signals) == 0 {
        return
    }

    sort.Sort(ByID(signals))
    key := signals.Key()
    // if f.smart {
    //     var signalsList []Signals
    //     for mapKey, mapFingerprints := range f.fingerprintMap {
    //         if setDifference(key, mapKey) == 0 {
    //             signalsList = append(signalsList, mapFingerprints.signals() ...)
    //         }
    //     }
    //     for _, mapSignals := range signalsList {
    //         if euclidianDistance(signals, mapSignals) < 4 {
    //             return
    //         }
    //     }
    // }
    fingerprint := fingerprint{signals, location}
    f.fingerprintMap[key] = append(f.fingerprintMap[key], fingerprint)
}


func (f *fingerprinting) read(signals Signals, realLocation *Location) (*Location, bool) {
    pointerMap := make([]fingerprints, 50)
    sort.Sort(ByID(signals))
    ids := signals.Key()
    var dist int
    for key, fingerprints := range f.fingerprintMap {
        dist = setDifference(ids, key)
        pointerMap[dist] = append(pointerMap[dist], fingerprints...)
    }

    var locations []*Location
    var breakers fingerprints
    for distance, fingerprints := range pointerMap {
        if distance == len(signals) {
            break
        }
        if len(locations) + len(fingerprints) > f.bestMatches {
            breakers = fingerprints
            break
        }
        locations = append(locations, fingerprints.locations()...)
    }

    if locations == nil && breakers == nil {
        return nil, false
    } else {
        if breakers != nil {
            s := make([]stuf, len(breakers))
            for i, fingerprint := range breakers {
                s[i] = stuf{fingerprint.location, euclidianDistance(signals, fingerprint.signals)}
            }
            sort.Sort(ByDistance(s))

            for _, stuf := range s[:f.bestMatches - len(locations)] {
                locations = append(locations, stuf.location)
            }
        }
        location := average(locations)

        // if f.enhanced {
        //     location.enhance(realLocation)
        // }

        if f.learning {
            if f.smart {
                f.food = append(f.food, food{location, signals})
                if len(f.food) > 1000 {
                    fmt.Println("Batch feeding")
                    for _, food := range f.food {
                        f.feed(food.signals, food.location)
                    }
                    f.food = nil
                }
            } else {
                f.feed(signals, location)
            }
            
        }
        return location, true
    }
}

type stuf struct {
    location *Location
    distance float64
}

type ByDistance []stuf
func (s ByDistance) Len() int           { return len(s) }
func (s ByDistance) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s ByDistance) Less(i, j int) bool { return s[i].distance < s[j].distance }

func average(locations []*Location) *Location {
    var x, y float64 = 0, 0
    for _, location := range locations {
        x += location.X
        y += location.Y
    }
    return &Location{x / float64(len(locations)), y / float64(len(locations))}
}

// signals MUST be sorted by ID
func euclidianDistance(signals1, signals2 Signals) float64 {
    var diff float64
    var sum float64 = 0
    var i1, i2 int = 0, 0
    for i1 < len(signals1) && i2 < len(signals2) {
        if signals1[i1].id == signals2[i2].id {
            diff = signals1[i1].signalStrength - signals2[i2].signalStrength
            sum += diff * diff
            i1 += 1
            i2 += 1
        } else if signals1[i1].id < signals2[i2].id {
            i1 += 1
        } else {
            i2 += 1
        }
        
    }
    return math.Sqrt(sum)
}

// keys MUST be sorted
// Returns the number of elements that appear in ids1 but NOT in ids2
func setDifference(ids1, ids2 Key) int {
    i1 := 0
    i2 := 0
    distance := 0
    for i1 < keyLength && i2 < keyLength && ids1[i1] != 0 && ids2[i2] != 0 {
        
        if ids1[i1] < ids2[i2] {
            i1++
            distance++
        } else if ids1[i1] > ids2[i2] {
            i2++
        } else {
            i1++
            i2++
        }
    }

    for i1 < keyLength && ids1[i1] != 0 {
        i1++
        distance++
    }

    return distance
}

func spearman(vector1 []int, vector2 []int) float32 {
    if len(vector1) != len(vector2) {
        fmt.Println("Error, unequal vector lengths: %v %v", vector1, vector2)
    }
    n := len(vector1)
    tmp := make([]int, n)

    for i, value := range vector1 {
        d := value - vector2[i]
        if d < 0 {
            d = -d
        }
        tmp[i] = d * d
    }

    sum := 0
    for _, value := range tmp {
        sum += value
    }
    return (1 - float32(6 * sum) / float32(n * (n * n - 1)))
}


