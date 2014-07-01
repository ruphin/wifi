package wifi

import (
    "math/rand"
    "fmt"
    "github.com/ruphin/go-gnuplot/pkg/gnuplot"
    "os"
    "strings"
)

type algorithm interface {
    feed(signals Signals, location *Location)
    read(signals Signals, location *Location) (*Location, bool)
}

type engine struct {
    m *Map
    algorithms map[string]algorithm
    accessPointGenerations []int
    config *Configuration
}

// Create a new Engine from the given configuration.
func NewEngine(config *Configuration) *engine {

    config.validate()

    engineMap := NewMap(config.MapWidth, config.MapHeight, config.RandomSeed)
    engine := &engine{engineMap, make(map[string]algorithm), make([]int, 0), config}

    accessPointCount := int(config.MapWidth * config.MapHeight) * config.AccessPointDensity / 1000000
    for i := 0; i < accessPointCount; i++ {
        engine.m.AddRandomAccessPoint()
    }

    engine.accessPointGenerations = append(engine.accessPointGenerations, accessPointCount)
    engine.m.Draw(engine.accessPointGenerations)
    return engine
}

func (e *engine) AddAlgorithm(name string, algorithm algorithm) {
    e.algorithms[name] = algorithm
}

func (e *engine) Run() {
    var testCycles = e.config.TestCycles
    var mapWidth = e.config.MapWidth
    var mapHeight = e.config.MapHeight
    var testMinWidth = mapWidth / 2 - 250.00
    var testMaxWidth = mapWidth / 2 + 250.00
    var testMinHeight = mapHeight / 2 - 250.00
    var testMaxHeight = mapHeight / 2 + 250.00
    var testDistance = e.config.TestDistance

    e.seed()

    // Initialize testing locations.
    // Locations are tested on a grid with specified testing distance.
    // No tests are performed within 75 meters of the edge of the map, to ensure access points can be found in all directions
    var locations []*Location
    for x := 85.0; x <= float64(mapWidth) - 85; x += testDistance {
        for y := 85.0; y <= float64(mapHeight) - 85; y += testDistance {
            locations =  append(locations, NewLocation(x,y))
        }
    }

    // Maps to record cumulative Error and miss counts per algorithm per cycle
    algorithmErrors := make(map[string][][]float64)
    algorithmMisses := make(map[string][]float64)
    centerAlgorithmErrors := make(map[string][][]float64)
    centerAlgorithmMisses := make(map[string][]float64)
    for name, _ := range e.algorithms {
        algorithmErrors[name] = make([][]float64, testCycles + 1)
        algorithmMisses[name] = make([]float64, testCycles + 1)
        centerAlgorithmErrors[name] = make([][]float64, testCycles + 1)
        centerAlgorithmMisses[name] = make([]float64, testCycles + 1)
    }

    var signals Signals
    var success bool
    var estimatedLocation *Location

    fmt.Printf("Starting simulation\n")
    fmt.Printf("Performing %d localizations in each of %d cycles\n\n", len(locations), testCycles)

    // Run the cycles
    for cycle := 0; cycle <= testCycles; cycle++ {

        // Before every cycle except the first, replace access points
        if cycle != 0 {
            e.replaceAccessPoints()
        }

        // Randomize the order of testing locations
        for i := range locations {
            j := rand.Intn(i + 1)
            locations[i], locations[j] = locations[j], locations[i]
        }

        // For every location, test each algorithm
        for _, location := range locations {
            signals = e.m.Read(location)
            for name, algorithm := range e.algorithms {
                estimatedLocation, success = algorithm.read(signals, location)
                if location.X >= testMinWidth && location.X <= testMaxWidth && location.Y >= testMinHeight && location.Y <= testMaxHeight {
                    if success {
                        centerAlgorithmErrors[name][cycle] =  append(centerAlgorithmErrors[name][cycle], distance(location, estimatedLocation))
                    } else {
                        centerAlgorithmMisses[name][cycle] += 1
                    }
                }
                if success {
                    algorithmErrors[name][cycle] =  append(algorithmErrors[name][cycle], distance(location, estimatedLocation))
                } else {
                    algorithmMisses[name][cycle] += 1
                }
            }
        }
        fmt.Printf("Completed tests for cycle %2d\n", cycle)
    }

    fmt.Printf("\nSimulation completed. Generating Graphs...\n")
    e.plot(centerAlgorithmErrors, centerAlgorithmMisses, "Center")
    e.plot(algorithmErrors, algorithmMisses, "Full")
    e.drawLastFrame()
    e.drawLastFrameCenter()
}

// Seed the Algorithms with initial data.
func (e *engine) seed() {
    mapWidth := e.config.MapWidth
    mapHeight := e.config.MapHeight
    distance := e.config.SeedDistance

    var location *Location
    var signals Signals
    for x := 0.0; x <= mapWidth; x += distance {
        for y := 0.0; y <= mapHeight; y += distance {
            location = NewLocation(x, y)
            signals = e.m.Read(location)
            for _, algorithm := range e.algorithms {
                algorithm.feed(signals, location)
            }
        }
    }
}

