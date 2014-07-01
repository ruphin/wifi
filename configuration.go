package wifi

// Configuration options
type Configuration struct {
    // The height of the map in meters
    MapHeight float64

    // The width of the map in meters
    MapWidth float64

    // The amount of accesspoints on the map in ap/km^2
    AccessPointDensity int

    // The distance between seed readings in m
    SeedDistance float64

    // The distance between test readings in m
    TestDistance float64

    // The number of test cycles to execute in this test
    TestCycles int

    // The amount of access points to be replaced after every test cycle, expressed between 0 and 1
    ReplacementRate float64

    // The strategy to use when replacing access points. Can be FiFoReplacement or RandomReplacement
    ReplacementStrategy int

    // The random seed to use when generating test readings. When set to 0, a random value is generated and used instead.
    RandomSeed int64

    // The directory to save graphs and images
    OutputDir string
}

// Values for Replacementstrategy configuration
const (
    FiFoReplacement = 1
    RandomReplacement = 2
)

func NewConfiguration() *Configuration {
    return &Configuration{}
}

func (config *Configuration) validate() {
    if config.ReplacementRate != 0 && config.ReplacementStrategy == 0 {
        err_string := "** err: Replacement Rate set without a Replacement Strategy"
        panic(err_string)
    }
    if config.AccessPointDensity == 0 {
        err_string := "** err: Access Point Density cannot be 0"
        panic(err_string)
    }
    if config.MapWidth < 160 {
        err_string := "** err: Map Width cannot be less than 160 meters"
        panic(err_string)
    }
    if config.MapHeight < 160 {
        err_string := "** err: Map Height cannot be less than 160 meters"
        panic(err_string)
    }
    if config.SeedDistance == 0 {
        err_string := "** err: Seed Distance cannot be 0"
        panic(err_string)
    }
    if config.TestDistance == 0 {
        err_string := "** err: Test Distance cannot be 0"
        panic(err_string)
    }
}