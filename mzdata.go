//  Copyright 2013 Thomas McGrew
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

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
  "path/filepath"
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
  r.Filename,_ = filepath.Abs(filename)
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
  var chans []chan *Scan
  for i := range mz.SpectrumList.Scans {
    c := make(chan *Scan)
    go mz.SpectrumList.Scans[i].scanInfo(c)
    chans = append(chans, c)
  }
  // wait for everything to finish
  for _,c := range chans {
    s := <-c
    iso,_ := param(&mz.ProcessingMethod, "Deisotoping")
    // sanity check
    if len(s.MzArray) != len(s.IntensityArray) {
      panic(fmt.Sprintf(
            "Lengths of Intensity and MZ do not match! Scan %d, %d vs %d",
            s.Id, len(s.IntensityArray), len(s.MzArray)))
    }
    s.DeIsotoped,_ = strconv.ParseBool(iso)
    r.Scans = append(r.Scans, *s)
  }
  return nil
}

func (scan *mzDataScan) scanInfo(c chan *Scan) {
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
  if len(scan.Precursor) > 0 {
    s.ParentScan = scan.Precursor[0].ParentScan
    mass,_ := param(&scan.Precursor[0].IonSelection, "MassToChargeRatio")
    s.PrecursorMz,_ = strconv.ParseFloat(mass, 64)
    ce,_ := param(&scan.Precursor[0].Activation, "CollisionEnergy")
    s.CollisionEnergy,_ = strconv.ParseFloat(ce, 64)
  }
  s.Continuous = scan.Specification.SpectrumType == "continuous"
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
  c <- s
}

// Writes the data to disk in MzData format
//
// Parameters:
//   filename: The name of the file to be written to
//
// Return value:
//   error: Indicates whether or not an error occurred while writing the file
func (r *RawData) WriteMzData( filename string ) error {
  outFile,err := os.OpenFile(filename,
                             os.O_WRONLY | os.O_CREATE | os.O_TRUNC,
                             0770)
  if err != nil {
    return err;
  }
  out := bufio.NewWriter(outFile)
  defer outFile.Close()
  err = r.EncodeMzData(out)
  if err != nil {
    return err;
  }
  out.Flush()
  return nil
}

func (r *RawData) EncodeMzData(writer io.Writer) error {
  deIsotoped := (len(r.Scans) > 0 && r.Scans[0].DeIsotoped)
  var sourceFileName string
  var sourceFilePath string
  pathIndex := strings.LastIndex(r.Filename, "/")
  if pathIndex >= 0 {
    sourceFileName = r.Filename[pathIndex+1:]
    sourceFilePath = r.Filename[:pathIndex]
  } else {
    sourceFileName = r.Filename
  }
  _,err := writer.Write(([]byte)(fmt.Sprintf(
`<?xml version="1.0" encoding="UTF-8"?>
<mzData version="1.05" accessionNumber="psi-ms:100" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <cvLookup cdLabel="psi" fullName="The PSI Ontology" version="1.00" address="http://psidev.sourceforge.net/ontology" />
  <description>
    <admin>
      <sampleName/>
      <sampleDescription comment="" />
      <sourceFile>
        <nameOfFile>%s</nameOfFile>
        <pathToFile>%s</pathToFile>
      </sourceFile>
      <contact>
        <name />
        <institution />
        <contactInfo />
      </contact>
    </admin>
    <instrument>
      <instrumentName>%s</instrumentName>
      <source />
      <analyzerList count="1">
        <analyzer>
          <cvParam cvLabel="psi" accession="PSI:1000010" name="AnalyzerType" value="%s" />
        </analyzer>
      </analyzerList>
      <detector>
          <cvParam cvLabel="psi" accession="PSI:1000026" name="DetectorType" value="unknown" />
          <cvParam cvLabel="psi" accession="PSI:1000029" name="SamplingFrequency" value="unknown" />
      </detector>
    </instrument>
    <dataProcessing>
      <software>
        <name>gomzlib, Version=%s</name>
        <version>%s</version>
        <comments />
      </software>
      <processingMethod>
          <cvParam cvLabel="psi" accession="PSI:1000033" name="deisotoped" value="%t" />
          <cvParam cvLabel="psi" accession="PSI:1000034" name="chargeDeconvolved" value="unknown" />
          <cvParam cvLabel="psi" accession="PSI:1000035" name="peakProcessing" value="unknown" />
      </processingMethod>
    </dataProcessing>
  </description>
  <spectrumList count="%d">`, sourceFileName, sourceFilePath,
    r.Instrument.Model, r.Instrument.MassAnalyzer, Version, Version,
    deIsotoped, len(r.Scans))))
  if err != nil {
    return err
  }
  for _,scan := range r.Scans {
    var polarity string
    if scan.Polarity > 0 {
      polarity = "positive"
    } else {
      polarity = "negative"
    }
    mzBase64 := Base64FromFloat64(&scan.MzArray, 64, binary.LittleEndian)
    intensityBase64 := Base64FromFloat64(&scan.IntensityArray, 64,
                                         binary.LittleEndian)
    var spectrumType string
    method := ""
    if scan.Continuous {
      spectrumType = "continuous"
    } else {
      spectrumType = "discrete"
      method = ` methodOfCombination="sum"`
    }
    _,err = writer.Write(([]byte)(fmt.Sprintf(`
    <spectrum id="%d">
        <spectrumDesc>
          <spectrumSettings>
            <acqSpecification spectrumType="%s"%s count="1">
              <acquisition number="%d" />
            </acqSpecification>
            <spectrumInstrument msLevel="%d" mzRangeStart="%f" mzRangeStop="%f">
              <cvParam cvLabel="psi" accession="PSI:1000036" name="ScanMode" value="Scan" />
              <cvParam cvLabel="psi" accession="PSI:1000037" name="Polarity" value="%s" />
              <cvParam cvLabel="psi" accession="PSI:1000038" name="TimeInMinutes" value="%f" />
            </spectrumInstrument>
          </spectrumSettings>`, scan.Id, spectrumType, method, scan.Id,
            scan.MsLevel, scan.MzRange[0], scan.MzRange[1], polarity,
            scan.RetentionTime)))
    if err != nil {
      return err
    }
    if scan.ParentScan != 0 {
      _,err = writer.Write(([]byte)(fmt.Sprintf(`
				<precursorList count = "1">
					<precursor msLevel="%d" spectrumRef="%d">
						<ionSelection>
							<cvParam cvLabel="psi" accession="PSI:1000040" name="MassToChargeRatio" value="%f"/>
						</ionSelection>
						<activation>
							<cvParam cvLabel="psi" accession="PSI:1000045" name="CollisionEnergy" value="%f"/>
						</activation>
					</precursor>
				</precursorList>`, scan.MsLevel - 1, scan.ParentScan, scan.PrecursorMz,
          scan.CollisionEnergy)))
      if err != nil {
        return err
      }
    }
    _,err = writer.Write(([]byte)(fmt.Sprintf(`
        </spectrumDesc>
        <mzArrayBinary>
          <data precision="64" endian="little" length="%d">%s</data>
        </mzArrayBinary>
        <intenArrayBinary>
          <data precision="64" endian="little" length="%d">%s</data>
        </intenArrayBinary>
      </spectrum>`, len(scan.MzArray), mzBase64, len(scan.IntensityArray),
        intensityBase64 )))
    if err != nil {
      return err
    }
  }
  _,err = writer.Write(([]byte)(
`  </spectrumList>
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

