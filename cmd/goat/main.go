// Copyright 2022 gorse Project Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/spf13/cobra"
	"modernc.org/cc/v3"
)

var supportedTypes = mapset.NewSet("int64_t", "long")

type TranslateUnit struct {
	Source     string
	Assembly   string
	Object     string
	GoAssembly string
	Go         string
	Package    string
	Options    []string
	Offset     int
}

func NewTranslateUnit(source string, outputDir string, options ...string) TranslateUnit {
	sourceExt := filepath.Ext(source)
	noExtSourcePath := source[:len(source)-len(sourceExt)]
	noExtSourceBase := filepath.Base(noExtSourcePath)
	return TranslateUnit{
		Source:     source,
		Assembly:   noExtSourcePath + ".s",
		Object:     noExtSourcePath + ".o",
		GoAssembly: filepath.Join(outputDir, noExtSourceBase+".s"),
		Go:         filepath.Join(outputDir, noExtSourceBase+".go"),
		Package:    filepath.Base(outputDir),
		Options:    options,
	}
}

// parseSource parse C source file and extract functions declarations.
func (t *TranslateUnit) parseSource() ([]Function, error) {
	// List include paths.
	includePaths, err := listIncludePaths()
	if err != nil {
		return nil, err
	}
	source, err := t.fixSource(t.Source)
	if err != nil {
		return nil, err
	}
	ast, err := cc.Parse(&cc.Config{}, nil, includePaths,
		[]cc.Source{{Name: t.Source, Value: source}})
	if err != nil {
		return nil, err
	}
	var functions []Function
	for _, nodes := range ast.Scope {
		if len(nodes) != 1 || nodes[0].Position().Filename != t.Source {
			continue
		}
		node := nodes[0]
		if declarator, ok := node.(*cc.Declarator); ok {
			funcIdent := declarator.DirectDeclarator
			if funcIdent.Case != cc.DirectDeclaratorFuncParam {
				continue
			}
			if function, err := t.convertFunction(funcIdent); err != nil {
				return nil, err
			} else {
				functions = append(functions, function)
			}
		}
	}
	sort.Slice(functions, func(i, j int) bool {
		return functions[i].Position < functions[j].Position
	})
	return functions, nil
}

func (t *TranslateUnit) generateGoStubs(functions []Function) error {
	// generate code
	var builder strings.Builder
	builder.WriteString(buildTags)
	builder.WriteString(fmt.Sprintf("package %v\n\n", t.Package))
	builder.WriteString("import \"unsafe\"\n")
	for _, function := range functions {
		builder.WriteString("\n//go:noescape\n")
		builder.WriteString(fmt.Sprintf("func %v(%s unsafe.Pointer)\n",
			function.Name, strings.Join(function.Parameters, ", ")))
	}

	// write file
	f, err := os.Create(t.Go)
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		if err = f.Close(); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}(f)
	_, err = f.WriteString(builder.String())
	return err
}

func (t *TranslateUnit) compile(args ...string) error {
	args = append(args,
		"-mno-red-zone",
		"-mstackrealign",
		"-mllvm",
		"-inline-threshold=1000",
		"-fno-asynchronous-unwind-tables",
		"-fno-exceptions",
		"-fno-rtti",
	)
	_, err := runCommand("clang", append([]string{"-S", "-c", t.Source, "-o", t.Assembly}, args...)...)
	if err != nil {
		return err
	}
	_, err = runCommand("clang", append([]string{"-c", t.Assembly, "-o", t.Object}, args...)...)
	return err
}

func (t *TranslateUnit) Translate() error {
	functions, err := t.parseSource()
	if err != nil {
		return err
	}
	if err = t.generateGoStubs(functions); err != nil {
		return err
	}
	if err = t.compile(t.Options...); err != nil {
		return err
	}
	assembly, err := parseAssembly(t.Assembly)
	if err != nil {
		return err
	}
	dump, _ := runCommand("objdump", "-d", t.Object, "--insn-width", "16")
	err = parseObjectDump(dump, assembly)
	if err != nil {
		return err
	}
	for i, name := range functions {
		functions[i].Lines = assembly[name.Name]
	}
	return generateGoAssembly(t.GoAssembly, functions)
}

