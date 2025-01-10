// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package experimental

import (
	"bytes"
	"fmt"
	"io"

	goyaml "github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/google/yamlfmt"
	"github.com/mitchellh/mapstructure"
)

const ExperimentalFormatterType string = "experimental"

type ExperimentalFormatter struct {
	Config     *Config
	Features   yamlfmt.FeatureList
	CommentMap goyaml.CommentMap
}

// yamlfmt.Formatter interface

func (f *ExperimentalFormatter) Type() string {
	return ExperimentalFormatterType
}

func (f *ExperimentalFormatter) Format(input []byte) ([]byte, error) {
	buf := bytes.NewBuffer(input)
	d := f.newDecoder(buf)

	var n ast.Node
	err := d.Decode(&n)
	if err != nil {
		return nil, err
	}

	var result bytes.Buffer
	e := f.newEncoder(&result)
	err = e.Encode(&n)
	if err != nil {
		return nil, err
	}

	fmt.Println(result.String())

	return result.Bytes(), nil
}

func (f *ExperimentalFormatter) newDecoder(r io.Reader) *goyaml.Decoder {
	f.CommentMap = goyaml.CommentMap{}
	return goyaml.NewDecoder(r, goyaml.CommentToMap(f.CommentMap))
}

func (f *ExperimentalFormatter) newEncoder(w io.Writer) *goyaml.Encoder {
	return goyaml.NewEncoder(
		w,
		goyaml.WithComment(f.CommentMap),
		goyaml.Indent(2),
	)
}

func (f *ExperimentalFormatter) ConfigMap() (map[string]any, error) {
	configMap := map[string]any{}
	err := mapstructure.Decode(f.Config, &configMap)
	if err != nil {
		return nil, err
	}
	configMap["type"] = ExperimentalFormatterType
	return configMap, err
}
