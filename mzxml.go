package mzlib

import (
  "encoding/xml"
  "encoding/base64"
  "os"
  "io"
  "strconv"
  "math"
)

type mzxml struct {
  Run msrun `xml:"msRun"`
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
  PeakCount int64 `xml:"peaksCount,attr"`
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

type msrun struct {
  ScanCount uint64 `xml:"scanCount,attr"`
  SourceFile struct {
    Name string `xml:"fileName,attr"`
  } `xml:"parentFile"`
  Instrument msinstrument `xml:"msInstrument"`
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
  // set up a dummy CharsetReader
  decoder.CharsetReader =
    func (charset string, input io.Reader) (io.Reader, error){
      return input, nil
    }
  err = decoder.Decode(&mz)
  if err != nil {
    return err
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
    scanInfo(s, &(mz.Run.Scans[i]), 0)
    scans = append(scans, *s)
    if len(mz.Run.Scans[i].Scans) > 0 {
      for j:=0; j < len(mz.Run.Scans[i].Scans); j++ {
        s := new(Scan)
        scanInfo(s, &(mz.Run.Scans[i].Scans[j]), mz.Run.Scans[i].Id)
        scans = append(scans, *s)
      }
    }
    r.Scans = scans
  }
  return nil
}

func (r *RawData) WriteMzXml(filename string) error {
  return nil
}

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
  decoder := base64.NewEncoding("encodeStd")
  peakData,_ := decoder.DecodeString(mzs.Peaks.PeakList)
  println(mzs.Peaks.PeakList)
  println(len(peakData)) // something wrong here
  s.MzArray = make([]float64,mzs.PeakCount,mzs.PeakCount)
  s.IntensityArray = make([]float64,mzs.PeakCount,mzs.PeakCount)
  decodePeaks(&s.MzArray, &s.IntensityArray, peakData,
              mzs.Peaks.Precision, mzs.Peaks.ByteOrder, mzs.Peaks.PairOrder)
}

func decodePeaks(mz *[]float64, intensity *[]float64, peakData []byte,
                 precision uint8, byteorder string, pairOrder string) {
  // this whole thing is a bit messy, could probably be done more efficiently
  // assume big endian for now (byteOrder "network")
  pos := 0
  if precision == 32 {
    const typesize int = 4
    for pos < len(peakData) {
      var value uint32 = 0
      for i := 0; i < typesize; i++ {
        value <<= 8
        value = value | uint32(peakData[pos])
        pos++
      }
      if pairOrder == "m/z-int" {
        (*mz)[pos % typesize] = float64(math.Float32frombits(value))
      } else {
        (*intensity)[pos % typesize] = float64(math.Float32frombits(value))
      }
      value = 0
      for i := 0; i < typesize; i++ {
        value <<= 8
        value |= uint32(peakData[pos])
        pos++
      }
      if pairOrder == "m/z-int" {
        (*intensity)[pos % typesize] = float64(math.Float32frombits(value))
      } else {
        (*mz)[pos % typesize] = float64(math.Float32frombits(value))
      }
    }
  } else {
//    var value uint64 = 0
//    const typesize int = 8
//    for i := 0; i < typesize; i++ {
//      value <<= 8
//      value |= uint64(peakData[pos])
//      pos++
//    }
//    if pairOrder == "m/z-int" {
//      (*mz)[pos % typesize] = math.Float64frombits(value)
//    } else {
//      (*intensity)[pos % typesize] = math.Float64frombits(value)
//    }
//    value = 0
//    for i := 0; i < typesize; i++ {
//      value <<= 8
//      value |= uint64(peakData[pos])
//      pos++
//    }
//    if pairOrder == "m/z-int" {
//      (*intensity)[pos % typesize] = math.Float64frombits(value)
//    } else {
//      (*mz)[pos % typesize] = math.Float64frombits(value)
//    }
  }
}



