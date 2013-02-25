package mzlib

import (
  "errors"
  "io"
)

func (r *RawData) ReadJson(filename string) error {
  return errors.New("Reading this file type has not yet been implemented")
}

func (r *RawData) DecodeJson(reader io.Reader) error {
  return errors.New("Reading this file type has not yet been implemented")
}

func (r *RawData) WriteJson(filename string) error {
  return errors.New("Writing this file type has not yet been implemented")
}

func (r *RawData) EncodeJson(writer io.Writer) error {
  return errors.New("Writing this file type has not yet been implemented")
}





func (r *RawData) ReadJsonGz(filename string) error {
  return errors.New("Reading this file type has not yet been implemented")
}

func (r *RawData) DecodeJsonGz(reader io.Reader) error {
  return errors.New("Reading this file type has not yet been implemented")
}

func (r *RawData) WriteJsonGz(filename string) error {
  return errors.New("Writing this file type has not yet been implemented")
}

func (r *RawData) EncodeJsonGz(writer io.Writer) error {
  return errors.New("Writing this file type has not yet been implemented")
}