// Replace accesspoints
func (e *engine) replaceAccessPoints() {
    replacementRate := e.config.ReplacementRate
    replacementStrategy := e.config.ReplacementStrategy

    replacementCount := int(float64(len(e.m.accessPoints)) * replacementRate)

    for i := 0; i < replacementCount; i++ {
        if replacementStrategy == FiFoReplacement {
            e.m.RemoveOldestAccessPoint()
        } else if replacementStrategy == RandomReplacement {
            e.m.RemoveRandomAccessPoint()
        } else {
            err_string := "** err: Unknown Replacement Strategy"
            panic(err_string)
        }
    }

    for i := 0; i < replacementCount; i++ {
        e.m.AddRandomAccessPoint()
    }

    e.accessPointGenerations = append(e.accessPointGenerations, e.m.accessPoints[len(e.m.accessPoints)-1].id)

    // accessPointCounts := make([]int, len(e.accessPointGenerations))
    // for _, accessPoint := range e.m.accessPoints {
    //     for i, generationLimit := range e.accessPointGenerations {
    //         if accessPoint.id <= generationLimit {
    //             accessPointCounts[i] += 1
    //             break
    //         }
    //     }
    // }
    //
    // fmt.Printf("Access point counts: %v\n", accessPointCounts)
    //
    // fmt.Printf("Access point generations: %v\n", e.accessPointGenerations)

    e.m.Draw(e.accessPointGenerations)
}

func (e *engine) plot(algorithmErrors map[string][][]float64, algorithmMisses map[string][]float64, plottype string) {
    testCycles := e.config.TestCycles
    directory := e.config.OutputDir
    perCycle := make([]float64, testCycles + 1)
    for i := 0; i <= testCycles; i++ {
        perCycle[i] = float64(i)
    }

    // Initialize the plotters
    fname := ""
    persist := false
    debug := false

    errorPlot,err := gnuplot.NewPlotter(fname, persist, debug)
    if err != nil {
        err_string := fmt.Sprintf("** err: %v\n", err)
        panic(err_string)
    }
    defer errorPlot.Close()

    missPlot,err := gnuplot.NewPlotter(fname, persist, debug)
    if err != nil {
        err_string := fmt.Sprintf("** err: %v\n", err)
        panic(err_string)
    }
    defer missPlot.Close()

    errorPlot.CheckedCmd(fmt.Sprintf("set xrange [0:%d]", testCycles))
    errorPlot.CheckedCmd("set datafile missing 'NaN'")
    errorPlot.CheckedCmd("set key left top")
    errorPlot.CheckedCmd("set yrange [0:60]")
    missPlot.CheckedCmd(fmt.Sprintf("set xrange [0:%d]", testCycles))
    missPlot.CheckedCmd("set key left top")
    missPlot.CheckedCmd("set yrange [0:100]")

    graphStyle := []int{4,6,8,12,5,7,9,13}
    graph := 0

    var errors []float64
    var misses []float64
    var sum float64
    var hits float64
    for name, _ := range e.algorithms {
        misses = algorithmMisses[name]
        errors = make([]float64, 0)
        for i, _ := range algorithmMisses[name] {
            sum = 0
            hits = float64(len(algorithmErrors[name][i]))
            for j := 0; j < int(hits); j++ {
                sum += algorithmErrors[name][i][j]
            }
            errors =  append(errors, sum / hits)
            misses[i] = misses[i] / (misses[i] + hits) * 100
        }

        errorPlot.SetStyle(fmt.Sprintf("linespoints lt 1 lw 2 pi 2 pt %d linecolor rgb 'black'", graphStyle[graph]))
        errorPlot.PlotXY(perCycle, errors, name)

        missPlot.SetStyle(fmt.Sprintf("linespoints lt 1 lw 2 pi 2 pt %d linecolor rgb 'black'", graphStyle[graph]))
        missPlot.PlotXY(perCycle, misses, name)
        graph += 1
        fmt.Println(fmt.Sprintf("%v Errors: %v", name, errors))
    }

    if os.MkdirAll(directory, 0777) != nil {
        panic("Unable to create directory for graphs")
    }
    algorithms := ""
    for name, _ := range e.algorithms {
        if strings.Contains(name, "Enhanced") {
            algorithms = strings.Join([]string{algorithms, "E"}, "")
        }
        if strings.Contains(name, "Learning") {
            algorithms = strings.Join([]string{algorithms, "L"}, "")
        }
        if strings.Contains(name, "Centroid") {
            algorithms = strings.Join([]string{algorithms, "C"}, "")
        }
        if strings.Contains(name, "Fingerprinting") {
            algorithms = strings.Join([]string{algorithms, "F"}, "")
        }
    }

    filename := fmt.Sprintf("dens%d-dist%.0f-Cyc%d-%d-%d-%v", e.config.AccessPointDensity / 100, e.config.SeedDistance, e.config.TestCycles, int(e.config.ReplacementRate * 100), e.config.ReplacementStrategy, algorithms)

    errorPlot.SetXLabel("Cycles")
    errorPlot.SetYLabel("Average Error")
    errorPlot.CheckedCmd("set terminal pdf")
    errorPlot.CheckedCmd(fmt.Sprintf("set output '%v/%v-errors-%v.pdf'", directory, filename, plottype))
    errorPlot.CheckedCmd("replot")

    missPlot.SetXLabel("Cycles")
    missPlot.SetYLabel("Miss Percentage")
    missPlot.CheckedCmd("set terminal pdf")
    missPlot.CheckedCmd(fmt.Sprintf("set output '%v/%v-misses-%v.pdf'", directory, filename, plottype))
    missPlot.CheckedCmd("replot")
}

