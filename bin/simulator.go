package main

import (
    "github.com/ruphin/wifi"
)

func main() {
    configuration := wifi.NewConfiguration()
    configuration.OutputDir = "graphs/durdle"
    configuration.MapHeight = 1000
    configuration.MapWidth = 1000
    configuration.AccessPointDensity = 1500
    configuration.SeedDistance = 5
    configuration.TestDistance = 10
    configuration.TestCycles = 50
    configuration.ReplacementRate = 0.1
    configuration.ReplacementStrategy = wifi.FiFoReplacement

    engine := wifi.NewEngine(configuration)
    engine.AddAlgorithm("Centroid", wifi.NewCentroid())
    engine.AddAlgorithm("Learning Centroid", wifi.NewLearningCentroid())
    engine.AddAlgorithm("Smart Learning Centroid", wifi.NewSmartLearningCentroid())
    engine.AddAlgorithm("Enhanced Centroid", wifi.NewEnhancedCentroid())
    engine.AddAlgorithm("Enhanced Learning Centroid", wifi.NewEnhancedLearningCentroid())
    // engine.AddAlgorithm("Fingerprinting", wifi.NewFingerprinting())
    // engine.AddAlgorithm("Learning Fingerprinting", wifi.NewLearningFingerprinting())
    // engine.AddAlgorithm("Fingerprinting", wifi.NewSmartLearningFingerprinting())
    // engine.AddAlgorithm("Enhanced Fingerprinting", wifi.NewEnhancedFingerprinting())
    // engine.AddAlgorithm("Enhanced Learning Fingerprinting", wifi.NewEnhancedLearningFingerprinting())
    engine.Run()

    // wifi.Test()
}