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
