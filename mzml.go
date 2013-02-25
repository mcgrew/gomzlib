package mzlib

import (
  "errors"
  "io"
)

func (r *RawData) ReadMzMl(filename string) error {
  return errors.New("Reading this file type has not yet been implemented")
}

func (r *RawData) DecodeMzMl(reader io.Reader) error {
  return errors.New("Reading this file type has not yet been implemented")
}

func (r *RawData) WriteMzMl(filename string) error {
  return errors.New("Writing this file type has not yet been implemented")
}

func (r *RawData) EncodeMzMl(writer io.Writer) error {
  return errors.New("Writing this file type has not yet been implemented")
}

