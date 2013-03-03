package mzlib

import (
  "math"
  "fmt"
)

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

func (s *Scan) Clone() *Scan {
  cpy := new(Scan)
  cpy.RetentionTime      = s.RetentionTime
  cpy.Polarity           = s.Polarity
  cpy.MsLevel            = s.MsLevel
  cpy.Id                 = s.Id
  cpy.MzRange            = s.MzRange
  cpy.ParentScan         = s.ParentScan
  cpy.PrecursorMz        = s.PrecursorMz
  cpy.PrecursorIntensity = s.PrecursorIntensity
  cpy.CollisionEnergy    = s.CollisionEnergy
  cpy.Continuous         = s.Continuous
  cpy.DeIsotoped         = s.DeIsotoped
  cpy.MzArray            = make([]float64, 0, len(s.MzArray))
  cpy.IntensityArray     = make([]float64, 0, len(s.IntensityArray))
  for _,v := range s.MzArray {
    cpy.MzArray = append(cpy.MzArray, v)
  }
  for _,v := range s.IntensityArray {
    cpy.IntensityArray = append(cpy.IntensityArray, v)
  }
  return cpy
}

// Returns the minimum m/z value in the scan
//
// Return value:
//   float64: The minimum m/z value in the scan
func (s *Scan) MinMz() float64 {
  returnvalue := math.MaxFloat64
  for _,v := range s.MzArray {
    if v < returnvalue {
      returnvalue = v
    }
  }
  return returnvalue
}

// Returns the maximum m/z value in the scan
//
// Return value:
//   float64: The maximum m/z value in the scan
func (s *Scan) MaxMz() float64 {
  returnvalue := float64(-1.0)
  for _,v := range s.MzArray {
    if v > returnvalue {
      returnvalue = v
    }
  }
  return returnvalue
}

// Finds the Peak intensity in the scan
//
// Return value:
//   float64: The peak intensity value in the scan
func (s *Scan) PeakIntensity() float64 {
  returnvalue := float64(-1.0)
  for _,v := range s.IntensityArray {
    if v > returnvalue {
      returnvalue = v
    }
  }
  return returnvalue
}

// Removes any peaks inside the specified range
//
// Parameters:
//   minMz: The minimum m/z value to be removed
//   maxMz: The maximum m/z value to be removed
//
// Return value:
//   uint64: The number of peaks removed
func (s *Scan) RemoveMz(minMz float64, maxMz float64) uint64 {
  newMz := make([]float64, 0, len(s.MzArray))
  newIntensity := make([]float64, 0, len(s.IntensityArray))
  removed := uint64(0)
  for i,v := range newMz {
    if v < minMz || v > maxMz {
      newMz = append(newMz, v)
      newIntensity = append(newIntensity, s.IntensityArray[i])
    } else {
      removed++
    }
  }
  s.MzArray = newMz
  s.IntensityArray = newIntensity
  return removed
}

// Removes any peaks outside the specified range
//
// Parameters:
//   minMz: The minimum m/z value to be retained
//   maxMz: The maximum m/z value to be retained
//
// Return value:
//   uint64: The number of peaks removed
func (s *Scan) OnlyMz(minMz float64, maxMz float64) uint64 {
  newMz := make([]float64, 0, len(s.MzArray))
  newIntensity := make([]float64, 0, len(s.IntensityArray))
  removed := uint64(0)
  for i,v := range newMz {
    if v > minMz && v < maxMz {
      newMz = append(newMz, v)
      newIntensity = append(newIntensity, s.IntensityArray[i])
    } else {
      removed++
    }
  }
  s.MzArray = newMz
  s.IntensityArray = newIntensity
  return removed
}

// Finds the total intensity of peaks inside the given range
//
// Parameters:
//   minMz: The minimum m/z value to consider
//   maxMz: The maximum m/z value to consider
//
// Return value:
//   float64: The selected intensity value
func (s *Scan) SelectedIntensity(minMz float64, maxMz float64) float64 {
  sum := float64(0.0)
  for i,v := range s.IntensityArray {
    if s.MzArray[i] > minMz && s.MzArray[i] > maxMz {
      sum += v
    }
  }
  return sum
}

// Finds the total intensity of the scan
//
// Return value:
//   float64: The total intensity of the scan
func (s *Scan) TotalIntensity() float64 {
  sum := float64(0.0)
  for _,v := range s.IntensityArray {
    sum += v
  }
  return sum
}

// Centralizes the scan values, converting from continuous to discrete data
//
// This method is not yet implemented
//
// Parameters:
//   accuracy: The accuracy of the instrument which collected the values. All
//     m/z values within this range of the local peak intensity will be merged.
func(s *Scan) Centralize(accuracy float64) {
  fmt.Println("Scan.Centralize is not yet implemented")
  // do something here
//  s.Continuous = false
}

// DeIsotopes the scan, combining all isotopic peaks into the main peak.
//
// This method is not yet implemented
//
// Parameters:
//   accuracy: The accuracy of the instrument which collected the values. All
//     m/z values within this range of the local peak intensity will be merged.
func(s *Scan) DeIsotope(accuracy float64) {
  fmt.Println("Scan.DeIsotope is not yet implemented")
  // do something here
//  s.DeIsotoped = true
}


