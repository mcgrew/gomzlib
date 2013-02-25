package mzlib

import (
  "strings"
  "errors"
  "fmt"
  "math"
)

const (
// The mzlib version.
  Version string = "0.1.2012.02.24"
)

// Represents the raw data from a mass spectrometry file.
type RawData struct {
  Filename string
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
func (r *RawData) GetScan(retentionTime float64) *Scan {
  var returnvalue *Scan
  if len(r.Scans) > 0 {
    returnvalue = &r.Scans[0]
  }
  diff := math.Abs((*returnvalue).RetentionTime - retentionTime)
  for i,l := 1, len(r.Scans); i < l; i++ {
    if math.Abs(r.Scans[i].RetentionTime - retentionTime) < diff {
      returnvalue = &r.Scans[i]
      diff = math.Abs((*returnvalue).RetentionTime - retentionTime)
    }
  }
  return returnvalue
}

// Removes any scans inside the specified range.
//
// Parameters:
//   minTime: The minimum retention time value in minutes to discard (inclusive)
//   maxTime: The maximum retention time value in minutes to discard (exclusive)
//
// Return value:
//   uint64: The number of scans which were removed
func (r *RawData) RemoveScans(minTime float64, maxTime float64) uint64 {
  var newScans []Scan
  removed := uint64(0)
  r.ScanCount = 0
  for _,v := range r.Scans {
    if v.RetentionTime < minTime || v.RetentionTime > maxTime {
      newScans = append(newScans, v)
      r.ScanCount++
    } else {
      removed++
    }
  }
  r.Scans = newScans
  return removed
}

// Removes any scans outside the specified range.
//
// Parameters:
//   minTime: The minimum retention time value in minutes to retain (inclusive).
//   maxTime: The maximum retention time value in minutes to retain (exclusive).
func (r *RawData) OnlyScans(minTime float64, maxTime float64) uint64{
  var newScans []Scan
  removed := uint64(0)
  r.ScanCount = 0
  for _,v := range r.Scans {
    if v.RetentionTime > minTime && v.RetentionTime < maxTime {
      newScans = append(newScans, v)
      r.ScanCount++
    } else {
      removed++
    }
  }
  r.Scans = newScans
  return removed
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
  returnvalue := make([]float64, 0, r.ScanCount)
  var sum float64
  for _,s := range r.Scans {
    if s.MsLevel == 1 {
      sum = 0.0
      for i,v := range s.IntensityArray {
        if s.MzArray[i] > minMz && s.MzArray[i] > maxMz {
          sum += v 
        }
      }
      returnvalue = append(returnvalue, sum)
    }
  }
  return returnvalue
}

// Returns a total ion chromatogram for the data.
//
// Return value:
// []float64: An array containing the total intensity for each scan.
func (r *RawData) Tic() []float64 {
  returnvalue := make([]float64, 0, r.ScanCount)
  var sum float64
  for _,s := range r.Scans {
    if s.MsLevel == 1 {
      sum = 0.0
      for _,v := range s.IntensityArray {
        sum += v 
      }
      returnvalue = append(returnvalue, sum)
    }
  }
  return returnvalue
}

// Returns a base peak chromatogram for the data.
//
// Return value:
// []float64: An array containing the intensity of the largest peak for each 
//   level 1 scan
func (r *RawData) Bpc() []float64 {
  returnvalue := make([]float64, 0, r.ScanCount)
  var val float64
  for _,s := range r.Scans {
    if s.MsLevel == 1 {
      val = 0.0
      for _,v := range s.IntensityArray {
        if v > val {
          val = v
        }
      }
      returnvalue = append(returnvalue, val)
    }
  }
  return returnvalue
}

// Finds the minimum m/z value in the data.
//
// Return value:
//   float64: The minimum m/z value.
func (r *RawData) MinMz () float64 {
  returnvalue := math.MaxFloat64
  for _,s := range r.Scans {
    for _,v := range s.MzArray {
      if v < returnvalue {
        returnvalue = v
      }
    }
  }
  return returnvalue
}

// Finds the maximum m/z value in the data.
//
// Return value:
//   float64: The maximum m/z value.
func (r *RawData) MaxMz () float64 {
  returnvalue := float64(-1.0)
  for _,s := range r.Scans {
    for _,v := range s.MzArray {
      if v > returnvalue {
        returnvalue = v
      }
    }
  }
  return returnvalue
}

// Finds the maximum intensity value in the data.
func (r *RawData) PeakIntensity () float64 {
  returnvalue := float64(-1.0)
  for _,s := range r.Scans {
    for _,v := range s.IntensityArray {
      if v > returnvalue {
        returnvalue = v
      }
    }
  }
  return returnvalue
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

func (s *Scan) MinMz() float64 {
  returnvalue := math.MaxFloat64
  for _,v := range s.MzArray {
    if v < returnvalue {
      returnvalue = v
    }
  }
  return returnvalue
}

func (s *Scan) MaxMz() float64 {
  returnvalue := float64(-1.0)
  for _,v := range s.MzArray {
    if v > returnvalue {
      returnvalue = v
    }
  }
  return returnvalue
}

func (s *Scan) PeakIntensity() float64 {
  returnvalue := float64(-1.0)
  for _,v := range s.IntensityArray {
    if v > returnvalue {
      returnvalue = v
    }
  }
  return returnvalue
}

func (s *Scan) RemoveMz(minMz float64, maxMz float64) {
}

func (s *Scan) OnlyMz(minMz float64, maxMz float64) {
}

