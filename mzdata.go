package mzlib

import (
  "encoding/xml"
  "strconv"
  "errors"
  "os"
  "io"
  "fmt"
  "strings"
  "bufio"
  "encoding/binary"
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
  file,err := os.OpenFile(filename, os.O_RDONLY, 0777)
  if err != nil {
    return err
  }
  defer file.Close()
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
    var byteOrder binary.ByteOrder
    if scan.MzArray.Endian == "big" {
      byteOrder = binary.BigEndian
    } else {
      byteOrder = binary.LittleEndian
    }
    _ = Float64FromBase64(&s.MzArray, scan.MzArray.PeakList,
                          scan.MzArray.PeakCount, scan.MzArray.Precision,
                          false, byteOrder)
    _ = Float64FromBase64(&s.IntensityArray, scan.IntensityArray.PeakList,
                          scan.IntensityArray.PeakCount,
                          scan.IntensityArray.Precision,
                          false, byteOrder)
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
  outFile,err := os.OpenFile(filename, os.O_WRONLY, 0770)
  if err != nil {
    return err;
  }
  defer outFile.Close()
  out := bufio.NewWriter(outFile)
  _,err = out.Write(([]byte)(fmt.Sprintf(
`<?xml version="1.0" encoding="UTF-8"?>
<mzData version="1.05" accessionNumber="psi-ms:100" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <cvLookup cdLabel="psi" fullName="The PSI Ontology" version="1.00" address="http://psidev.sourceforge.net/ontology" />
  <description>
    <admin>
      <sampleName/>
      <sampleDescription comment="" />
      <sourceFile>\n
        <nameOfFile />\n
        <pathToFile />\n
      </sourceFile>\n
      <contact>\n
        <name />\n
        <institution />\n
        <contactInfo />\n
      </contact>\n
    </admin>\n
    <instrument>\n
      <instrumentName />\n
      <source />\n
      <analyzerList count="1">\n
        <analyzer>\n
          <cvParam cvLabel="psi" accession="PSI:1000010" name="AnalyzerType" value="unknown" />\n
        </analyzer>\n
      </analyzerList>\n
      <detector>\n
          <cvParam cvLabel="psi" accession="PSI:1000026" name="DetectorType" value="unknown" />\n
          <cvParam cvLabel="psi" accession="PSI:1000029" name="SamplingFrequency" value="unknown" />\n
      </detector>\n
      <additional />\n
    </instrument>\n
    <dataProcessing>\n
      <software completionTime="">\n
        <name>pymzlib, Version=%s</name>\n
        <version>%s</version>\n
        <comments />\n
      </software>\n
      <processingMethod>\n
          <cvParam cvLabel="psi" accession="PSI:1000033" name="deisotoped" value="unknown" />\n
          <cvParam cvLabel="psi" accession="PSI:1000034" name="chargeDeconvolved" value="unknown" />\n
          <cvParam cvLabel="psi" accession="PSI:1000035" name="peakProcessing" value="unknown" />\n
      </processingMethod>\n
    </dataProcessing>\n
  </description>\n
  <spectrumList count="%d">`, Version, Version, len(r.Scans))))
  if err != nil {
    return err
  }
  for _,scan := range r.Scans {
    var polarity string
    if scan.Polarity > 0 {
      polarity = "Positive"
    } else {
      polarity = "Negative"
    }
    var mzBase64 string
    var intensityBase64 string
    _,err = out.Write(([]byte)(fmt.Sprintf(`
    <spectrum id="%d">
        <spectrumDesc>\n
          <spectrumSettings>\n
            <acqSpecification spectrumType="unknown" methodOfCombination="unknown" count="1">\n
              <acquisition number="%d" />
            </acqSpecification>
            <spectrumInstrument msLevel="%d" mzRangeStart="%f" mzRangeStop="%f">
              <cvParam cvLabel="psi" accession="PSI:1000036" name="ScanMode" value="Scan" />
              <cvParam cvLabel="psi" accession="PSI:1000037" name="Polarity" value="%s" />
              <cvParam cvLabel="psi" accession="PSI:1000038" name="TimeInMinutes" value="%f" />
            </spectrumInstrument>
          </spectrumSettings>
        </spectrumDesc>
        <mzArrayBinary>
          <data precision="64" endian="little" length="%d">%s</data>
        </mzArrayBinary>
        <intenArrayBinary>
          <data precision="64" endian="little" length="%d">%s</data>
        </intenArrayBinary>
      </spectrum>`, scan.Id, scan.MsLevel, scan.MzRange[0], scan.MzRange[1],
      polarity, scan.RetentionTime, len(scan.MzArray), mzBase64,
      len(scan.IntensityArray), intensityBase64 )))
    if err != nil {
      return err
    }
  }
  _,err = out.Write(([]byte)(
`  </spectrumList>\n
</mzData>`))
  if err != nil {
    return err
  }
  return nil
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

