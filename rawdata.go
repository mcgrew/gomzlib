package mzlib

import (
  "strings"
  "errors"
  "fmt"
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

func (r *RawData) GetScan(retentionTime float64) Scan {
  return *new(Scan)
}

func (r *RawData) RemoveScans(minTime float64, maxTime float64) {
}

func (r *RawData) OnlyScans(minTime float64, maxTime float64) {
}

func (r *RawData) RemoveMz(mz float64, tolerance float64) {
}

func (r *RawData) OnlyMz(mz float64, tolerance float64) {
}

func (r *RawData) Sic() []float64 {
  return nil
}

func (r *RawData) Tic() []float64 {
  return nil
}

func (r *RawData) Bpc() []float64 {
  return nil
}

func (r *RawData) MinMz () float64 {
  return float64(-1.0)
}

func (r *RawData) MaxMz () float64 {
  return float64(-1.0)
}

func (r *RawData) Read(filename string) error {
  if strings.ToLower(filename[len(filename)-6:]) == ".mzxml" {
    return r.ReadMzXml(filename)
  }
  if strings.ToLower(filename[len(filename)-7:]) == ".mzdata" {
    return r.ReadMzData(filename)
  }
  if strings.ToLower(filename[len(filename)-11:]) == ".mzdata.xml" {
    return r.ReadMzData(filename)
  }
  if strings.ToLower(filename[len(filename)-5:]) == ".mzml" {
    return r.ReadMzMl(filename)
  }
  return errors.New(fmt.Sprintf("Filetype '%s' not recognized", 
                                filename[strings.LastIndex(filename,"."):]))
}

func (r *RawData) write(filename string) error {
  return nil
}

