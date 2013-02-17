package mzlib

import (
  "encoding/xml"
  "os"
  "io"
)

type mzxmlscan struct {
  Peaks peaks `xml:"peaks"`
  Scans []mzxmlscan `xml:"scan"`
  LowMz float64 `xml:"lowMz,attr"`
  HighMz float64 `xml:"highMz,attr"`
  BasePeakMz float64 `xml:"basePeakMz"`
  BasePeakIntensity float64 `xml:"basePeakIntensity"`
  TIC float64 `xml:"totIonCurrent"`
  Polarity string `xml:"polarity"`
  PeakCount string `xml:"peaksCount"`
  RetentionTime string `xml:"retentionTime"`
}

type peaks struct {
  Data string `xml:""`
  Precision int8 `xml:"precision,attr"`
  ByteOrder string `xml"byteOrder,attr"`
  PairOrder string `xml"pairOrder,attr"`
}

type msinstrument struct {
  Manufacturer msmanufacturer `xml:"msManufacturer"` 
  Model msmodel `xml:"msModel"`
  MassAnalyzer massanalyzer `xml:"msMassAnalyzer"`
}

type msmanufacturer struct {
  Name string `xml:"value,attr"`
}

type msmodel struct {
  Name string `xml:value,attr"`
}

type massanalyzer struct {
  Name string `xml:"value,attr"`
}

type msrun struct {
  ScanCount uint64 `xml:"scanCount,attr"`
  SourceFile string `xml:"parentFile>fileName"`
  Instrument msinstrument `xml:"msRun>msInstrument"`
}

type sourceFile struct {
  Name string `xml:"fileName,attr"`
}

type mzxml struct {
//  XMLName xml.Name `xml:"mzXML"`
  Run msrun `xml:"msRun"`
  Scans []mzxmlscan `xml:"scan"`
}

func (r *RawData) ReadMzXml(filename string) error {
  mz := mzxml{}
//  xmldata [0]byte
  file,err := os.Open(filename)
  if err != nil {
    return err
  }
  reader := io.Reader(file)
  decoder := xml.NewDecoder(reader)
  decoder.CharsetReader = dummyCharsetReader
  err = decoder.Decode(&mz)
  if err != nil {
    return err
  }
  r.SourceFile = mz.Run.SourceFile
  r.Instrument.Model = mz.Run.Instrument.Model.Name
  r.Instrument.Manufacturer = mz.Run.Instrument.Manufacturer.Name
  r.Instrument.MassAnalyzer = mz.Run.Instrument.MassAnalyzer.Name
  r.ScanCount = mz.Run.ScanCount
  var scans []Scan

  // copy scan information
  for i:=0; i < len(mz.Scans); i++ {
    s := new(Scan)
    scanInfo(s, &(mz.Scans[i]))
    scans = append(scans, *s)
    if len(mz.Scans[i].Scans) > 0 {
      for j:=0; j < len(mz.Scans[i].Scans); j++ {
        s := new(Scan)
        scanInfo(s, &(mz.Scans[i]))
        scans = append(scans, *s)
      }
    }
  }
  return nil
}

func (r *RawData) WriteMzXml(filename string) error {
  return nil
}

func scanInfo( s *Scan, mzs *mzxmlscan ) {
}

func dummyCharsetReader(charset string, input io.Reader) (io.Reader, error){
  return input, nil
}

