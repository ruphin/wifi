package wifi

import (
    "fmt"
    "math"
    "math/rand"
    "sort"
    "time"
    "os"
    "image"
    "image/color"
    "image/draw"
    "image/png"
    "strconv"
    "github.com/ruphin/go-gnuplot/pkg/gnuplot"
)

func init() {
    rand.Seed(time.Now().UTC().UnixNano())
    random = rand.New(rand.NewSource(rand.Int63()))
    log10 = math.Log(10)
}

var random *rand.Rand
var log10 float64

//////////////
// Location //
//////////////

type Location struct {
    X, Y float64
}

func NewLocation(x, y float64) *Location {
    return &Location{x, y}
}

func NewRandomLocation(xmax, ymax float64) *Location {
    return NewLocation(random.Float64()*xmax, random.Float64()*ymax)
}

func distance(l1, l2 *Location) float64 {
    x := l1.X - l2.X
    y := l1.Y - l2.Y
    return math.Sqrt(x*x + y*y)
}

func (l *Location) String() string {
    return "(" + fmt.Sprintf("%.2f", l.X) + ", " + fmt.Sprintf("%.2f", l.Y) + ")"
}

func (l *Location) enhance(realLocation *Location) {
    l.X = l.X + ((realLocation.X - l.X) * 0.1)
    l.Y = l.Y + ((realLocation.Y - l.Y) * 0.1)
}

//////////////////
// Access Point //
//////////////////

type AccessPoint struct {
    id int
    location *Location
}

var accessPointCount int

func NewAccessPoint(location *Location) *AccessPoint {
    accessPointCount += 1
    return &AccessPoint{accessPointCount, location}
}

func (ap *AccessPoint) String() string {
    return fmt.Sprint(ap.location)
}

////////////
// Signal //
////////////

type Signal struct {
    id int
    signalStrength float64
}


/////////////
// Signals //
/////////////

type Signals []Signal

// Sorting helpers
type BySignalStrength Signals
func (signals BySignalStrength) Len() int           { return len(signals) }
func (signals BySignalStrength) Swap(i, j int)      { signals[i], signals[j] = signals[j], signals[i] }
func (signals BySignalStrength) Less(i, j int) bool { return signals[i].signalStrength > signals[j].signalStrength }

type ByID Signals
func (signals ByID) Len() int           { return len(signals) }
func (signals ByID) Swap(i, j int)      { signals[i], signals[j] = signals[j], signals[i] }
func (signals ByID) Less(i, j int) bool { return signals[i].id < signals[j].id }

// Key() returns a Key object that reprsents the IDs of the signals
func (signals Signals) Key() Key {
    if len(signals) > keyLength {
        err_string := "** err: Signal length exceeds maximum keysize.\n"
        panic(err_string)
    }
    var key Key
    ids := make([]int, len(signals))
    for i, signal := range signals {
        ids[i] = signal.id
    }
    sort.Ints(ids)
    copy(key[:], ids[:])
    return key
}

/////////
// Key //
/////////

const keyLength int = 50
// A Key is a fixed length list of integers representing the IDs in a Signals struct
type Key [keyLength]int

/////////
// Map //
/////////

type Map struct {
    width, height float64
    accessPoints []AccessPoint
}

func NewMap(width, height float64, source int64) *Map {
    if source != 0 {
        random = rand.New(rand.NewSource(source))
    }
    return &Map{width, height, []AccessPoint{}}
}

func (m *Map) AddAccessPoint(location *Location) {
    m.accessPoints = append(m.accessPoints, *NewAccessPoint(location))
}

func (m *Map) AddRandomAccessPoint() {
    m.AddAccessPoint(NewRandomLocation(m.width, m.height))
}

func (m *Map) RemoveAccessPoint(id int) {
    for i, v := range m.accessPoints {
        if v.id == id {
            m.accessPoints = append(m.accessPoints[:i], m.accessPoints[i+1:]...)
            break
        }
    }
}

func (m *Map) RemoveOldestAccessPoint() {
    m.accessPoints = m.accessPoints[1:]
}

