// Copyright 2016 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package template

// InterfaceTemplate defines the template used to generate the adapter
// interfaces for Mixer for a given aspect.
var InterfaceTemplate = `// Copyright 2017 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// THIS FILE IS AUTOMATICALLY GENERATED.

package {{.PkgName}}

import (
	"github.com/gogo/protobuf/proto"
	"fmt"
	"context"
	"istio.io/mixer/pkg/attribute"
	"istio.io/mixer/pkg/expr"
	"istio.io/mixer/pkg/adapter"
	"istio.io/api/mixer/v1/config/descriptor"
	"istio.io/mixer/pkg/template"
	"github.com/golang/glog"
	adptTmpl "istio.io/api/mixer/v1/template"
	"errors"
	{{range .TemplateModels}}
		"{{.PackageImportPath}}"
	{{end}}
	$$additional_imports$$
)


const emptyQuotes = "\"\""
var (
	SupportedTmplInfo = map[string]template.Info {
	{{range .TemplateModels}}
		{{.GoPackageName}}.TemplateName: {
			Name: {{.GoPackageName}}.TemplateName,
			Impl: "{{.PackageName}}",
			CtrCfg:  &{{.GoPackageName}}.InstanceParam{},
			Variety:   adptTmpl.{{.VarietyName}},
			BldrInterfaceName:  {{.GoPackageName}}.TemplateName + "." + "HandlerBuilder",
			HndlrInterfaceName: {{.GoPackageName}}.TemplateName + "." + "Handler",
			BuilderSupportsTemplate: func(hndlrBuilder adapter.HandlerBuilder) bool {
				_, ok := hndlrBuilder.({{.GoPackageName}}.HandlerBuilder)
				return ok
			},
			HandlerSupportsTemplate: func(hndlr adapter.Handler) bool {
				_, ok := hndlr.({{.GoPackageName}}.Handler)
				return ok
			},
			InferType: func(cp proto.Message, tEvalFn template.TypeEvalFn) (proto.Message, error) {
				var err error = nil
				cpb := cp.(*{{.GoPackageName}}.InstanceParam)
				infrdType := &{{.GoPackageName}}.Type{}

				{{range .TemplateMessage.Fields}}
					{{if containsValueType .GoType}}
						{{if .GoType.IsMap}}
							infrdType.{{.GoName}} = make(map[{{.GoType.MapKey.Name}}]istio_mixer_v1_config_descriptor.ValueType, len(cpb.{{.GoName}}))
							for k, v := range cpb.{{.GoName}} {
								if infrdType.{{.GoName}}[k], err = tEvalFn(v); err != nil {
									return nil, err
								}
							}
						{{else}}
							if cpb.{{.GoName}} == "" || cpb.{{.GoName}} == emptyQuotes {
								return nil, errors.New("expression for field {{.GoName}} cannot be empty")
							}
							if infrdType.{{.GoName}}, err = tEvalFn(cpb.{{.GoName}}); err != nil {
								return nil, err
							}
						{{end}}
					{{else}}
						{{if .GoType.IsMap}}
							for _, v := range cpb.{{.GoName}} {
								if t, e := tEvalFn(v); e != nil || t != {{getValueType .GoType.MapValue}} {
									if e != nil {
										return nil, fmt.Errorf("failed to evaluate expression for field {{.GoName}}: %v", e)
									}
									return nil, fmt.Errorf("error type checking for field {{.GoName}}: Evaluated expression type %v want %v", t, {{getValueType .GoType.MapValue}})
								}
							}
						{{else}}
							if cpb.{{.GoName}} == "" || cpb.{{.GoName}} == emptyQuotes {
								return nil, errors.New("expression for field {{.GoName}} cannot be empty")
							}
							if t, e := tEvalFn(cpb.{{.GoName}}); e != nil || t != {{getValueType .GoType}} {
								if e != nil {
									return nil, fmt.Errorf("failed to evaluate expression for field {{.GoName}}: %v", e)
								}
								return nil, fmt.Errorf("error type checking for field {{.GoName}}: Evaluated expression type %v want %v", t, {{getValueType .GoType}})
							}
						{{end}}
					{{end}}
				{{end}}
				_ = cpb
				return infrdType, err
			},
			SetType: func(types map[string]proto.Message, builder adapter.HandlerBuilder) {
				// Mixer framework should have ensured the type safety.
				castedBuilder := builder.({{.GoPackageName}}.HandlerBuilder)
				castedTypes := make(map[string]*{{.GoPackageName}}.Type, len(types))
				for k, v := range types {
					// Mixer framework should have ensured the type safety.
					v1 := v.(*{{.GoPackageName}}.Type)
					castedTypes[k] = v1
				}
				castedBuilder.Set{{.InterfaceName}}Types(castedTypes)
			},
			{{if eq .VarietyName "TEMPLATE_VARIETY_REPORT"}}
				ProcessReport: func(ctx context.Context, insts map[string]proto.Message, attrs attribute.Bag, mapper expr.Evaluator, handler adapter.Handler) error {
					var instances []*{{.GoPackageName}}.Instance
					for name, inst := range insts {
						md := inst.(*{{.GoPackageName}}.InstanceParam)
						{{range .TemplateMessage.Fields}}
							{{if .GoType.IsMap}}
								{{.GoName}}, err := template.EvalAll(md.{{.GoName}}, attrs, mapper)
							{{else}}
								{{.GoName}}, err := mapper.Eval(md.{{.GoName}}, attrs)
							{{end}}
								if err != nil {
									msg := fmt.Sprintf("failed to eval {{.GoName}} for instance '%s': %v", name, err)
									glog.Error(msg)
									return errors.New(msg)
								}
						{{end}}

						instances = append(instances, &{{.GoPackageName}}.Instance{
							Name:       name,
							{{range .TemplateMessage.Fields}}
								{{if containsValueType .GoType}}
									{{.GoName}}: {{.GoName}},
								{{else}}
									{{if .GoType.IsMap}}
										{{.GoName}}: func(m map[string]interface{}) map[string]{{.GoType.MapValue.Name}} {
											res := make(map[string]{{.GoType.MapValue.Name}}, len(m))
											for k, v := range m {
												res[k] = v.({{.GoType.MapValue.Name}})
											}
											return res
										}({{.GoName}}),
									{{else}}
										{{.GoName}}: {{.GoName}}.({{.GoType.Name}}),{{reportTypeUsed .GoType}}
									{{end}}
								{{end}}
							{{end}}
						})
						_ = md
					}

					if err := handler.({{.GoPackageName}}.Handler).Handle{{.InterfaceName}}(ctx, instances); err != nil {
						return fmt.Errorf("failed to report all values: %v", err)
					}
					return nil
				},
			{{else if eq .VarietyName "TEMPLATE_VARIETY_CHECK"}}
				ProcessCheck: func(ctx context.Context, instName string, inst proto.Message, attrs attribute.Bag,
				mapper expr.Evaluator, handler adapter.Handler) (adapter.CheckResult, error) {
					castedInst := inst.(*{{.GoPackageName}}.InstanceParam)
					{{range .TemplateMessage.Fields}}
						{{if .GoType.IsMap}}
							{{.GoName}}, err := template.EvalAll(castedInst.{{.GoName}}, attrs, mapper)
						{{else}}
							{{.GoName}}, err := mapper.Eval(castedInst.{{.GoName}}, attrs)
						{{end}}
							if err != nil {
								msg := fmt.Sprintf("failed to eval {{.GoName}} for instance '%s': %v", instName, err)
								glog.Error(msg)
								return adapter.CheckResult{}, errors.New(msg)
							}
					{{end}}
					_ = castedInst

					instance := &{{.GoPackageName}}.Instance{
						Name:	instName,
						{{range .TemplateMessage.Fields}}
							{{if containsValueType .GoType}}
								{{.GoName}}: {{.GoName}},
							{{else}}
								{{if .GoType.IsMap}}
									{{.GoName}}: func(m map[string]interface{}) map[string]{{.GoType.MapValue.Name}} {
										res := make(map[string]{{.GoType.MapValue.Name}}, len(m))
										for k, v := range m {
											res[k] = v.({{.GoType.MapValue.Name}})
										}
										return res
									}({{.GoName}}),
								{{else}}
									{{.GoName}}: {{.GoName}}.({{.GoType.Name}}),{{reportTypeUsed .GoType}}
								{{end}}
							{{end}}
						{{end}}
					}
					return handler.({{.GoPackageName}}.Handler).Handle{{.InterfaceName}}(ctx, instance)
				},
			{{else}}
				ProcessQuota: func(ctx context.Context, quotaName string, inst proto.Message, attrs attribute.Bag,
				 mapper expr.Evaluator, handler adapter.Handler, args adapter.QuotaArgs) (adapter.QuotaResult, error) {
					castedInst := inst.(*{{.GoPackageName}}.InstanceParam)
					{{range .TemplateMessage.Fields}}
						{{if .GoType.IsMap}}
							{{.GoName}}, err := template.EvalAll(castedInst.{{.GoName}}, attrs, mapper)
						{{else}}
							{{.GoName}}, err := mapper.Eval(castedInst.{{.GoName}}, attrs)
						{{end}}
							if err != nil {
								msg := fmt.Sprintf("failed to eval {{.GoName}} for instance '%s': %v", quotaName, err)
								glog.Error(msg)
								return adapter.QuotaResult{}, errors.New(msg)
							}
					{{end}}

					instance := &{{.GoPackageName}}.Instance{
						Name:       quotaName,
						{{range .TemplateMessage.Fields}}
							{{if containsValueType .GoType}}
								{{.GoName}}: {{.GoName}},
							{{else}}
								{{if .GoType.IsMap}}
									{{.GoName}}: func(m map[string]interface{}) map[string]{{.GoType.MapValue.Name}} {
										res := make(map[string]{{.GoType.MapValue.Name}}, len(m))
										for k, v := range m {
											res[k] = v.({{.GoType.MapValue.Name}})
										}
										return res
									}({{.GoName}}),
								{{else}}
									{{.GoName}}: {{.GoName}}.({{.GoType.Name}}),{{reportTypeUsed .GoType}}
								{{end}}
							{{end}}
						{{end}}
					}

					return handler.({{.GoPackageName}}.Handler).Handle{{.InterfaceName}}(ctx, instance, args)
				},
			{{end}}

		},
	{{end}}
	}
)

`