// fixSource fixes compile errors in source.
func (t *TranslateUnit) fixSource(path string) (string, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	if runtime.GOARCH == "amd64" {
		t.Offset = -1
		var builder strings.Builder
		builder.WriteString("#define __STDC_HOSTED__ 1\n")
		builder.Write(bytes)
		return builder.String(), nil
	} else if runtime.GOARCH == "arm64" {
		var (
			builder     strings.Builder
			clauseCount int
		)
		for _, line := range strings.Split(string(bytes), "\n") {
			if strings.HasPrefix(line, "#include") {
				// Do nothing
			} else if strings.Contains(line, "{") {
				if clauseCount == 0 {
					builder.WriteString(line[:strings.Index(line, "{")+1])
				}
				clauseCount++
			} else if strings.Contains(line, "}") {
				clauseCount--
				if clauseCount == 0 {
					builder.WriteString(line[strings.Index(line, "}"):])
				}
			} else if clauseCount == 0 {
				builder.WriteString(line)
			}
			builder.WriteRune('\n')
		}
		return builder.String(), nil
	}
	return "", fmt.Errorf("unsupported arch: %s", runtime.GOARCH)
}

// listIncludePaths lists include paths used by clang.
func listIncludePaths() ([]string, error) {
	out, err := runCommand("bash", "-c", "echo | gcc -xc -E -v -")
	if err != nil {
		return nil, err
	}
	var start bool
	var paths []string
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "#include <...> search starts here:") {
			start = true
		} else if strings.HasPrefix(line, "End of search list.") {
			start = false
		} else if start {
			path := strings.TrimSpace(line)
			paths = append(paths, path)
		}
	}
	return paths, nil
}

type Function struct {
	Name       string
	Position   int
	Parameters []string
	Lines      []Line
}

// convertFunction extracts the function definition from cc.DirectDeclarator.
func (t *TranslateUnit) convertFunction(declarator *cc.DirectDeclarator) (Function, error) {
	params, err := t.convertFunctionParameters(declarator.ParameterTypeList.ParameterList)
	if err != nil {
		return Function{}, err
	}
	return Function{
		Name:       declarator.DirectDeclarator.Token.Value.String(),
		Position:   declarator.Position().Line,
		Parameters: params,
	}, nil
}

// convertFunctionParameters extracts function parameters from cc.ParameterList.
func (t *TranslateUnit) convertFunctionParameters(params *cc.ParameterList) ([]string, error) {
	declaration := params.ParameterDeclaration
	paramName := declaration.Declarator.DirectDeclarator.Token.Value
	paramType := declaration.DeclarationSpecifiers.TypeSpecifier.Token.Value
	isPointer := declaration.Declarator.Pointer != nil
	if !isPointer && !supportedTypes.Contains(paramType.String()) {
		position := declaration.Position()
		return nil, fmt.Errorf("%v:%v:%v: error: unsupported type: %v\n",
			position.Filename, position.Line+t.Offset, position.Column, paramType)
	}
	paramNames := []string{paramName.String()}
	if params.ParameterList != nil {
		if nextParamNames, err := t.convertFunctionParameters(params.ParameterList); err != nil {
			return nil, err
		} else {
			paramNames = append(paramNames, nextParamNames...)
		}
	}
	return paramNames, nil
}

// runCommand runs a command and extract its output.
func runCommand(name string, arg ...string) (string, error) {
	cmd := exec.Command(name, arg...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if output != nil {
			return "", errors.New(string(output))
		} else {
			return "", err
		}
	}
	return string(output), nil
}

var command = &cobra.Command{
	Use:  "goat source [-o output_directory]",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		output, _ := cmd.PersistentFlags().GetString("output")
		if output == "" {
			var err error
			if output, err = os.Getwd(); err != nil {
				_, _ = fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}
		var options []string
		machineOptions, _ := cmd.PersistentFlags().GetStringSlice("machine-option")
		for _, m := range machineOptions {
			options = append(options, "-m"+m)
		}
		optimizeLevel, _ := cmd.PersistentFlags().GetInt("optimize-level")
		options = append(options, fmt.Sprintf("-O%d", optimizeLevel))
		file := NewTranslateUnit(args[0], output, options...)
		if err := file.Translate(); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	},
}

func init() {
	command.PersistentFlags().StringP("output", "o", "", "output directory of generated files")
	command.PersistentFlags().StringSliceP("machine-option", "m", nil, "machine option for clang")
	command.PersistentFlags().IntP("optimize-level", "O", 0, "optimization level for clang")
}

func main() {
	if err := command.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// cd cmd/goat
// go run . ../../tensor/intrinsics/src/mul_avx2.c -O3 -mavx
