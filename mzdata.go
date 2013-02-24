package mzlib

import (
  "encoding/xml"
  "strconv"
  "errors"
  "os"
  "io"
  "fmt"
  "strings"
)

type mzData struct {
  SourceFile string `xml:"description>admin>sourceFile>nameOfFile"`
  SourcePath string `xml:"description>admin>sourceFile>pathToFile"`
  InstrumentName string `xml:"description>instrument>instrumentName"`
  MassAnalyzer []cvParam `xml:"description>instrument>analyzerList>analyzer>cvParam"`
  ProcessingSoftware string `xml:"description>dataProcessing>software>name"`
  ProcessingSwVersion string `xml:"description>dataProcessing>software>version"`
  Detector []cvParam `xml:"description>instrument>detector>cvParam"`
  ProcessingMethod []cvParam `xml:"description>dataProcessing>processingMethod>cvParam"`
  SpectrumList struct {
    Scans []mzDataScan `xml:"spectrum"`
    ScanCount uint64 `xml:"count,attr"`
  } `xml:"spectrumList"`
}

type mzDataScan struct {
  Id uint64 `xml:"id,attr"`
  Instrument struct {
    MsLevel uint8 `xml:"msLevel,attr"`
    MzMin float64 `xml:"mzRangeStart,attr"`
    MzMax float64 `xml:"mzRangeStop,attr"`
    Params []cvParam `xml:"cvParam"`
  } `xml:"spectrumDesc>spectrumSettings>spectrumInstrument"`
  Specification struct {
    SpectrumType string `xml:"spectrumType,attr"`
    MethodOfCombination string `xml:"methodOfCombination,attr"`
  } `xml:"spectrumDesc>acqSpecification"`
  Precursor []struct {
    ParentScan uint64 `xml:"spectrumRef,attr"`
    IonSelection []cvParam `xml:"ionSelection>cvParam"`
    Activation []cvParam `xml:"activation>cvParam"`
  } `xml:"spectrumDesc>precursorList>precursor"`
  MzArray peakArray `xml:"mzArrayBinary>data"`
  IntensityArray peakArray `xml:"intenArrayBinary>data"`
}

type peakArray struct {
  Precision uint8 `xml:"precision,attr"`
  Endian string `xml:"endian,attr"`
  PeakCount uint64 `xml:"length,attr"`
  PeakList string `xml:",chardata"`
}

type cvParam struct {
  CvLabel string `xml:"cvLabel,attr"`
  Accession string `xml:"accession,attr"`
  Name string `xml:"name,attr"`
  Value string `xml:"value,attr"`
}

// Reads data from an MzData file
//
// Paramters:
//   filename: The name of the file to read from
//
// Return value:
//   error: Indicates whether or not an error occurred while reading the file
func (r *RawData) ReadMzData(filename string) error {
  file,err := os.Open(filename)
  if err != nil {
    return err
  }
  reader := io.Reader(file)
  return r.DecodeMzData(reader)
}

// Decodes data from a Reader containing MzData formatted data
//
// Parameters:
//   reader: The reader to read raw data from
//
// Return value:
//   error: Indicates whether or not an error occurred when reading the data
func (r *RawData) DecodeMzData(reader io.Reader) error {
  mz := mzData{}
  decoder := xml.NewDecoder(reader)
  // set up a dummy CharsetReader
  decoder.CharsetReader =
    func (charset string, input io.Reader) (io.Reader, error){
      return input, nil
    }
  e := decoder.Decode(&mz)
  if e != nil {
    return e
  }
  r.SourceFile = strings.Join([]string{mz.SourcePath, mz.SourceFile}, "/")
  r.Instrument.Model = mz.InstrumentName
  r.Instrument.Manufacturer = mz.InstrumentName
  r.Instrument.MassAnalyzer,_ = param(&mz.MassAnalyzer, "AnalyzerType")
  r.ScanCount = mz.SpectrumList.ScanCount
//  var scans []Scan
  // copy scan information
  for _,scan := range mz.SpectrumList.Scans {
    s := new(Scan)
    rt,_ := param(&scan.Instrument.Params, "TimeInMinutes")
    s.RetentionTime,_ = strconv.ParseFloat(rt, 64)
    if p,_ := param(&scan.Instrument.Params, "Polarity"); p == "positive" {
      s.Polarity = 1
    } else {
      s.Polarity = -1
    }
    s.MsLevel = scan.Instrument.MsLevel
    s.Id = scan.Id
    s.MzRange[0] = scan.Instrument.MzMin
    s.MzRange[1] = scan.Instrument.MzMax
    println(scan.MzArray.PeakCount)
    println(scan.IntensityArray.PeakCount)
    if len(scan.Precursor) > 0 {
      s.ParentScan = scan.Precursor[0].ParentScan
      mass,_ := param(&scan.Precursor[0].IonSelection, "MassToChargeRatio")
      s.PrecursorMz,_ = strconv.ParseFloat(mass, 64)
      ce,_ := param(&scan.Precursor[0].Activation, "CollisionEnergy")
      s.CollisionEnergy,_ = strconv.ParseFloat(ce, 64)
    }
    s.Continuous = scan.Specification.SpectrumType == "continuous"
    iso,_ := param(&mz.ProcessingMethod, "Deisotoping")
    s.DeIsotoped,_ = strconv.ParseBool(iso)
    _ = Float64FromBase64(&s.MzArray, scan.MzArray.PeakList,
                          scan.MzArray.Precision,
                          scan.MzArray.Endian == "big")
    _ = Float64FromBase64(&s.IntensityArray, scan.IntensityArray.PeakList,
                          scan.IntensityArray.Precision,
                          scan.IntensityArray.Endian == "big")
    r.Scans = append(r.Scans, *s)
  }
  return nil
}

// Writes the data to disk in MzData format
//
// Parameters:
//   filename: The name of the file to be written to
//
// Return value:
//   error: Indicates whether or not an error occurred while writing the file
func (r *RawData) WriteMzData( filename string ) error {
  return errors.New("Writing this file type has not yet been implemented")
}

// Reads through a slice of cvParams to find the appropriate value
//
// Parameters:
//   params: A Pointer to the slice of cvParams to search
//   name: The name of the attribute to search for
//
// Return values:
//   string: The attribute's value, or an empty string if not found
//   error: An error if the attribute was not found
func param( params *[]cvParam, name string ) (string, error) {
  for _,v := range *params {
    if v.Name == name {
      return v.Value, nil
    }
  }
  return "", errors.New(fmt.Sprintf("Key '%s' Not Found", name ))
}

