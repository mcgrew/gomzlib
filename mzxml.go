package mzlib

import (
  "encoding/xml"
  "os"
  "io"
  "strconv"
  "errors"
  "encoding/binary"
)

type mzxml struct {
  Run struct {
    ScanCount uint64 `xml:"scanCount,attr"`
    SourceFile struct {
      Name string `xml:"fileName,attr"`
    } `xml:"parentFile"`
    Instrument msinstrument `xml:"msInstrument"`
    Processing struct {
      Centroided int8 `xml:"centroided,attr"`
    } `xml:"dataProcessing"`
    Scans []mzxmlscan `xml:"scan"`
  } `xml:"msRun"`
}

type mzxmlscan struct {
  Peaks peaks `xml:"peaks"`
  MsLevel uint8 `xml:"msLevel,attr"`
  Id uint64 `xml:"num,attr"`
  Scans []mzxmlscan `xml:"scan"`
  LowMz float64 `xml:"lowMz,attr"`
  HighMz float64 `xml:"highMz,attr"`
  BasePeakMz float64 `xml:"basePeakMz,attr"`
  BasePeakIntensity float64 `xml:"basePeakIntensity,attr"`
  Tic float64 `xml:"totIonCurrent,attr"`
  PeakCount uint64 `xml:"peaksCount,attr"`
  Polarity string `xml:"polarity,attr"`
  RetentionTime string `xml:"retentionTime,attr"`
  CollisionEnergy float64 `xml:"collisionEnergy,attr"`
  Precursor struct {
    Intensity float64 `xml:"precursorIntensity,attr"`
    Mz float64 `xml:",chardata"`
  } `xml:"precursorMz"`
}

type peaks struct {
  PeakList string `xml:",chardata"`
  Precision uint8 `xml:"precision,attr"`
  ByteOrder string `xml:"byteOrder,attr"`
  PairOrder string `xml:"pairOrder,attr"`
  CompressionType string `xml:"compressionType,attr"`
}

type msinstrument struct {
  Manufacturer struct {
    Name string `xml:"value,attr"`
  } `xml:"msManufacturer"` 
  Model struct {
    Name string `xml:value,attr"`
  } `xml:"msModel"`
  MassAnalyzer struct {
    Name string `xml:"value,attr"`
  } `xml:"msMassAnalyzer"`
}

// Reads data from an MzXML file
//
// Paramters:
//   filename: The name of the file to read from
//
// Return value:
//   error: Indicates whether or not an error occurred while reading the file
func (r *RawData) ReadMzXml(filename string) error {
  file,err := os.Open(filename)
  if err != nil {
    return err
  }
  defer file.Close()
  reader := io.Reader(file)
  return r.DecodeMzXml( reader )
}

// Decodes data from a Reader containing MzXML formatted data
//
// Parameters:
//   reader: The reader to read raw data from
//
// Return value:
//   error: Indicates whether or not an error occurred when reading the data
func (r *RawData) DecodeMzXml(reader io.Reader) error {
  mz := mzxml{}
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
  r.SourceFile = mz.Run.SourceFile.Name
  r.Instrument.Model = mz.Run.Instrument.Model.Name
  r.Instrument.Manufacturer = mz.Run.Instrument.Manufacturer.Name
  r.Instrument.MassAnalyzer = mz.Run.Instrument.MassAnalyzer.Name
  r.ScanCount = mz.Run.ScanCount
  var scans []Scan
  // copy scan information
  for i:=0; i < len(mz.Run.Scans); i++ {
    s := new(Scan)
    s.Continuous = mz.Run.Processing.Centroided == 0
    scanInfo(s, &(mz.Run.Scans[i]), 0)
    scans = append(scans, *s)
    if len(mz.Run.Scans[i].Scans) > 0 {
      for j:=0; j < len(mz.Run.Scans[i].Scans); j++ {
        s := new(Scan)
        s.Continuous = mz.Run.Processing.Centroided == 0
        scanInfo(s, &(mz.Run.Scans[i].Scans[j]), mz.Run.Scans[i].Id)
        scans = append(scans, *s)
      }
    }
    r.Scans = scans
  }
  return nil
}

// Writes the data to disk in MzXML format
//
// Parameters:
//   filename: The name of the file to be written to
//
// Return value:
//   error: Indicates whether or not an error occurred while writing the file
func (r *RawData) WriteMzXml(filename string) error {
  return errors.New("Writing this file type has not yet been implemented")
}

// Decodes scan information read from a file
//
// Parameters:
//   s: A pointer to the Scan struct to save the decoded data to
//   mzs: A pointer to the mzxmlscan object to read the data from
//   parentScan: The Id of the parent scan, or 0 if none
func scanInfo( s *Scan, mzs *mzxmlscan, parentScan uint64) {
  rt := mzs.RetentionTime
  s.RetentionTime,_ = strconv.ParseFloat(rt[2:len(rt)-1], 64)
  s.RetentionTime /= 60
  if mzs.Polarity == "-" {
    s.Polarity = -1
  } else if mzs.Polarity == "+" {
    s.Polarity = 1
  } else {
    s.Polarity = 0 // unknown
  }
  s.MsLevel = mzs.MsLevel
  s.Id = mzs.Id
  s.MzRange[0] = mzs.LowMz
  s.MzRange[1] = mzs.HighMz
  s.ParentScan = parentScan
  s.PrecursorMz = mzs.Precursor.Mz
  s.PrecursorIntensity = mzs.Precursor.Intensity
  s.CollisionEnergy = mzs.CollisionEnergy

  // now decode the peak data
  s.MzArray = make([]float64, 0, mzs.PeakCount)
  s.IntensityArray = make([]float64, 0, mzs.PeakCount)
  values := make([]float64, 0, mzs.PeakCount*2)
  // mzxml is always bigEndian per the spec
  _ = Float64FromBase64(&values, mzs.Peaks.PeakList, mzs.PeakCount*2,
                        mzs.Peaks.Precision, 
                        mzs.Peaks.CompressionType == "zlib", binary.BigEndian)
  n := len(values)
  for i := 0 ; i < n; i+=2 {
    s.MzArray = append(s.MzArray, values[i])
    s.IntensityArray = append(s.IntensityArray, values[i+1])
  }
}