func (m *Map) RemoveRandomAccessPoint() int {
    i := random.Intn(len(m.accessPoints))
    id := m.accessPoints[i].id
    m.accessPoints = append(m.accessPoints[:i], m.accessPoints[i+1:]...)
    return id
}

func (m *Map) String() string {
    s := "Map [" + fmt.Sprint(m.width) + " x " + fmt.Sprint(m.height) + "]\n{"
    for _, ap := range m.accessPoints {
        s += ap.String() + ", "
    }
    return s + "}"
}

// Returns a slice of Signals that are read at the given location.
func (m *Map) Read(location *Location) Signals {
    var dist float64
    signals := make(Signals, 0, len(m.accessPoints))
    for _, ap := range m.accessPoints {
        dist = distance(ap.location, location)
        if signalReceived(dist) {
            signals = signals[0:len(signals)+1]
            signals[len(signals)-1] = Signal{ap.id, signalStrength(dist)}
        }
    }
    trimmedSignals := make(Signals, len(signals))
    copy(trimmedSignals, signals)
    return trimmedSignals
}


func (m *Map) Draw(accessPointCutoffs []int) {
    width := int(m.width) + 5
    height := int(m.height) + 5
    mapImage := image.NewRGBA(image.Rect(0,0,width,height))
    background := color.RGBA{255,255,255,255}
    draw.Draw(mapImage, mapImage.Bounds(), &image.Uniform{background}, image.ZP, draw.Src)
    
    ruler := color.RGBA{210,210,210,255}
    for x := 102; x < width-3; x += 100 {
        draw.Draw(mapImage, image.Rect(x,0,x+1,height), &image.Uniform{ruler}, image.ZP, draw.Src)
    }
    for y := 102; y < height-3; y += 100 {
        draw.Draw(mapImage, image.Rect(0,y,width,y+1), &image.Uniform{ruler}, image.ZP, draw.Src)
    }

    border := color.RGBA{0,0,0,255}
    draw.Draw(mapImage, image.Rect(0,0,width,1), &image.Uniform{border}, image.ZP, draw.Src)
    draw.Draw(mapImage, image.Rect(0,0,1,height), &image.Uniform{border}, image.ZP, draw.Src)
    draw.Draw(mapImage, image.Rect(width,height,0,height-1), &image.Uniform{border}, image.ZP, draw.Src)
    draw.Draw(mapImage, image.Rect(width,height,width-1,0), &image.Uniform{border}, image.ZP, draw.Src)

    colors := make([]color.RGBA, len(accessPointCutoffs))
    colors[0] = color.RGBA{0,255,0,255}
    var shade uint8
    for i := 0; i < len(colors) - 1; i++ {
        shade = uint8(float64(i) / float64(len(colors) - 2) * 255)
        colors[len(colors) - i - 1] = color.RGBA{255-shade, 0, shade, 255}
    }

    var x,y int
    for _, accessPoint := range m.accessPoints {
        x = int(accessPoint.location.X)
        y = int(accessPoint.location.Y)
        for i, cutoff := range accessPointCutoffs {
            if accessPoint.id <= cutoff {
                draw.Draw(mapImage, image.Rect(x,y,x+5,y+5), &image.Uniform{colors[i]}, image.ZP, draw.Src)
                break
            }
        }
    }
    if os.MkdirAll("maps", 0777) != nil {
        panic("** err: Unable to create directory for maps")
    }
    image, _ := os.Create("maps/map" + strconv.Itoa(len(accessPointCutoffs)) + ".png")

    png.Encode(image, mapImage)
}

//////////////////////
// HELPER FUNCTIONS //
//////////////////////

func signalReceived(distance float64) bool {
    return random.Float64() < 0.6 - math.Log(distance/64 + 0.5)
}

func signalStrength(distance float64) float64 {
    rss := -58 - (14 * math.Log(distance + 5) / log10)

    stddev := 0.0497 * rss + 6.3438
    theta := 2 * math.Pi * random.Float64()
    rho := math.Sqrt(-2 * math.Log(1 - random.Float64()))
    scale := stddev * rho
    return rss + scale * math.Sin(theta)
}

