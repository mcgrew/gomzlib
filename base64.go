package mzlib

import( 
  "strings"
  "math"
)

// Converts a base64 string to an array of float64
func Float64FromBase64 (dst *[]float64, src string, 
                        precision uint8, byteOrder bool) int {
  n := len(src)
  shift := uint8(0)
  if precision == 32 {
    var value uint32 = 0
    for i :=0; i < n; i++ {
      if (src[i] == uint8(61)) { // the = sign (padding)
        break
      }
      bits := uint32(strings.IndexAny(charSet, string(src[i])))
      shift += 6
      if shift > precision {
        shift %= precision
        value <<= (6 - shift)
        value |= bits >> shift
        if !byteOrder {
          value = invertBytes32(value)
        }
        *dst = append(*dst, float64(math.Float32frombits(value)))
        value = bits
      } else {
        value <<= 6
        value |= bits
      }
    }
  } else if precision == 64 {
    var value uint64 = 0
    for i :=0; i < n; i++ {
      bits := uint64(strings.IndexAny(charSet, string(src[n])))
      shift += 6
      if shift > precision {
        shift %= precision
        value <<= (6 - shift)
        value |= bits >> shift
        if !byteOrder {
          value = invertBytes64(value)
        }
        *dst = append(*dst, math.Float64frombits(value))
        value = bits
      } else {
        value <<= 6
        value |= bits
      }
    }
  }
  return 0
}

// Converts an array of float64 to a base64 string
func Base64FromFloat64(src *[]float64, precision int, byteOrder bool) string {
  return ""
}

func invertBytes32( value uint32 ) uint32 {
  value = ((value >> 24) & 0x000000FF) |
          ((value >>  8) & 0x0000FF00) |
          ((value <<  8) & 0x00FF0000) |
          ((value << 24) & 0xFF000000)
  return value
}

func invertBytes64( value uint64 ) uint64 {
  value = ((value >> 56) & 0x00000000000000FF) |
          ((value >> 40) & 0x000000000000FF00) |
          ((value >> 24) & 0x0000000000FF0000) |
          ((value >>  8) & 0x00000000FF000000) |
          ((value <<  8) & 0x000000FF00000000) |
          ((value << 24) & 0x0000FF0000000000) |
          ((value << 40) & 0x00FF000000000000) |
          ((value << 56) & 0xFF00000000000000)
  return value
}

const LittleEndian bool = false
const BigEndian bool = true

const charSet string = 
  "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