func (e *engine) drawLastFrame() {
    mapWidth := e.config.MapWidth
    mapHeight := e.config.MapHeight
    directory := e.config.OutputDir
    distance := (mapWidth - 170) / 6

    sources := make(map[string][]*Location)
    results := make(map[string][]*Location)

    var location *Location
    var result *Location
    var signals Signals
    var success bool
    for x := 85.0 ; x <= mapWidth - 84.0; x += distance {
        for y := 85.0; y <= mapHeight - 84.0; y += distance {
            location = NewLocation(x,y)
            signals = e.m.Read(location)
            for name, algorithm := range e.algorithms {
                result, success = algorithm.read(signals, location)
                if success {
                    sources[name] = append(sources[name], location)
                    results[name] = append(results[name], result)
                }
            }
        }
    }

    for name, _ := range e.algorithms {
        // Initialize the plotters
        fname := ""
        persist := false
        debug := false

        p,err := gnuplot.NewPlotter(fname, persist, debug)
        if err != nil {
            err_string := fmt.Sprintf("** err: %v\n", err)
            panic(err_string)
        }
        defer p.Close()

        p.CheckedCmd("unset key")
        p.CheckedCmd(fmt.Sprintf("set xrange [0:%.0f]", mapWidth))
        p.CheckedCmd(fmt.Sprintf("set yrange [0:%.0f]", mapHeight))

        p.SetStyle("linespoints lt 1 lw 1 ps 0.6 pt 1 rgb 'black'")
        for i, _ := range sources[name] {
            p.CheckedCmd(fmt.Sprintf("set arrow from %.0f,%.0f to %.0f,%.0f", sources[name][i].X, sources[name][i].Y, results[name][i].X, results[name][i].Y))
        }
        p.PlotXY([]float64{0,0}, []float64{0,0}, "")

        p.SetXLabel("X-coordinate")
        p.SetYLabel("Y-coordinate")
        p.CheckedCmd("set terminal pdf")
        p.CheckedCmd(fmt.Sprintf("set output '%v/%v-full.pdf'", directory, name))
        p.CheckedCmd("replot")
    }
}

func (e *engine) drawLastFrameCenter() {
    mapWidth := e.config.MapWidth
    mapHeight := e.config.MapHeight
    directory := e.config.OutputDir
    distance := 330.0 / 6
    sources := make(map[string][]*Location)
    results := make(map[string][]*Location)

    var location *Location
    var result *Location
    var signals Signals
    var success bool
    for x := mapWidth / 2 - 165.0 ; x <= mapWidth / 2 + 165.0; x += distance {
        for y := mapHeight / 2 - 165.0; y <= mapHeight / 2 + 165.0; y += distance {
            location = NewLocation(x,y)
            signals = e.m.Read(location)
            for name, algorithm := range e.algorithms {
                result, success = algorithm.read(signals, location)
                if success {
                    sources[name] = append(sources[name], location)
                    results[name] = append(results[name], result)
                }
            }
        }
    }

    for name, _ := range e.algorithms {
        // Initialize the plotters
        fname := ""
        persist := false
        debug := false

        p,err := gnuplot.NewPlotter(fname, persist, debug)
        if err != nil {
            err_string := fmt.Sprintf("** err: %v\n", err)
            panic(err_string)
        }
        defer p.Close()

        p.CheckedCmd("unset key")
        p.CheckedCmd(fmt.Sprintf("set xrange [%.0f:%.0f]", mapWidth / 2 - 250.0, mapWidth / 2 + 250.0))
        p.CheckedCmd(fmt.Sprintf("set yrange [%.0f:%.0f]", mapHeight / 2 - 250.0, mapHeight / 2 + 250.0))

        p.SetStyle("linespoints lt 1 lw 1 ps 0.6 pt 1 rgb 'black'")
        for i, _ := range sources[name] {
            p.CheckedCmd(fmt.Sprintf("set arrow from %.0f,%.0f to %.0f,%.0f", sources[name][i].X, sources[name][i].Y, results[name][i].X, results[name][i].Y))
        }
        p.PlotXY([]float64{0,0}, []float64{0,0}, "")

        p.SetXLabel("X-coordinate")
        p.SetYLabel("Y-coordinate")
        p.CheckedCmd("set terminal pdf")
        p.CheckedCmd(fmt.Sprintf("set output '%v/%v-center.pdf'", directory, name))
        p.CheckedCmd("replot")
    }
}