func Test() {
    distances := []float64{5,10,15,20,25,30,35,40,45,50,55,60,65,70,75,80,85,90,95,100}
    results := make([]map[int]float64, len(distances))
    for i, _ := range results {
        results[i] = make(map[int]float64)
    }
    testSize := 1000000
    for i := 0; i < testSize; i++ {
        for i, distance := range distances {
            if signalReceived(distance) {
                results[i][int(signalStrength(distance))]++
            }
        }
    }

    resultLists := make([][]float64, len(distances))
    for i, _ := range resultLists {
        resultLists[i] = make([]float64, 50)
    }

    var resultList []float64
    var totalHits float64
    responseRates := make([]float64, len(distances))
    for i, _ := range distances {
        resultList = resultLists[i]
        totalHits = 0
        for j := -100; j < -50; j++ {
            resultList[j+100] = results[i][j]
            totalHits += results[i][j]
        }
        for j, result := range resultList {
            resultList[j] = result / totalHits * 100
        }
        responseRates[i] = totalHits / float64(testSize) * 100
    }

    graphReponseRate(distances, responseRates)
    graphSignalStrength()

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

    p.CheckedCmd("set xrange [-100:-50]")
    p.CheckedCmd("set datafile missing '0'")
    p.CheckedCmd("set key right top")
    p.CheckedCmd("set yrange [0:30]")

    bySignalStrength := make([]float64, 50)
    for i := 0; i < 50; i++ {
        bySignalStrength[i] = float64(i - 100)
    }

    p.SetStyle("linespoints lt 1 lw 1 ps 0.6 pt 5")
    p.PlotXY(bySignalStrength, resultLists[14], fmt.Sprintf("Response Rate=%.0f%v", responseRates[14], "%%"))

    p.SetStyle("linespoints lt 1 lw 1 ps 0.6 pt 9")
    p.PlotXY(bySignalStrength, resultLists[8], fmt.Sprintf("Response Rate=%.0f%v", responseRates[8], "%%"))

    p.SetStyle("linespoints lt 1 lw 1 ps 0.6 pt 13")
    p.PlotXY(bySignalStrength, resultLists[0], fmt.Sprintf("Response Rate=%.0f%v", responseRates[0], "%%"))

    p.SetXLabel("Signal Strength (dBm)")
    p.SetYLabel("Percentage of readings")
    p.CheckedCmd("set terminal pdf monochrome lw 2")
    p.CheckedCmd("set output 'graphs/readings.pdf'")
    p.CheckedCmd("replot")
}

func graphSignalStrength() {
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

    p.CheckedCmd("set xrange [0:140]")
    p.CheckedCmd("set datafile missing '0'")
    p.CheckedCmd("set key right top")
    p.CheckedCmd("set yrange [-100:-50]")

    signalStrengths := make([]float64, 125)
    byDistance := make([]float64, 125)
    for i := 5; i < 125; i++ {
        signalStrengths[i] = -58 - (14 * math.Log(float64(i) + 5) / log10)
        byDistance[i] = float64(i)
    }
    fmt.Println(byDistance)
    fmt.Println(signalStrengths)
    p.SetStyle("lines lt 1 lw 2 ps 0.6 pt 5")
    p.PlotXY(byDistance, signalStrengths, "Median signal strength")

    p.SetXLabel("Distance (meters)")
    p.SetYLabel("Signal Strength (dBm)")
    p.CheckedCmd("set terminal pdf monochrome lw 2")
    p.CheckedCmd("set output 'graphs/signalStrengths.pdf'")
    p.CheckedCmd("replot")

}

func graphReponseRate(distances []float64, responseRates []float64) {
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

    p.CheckedCmd("set xrange [0:100]")
    p.CheckedCmd("set key right top")
    p.CheckedCmd("set yrange [0:100]")

    p.SetStyle("linespoints lt 1 lw 1 ps 0.6 pt 5")
    p.PlotXY(distances, responseRates, "Response Rate")

    p.SetXLabel("Distance from AP (meters)")
    p.SetYLabel("Response Rate (%%)")
    p.CheckedCmd("set terminal pdf monochrome lw 2")
    p.CheckedCmd("set output 'graphs/responseRates.pdf'")
    p.CheckedCmd("replot")
}
