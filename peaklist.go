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
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/binary"
	"io/ioutil"
	"strings"
)

// Converts a base64 string to an array of float64
//
// Parameters:
//   dst: The destination array.
//   src: The base64 encoded source string.
//   peakCount: The number of peaks present in the encoded string.
//   precision: The number of bits in each value, either 32 or 64.
//   compressed: Whether or not the data is compressed with zlib.
//   byteOrder: The byte order of the data, either binary.BigEndian or
//     binary.LittleEndian
func Float64FromBase64(dst *[]float64, src string, peakCount uint64,
	precision uint8, compressed bool,
	byteOrder binary.ByteOrder) int {
	sr := strings.NewReader(src)
	reader := base64.NewDecoder(base64.StdEncoding, sr)
	if compressed {
		reader, _ = zlib.NewReader(reader)
	}
	if precision == 32 {
		for i := uint64(0); i < peakCount; i++ {
			var value float32
			binary.Read(reader, byteOrder, &value)
			*dst = append(*dst, float64(value))
		}
	} else if precision == 64 {
		for i := uint64(0); i < peakCount; i++ {
			var value float64
			binary.Read(reader, byteOrder, &value)
			*dst = append(*dst, value)
		}
	}
	return 0
}

// Converts an array of float64 to a base64 string
func Base64FromFloat64(src *[]float64, precision int,
	byteOrder binary.ByteOrder) string {
	dst := new(bytes.Buffer)
	writer := base64.NewEncoder(base64.StdEncoding, dst)
	if precision == 64 {
		for _, v := range *src {
			binary.Write(writer, byteOrder, v)
		}
	} else if precision == 32 {
		for _, v := range *src {
			binary.Write(writer, byteOrder, float32(v))
		}
	}
	writer.Close()
	result, _ := ioutil.ReadAll(dst)
	return string(result)
}
