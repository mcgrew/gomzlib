package mzlib

import (
)

type RawData struct {
  SourceFile string
  Instrument Instrument
  ScanCount uint64
  Scans []Scan
}

type Instrument struct {
  Manufacturer string
  Model string
  MassAnalyzer string
  Detector string
}

type Scan struct {
  RetentionTime float64
  Polarity int8
  MsLevel uint8
  Id uint64
  MzRange [2]float64
  ParentScan *Scan
  PrecursorMz float64
  CollisionEnergy float64
  MzArray []float64
  IntensityArray []float64
}

