package wifi

type centroidAccessPoint struct {
    x, y []float64
    location *Location
}

type centroid struct {
    accessPointMap map[int]centroidAccessPoint
    minMatches int
    enhanced bool
    learning bool
    smart bool
}

func NewCentroid() algorithm {
    return algorithm(&centroid{make(map[int]centroidAccessPoint), 4, false, false, false})
}

func NewEnhancedCentroid() algorithm {
    return algorithm(&centroid{make(map[int]centroidAccessPoint), 4, true, false, false})
}

func NewLearningCentroid() algorithm {
    return algorithm(&centroid{make(map[int]centroidAccessPoint), 4, false, true, false})
}

func NewEnhancedLearningCentroid() algorithm {
    return algorithm(&centroid{make(map[int]centroidAccessPoint), 4, true, true, false})
}

func NewSmartLearningCentroid() algorithm {
    return algorithm(&centroid{make(map[int]centroidAccessPoint), 4, false, true, true})
}

func (c *centroid) feed(signals Signals, location *Location) {
    var accessPoint centroidAccessPoint
    var x, y float64
    for _, signal := range signals {
        accessPoint = c.accessPointMap[signal.id]
        accessPoint.x = append(accessPoint.x, location.X)
        accessPoint.y = append(accessPoint.y, location.Y)
        if len(accessPoint.x) >= c.minMatches && (!c.smart || len(accessPoint.x) < 300) {
            x = 0
            for _, value := range accessPoint.x {
                x += value
            }
            x = x / float64(len(accessPoint.x))

            y = 0
            for _, value := range accessPoint.y {
                y += value
            }
            y = y / float64(len(accessPoint.y))
            accessPoint.location = NewLocation(x, y)
        }
        c.accessPointMap[signal.id] = accessPoint
    }
}

func (c *centroid) read(signals Signals, realLocation *Location) (*Location, bool) {
    var accessPoint centroidAccessPoint
    var xList, yList []float64
    var exists bool
    for _, signal := range signals {
        accessPoint, exists = c.accessPointMap[signal.id]
        if exists && accessPoint.location != nil {
            xList = append(xList, accessPoint.location.X)
            yList = append(yList, accessPoint.location.Y)
        }
    }
    if len(xList) == 0 {
        return nil, false
    } else {
        var x, y float64 = 0, 0
        for _, value := range xList {
            x += value
        }
        x = x / float64(len(xList))

        y = 0
        for _, value := range yList {
            y += value
        }
        y = y / float64(len(yList))

        location := NewLocation(x, y)

        if c.enhanced {
            location.enhance(realLocation)
        }

        if c.learning {
            c.feed(signals, location)
        }
        return location, true
    }
}