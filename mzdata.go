package mzlib

import (
  "errors"
)

func (r *RawData) ReadMzData( filename string ) error {
  return errors.New("Reading this file type has not yet been implemented")
}

func (r *RawData) WriteMzData( filename string ) error {
  return errors.New("Writing this file type has not yet been implemented")
}

