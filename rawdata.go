package mzlib

import (
  "strings"
  "errors"
  "fmt"
)

const (
// The mzlib version.
  Version string = "0.1"
)

// Represents the raw data from a mass spectrometry file.
type RawData struct {
  SourceFile string
  Instrument Instrument
  ScanCount uint64
  Scans []Scan
}

// Represents instrument metadata from the read in file.
type Instrument struct {
  Manufacturer string
  Model string
  MassAnalyzer string
  Detector string
  Resolution float64
  Accuracy float64
  Ionization string
}

// Represents a single scan in the mass spectrometry data.
type Scan struct {
  RetentionTime float64
  Polarity int8
  MsLevel uint8
  Id uint64
  MzRange [2]float64
  ParentScan uint64
  PrecursorMz float64
  PrecursorIntensity float64
  CollisionEnergy float64
  Continuous bool
  DeIsotoped bool
  MzArray []float64
  IntensityArray []float64
}

// Retreives the scan closest to the specified retention time
//
// Paramters:
//   retentionTime: The retention time value in minutes to locate the scan for.
func (r *RawData) GetScan(retentionTime float64) Scan {
  return *new(Scan)
}

// Removes any scans inside the specified range.
//
// Parameters:
//   minTime: The minimum retention time value in minutes to discard (inclusive)
//   maxTime: The maximum retention time value in minutes to discard (exclusive)
func (r *RawData) RemoveScans(minTime float64, maxTime float64) {
}

// Removes any scans outside the specified range.
//
// Parameters:
//   minTime: The minimum retention time value in minutes to retain (inclusive).
//   maxTime: The maximum retention time value in minutes to retain (exclusive).
func (r *RawData) OnlyScans(minTime float64, maxTime float64) {
}

// Removes any peaks inside the given range.
//
// Parameters:
//   mz: The m/z value to use for selecting peaks to keep.
//   tolerance: The tolerance value to use for selecting peaks. Peaks where
//     abs(mz - value) < tolerance will be discarded.
func (r *RawData) RemoveMz(mz float64, tolerance float64) {
}

// Removes any peaks outside the given range.
//
// Parameters:
//   mz: The m/z value to use for selecting peaks to keep.
//   tolerance: The tolerance value to use for selecting peaks. Peaks where
//     abs(mz - value) < tolerance will be kept.
func (r *RawData) OnlyMz(mz float64, tolerance float64) {
}

// Returns a selected ion chromatogram for the data.
//
// Parameters:
//   minMz: The minimum m/z value to select peaks from.
//   maxMz: The maximum m/z value to select peaks from.
//
// Return value:
// []float64: An array containing the total intensity of all peaks between
//   minMz and maxMz for each scan.
func (r *RawData) Sic(minMz float64, maxMz float64) []float64 {
  return nil
}

// Returns a total ion chromatogram for the data.
//
// Return value:
// []float64: An array containing the total intensity for each scan.
func (r *RawData) Tic() []float64 {
  return nil
}

// Returns a base peak chromatogram for the data.
//
// Return value:
// []float64: An array containing the intensity of the largest peak for each 
//   level 1 scan
func (r *RawData) Bpc() []float64 {
  return nil
}

// Finds the minimum m/z value in the data.
//
// Return value:
//   float64: The minimum m/z value.
func (r *RawData) MinMz () float64 {
  return float64(-1.0)
}

// Finds the maximum m/z value in the data.
//
// Return value:
//   float64: The maximum m/z value.
func (r *RawData) MaxMz () float64 {
  return float64(-1.0)
}

// Finds the maximum intensity value in the data.
func (r *RawData) PeakIntensity () float64 {
  return float64(-1.0)
}

// Reads mass spectrometry data from the specified file. The format is
// auto-detected based on the file name.
//
// Parameters:
//   filename: The name of the file to be written to
// 
// Return value:
//   error: Indicates whether or not an error occurred while reading the file
func (r *RawData) Read(filename string) error {
  if strings.ToLower(filename[len(filename)-6:]) == ".mzxml" {
    return r.ReadMzXml(filename)
  }
  if strings.ToLower(filename[len(filename)-7:]) == ".mzdata" {
    return r.ReadMzData(filename)
  }
  if strings.ToLower(filename[len(filename)-4:]) == ".xml" {
    return r.ReadMzData(filename)
  }
  if strings.ToLower(filename[len(filename)-5:]) == ".mzml" {
    return r.ReadMzMl(filename)
  }
  return errors.New(fmt.Sprintf("Filetype '%s' not recognized", 
                                filename[strings.LastIndex(filename,"."):]))
}

// Writes mass spectrometry data to the specified file. The format is auto-
//   detected base on the file name.
func (r *RawData) Write(filename string) error {
  if strings.ToLower(filename[len(filename)-6:]) == ".mzxml" {
    return r.WriteMzXml(filename)
  }
  if strings.ToLower(filename[len(filename)-7:]) == ".mzdata" {
    return r.WriteMzData(filename)
  }
  if strings.ToLower(filename[len(filename)-4:]) == ".xml" {
    return r.WriteMzData(filename)
  }
  if strings.ToLower(filename[len(filename)-5:]) == ".mzml" {
    return r.WriteMzMl(filename)
  }
  return errors.New(fmt.Sprintf("Filetype '%s' not recognized", 
                                filename[strings.LastIndex(filename,"."):]))
}

