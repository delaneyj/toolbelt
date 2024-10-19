package natsrpc

import (
	"fmt"

	"github.com/delaneyj/toolbelt"
	"google.golang.org/protobuf/compiler/protogen"
)

func Generate(gen *protogen.Plugin, file *protogen.File) error {

	pkgData, err := optsToPackageData(file)
	if err != nil {
		return fmt.Errorf("failed to convert options to data: %w", err)
	}

	if pkgData == nil {
		return nil
	}

	if err := generateGoFile(gen, pkgData); err != nil {
		return fmt.Errorf("failed to generate file: %w", err)
	}

	return nil
}

type methodTmplData struct {
	ServiceName, Name                    toolbelt.CasedString
	Subject                              string
	IsClientStreaming, IsServerStreaming bool
	InputType, OutputType                toolbelt.CasedString
}

type serviceTmplData struct {
	Name    toolbelt.CasedString
	Subject string
	Methods []*methodTmplData
}

type packageTmplData struct {
	GoImportPath protogen.GoImportPath
	FileBasepath string
	PackageName  toolbelt.CasedString
	Services     []*serviceTmplData
}

func optsToPackageData(file *protogen.File) (data *packageTmplData, err error) {
	if len(file.Services) == 0 {
		return nil, nil
	}

	// log.Printf("Generating package %+v", file)
	data = &packageTmplData{
		GoImportPath: file.GoImportPath,
		FileBasepath: file.GeneratedFilenamePrefix + "_nrpc.pb",
		PackageName:  toolbelt.ToCasedString(string(file.GoPackageName)),
		Services:     make([]*serviceTmplData, 0, len(file.Services)),
	}

	for _, s := range file.Services {
		if len(s.Methods) == 0 {
			continue
		}

		// log.Printf("Generating service %+v", s)
		sn := toolbelt.ToCasedString(s.GoName)
		svcData := &serviceTmplData{
			Name:    sn,
			Subject: "natsrpc." + sn.Kebab,
			Methods: make([]*methodTmplData, len(s.Methods)),
		}

		for i, m := range s.Methods {
			mn := toolbelt.ToCasedString(string(m.Desc.Name()))
			methodData := &methodTmplData{
				Name:              mn,
				ServiceName:       sn,
				Subject:           svcData.Subject + "." + mn.Kebab,
				IsClientStreaming: m.Desc.IsStreamingClient(),
				IsServerStreaming: m.Desc.IsStreamingServer(),
				InputType:         toolbelt.ToCasedString(m.Input.GoIdent.GoName),
				OutputType:        toolbelt.ToCasedString(m.Output.GoIdent.GoName),
			}
			svcData.Methods[i] = methodData
		}

		data.Services = append(data.Services, svcData)
	}

	return data, nil
}

func generateGoFile(gen *protogen.Plugin, data *packageTmplData) error {
	// log.Printf("Generating package %+v", data)

	files := map[string]string{
		data.FileBasepath + "_shared.go": goSharedTypesTemplate(data),
		data.FileBasepath + "_server.go": goServerTemplate(data),
		data.FileBasepath + "_client.go": goClientTemplate(data),
	}

	for filename, contents := range files {
		// log.Printf("Writing to file %s", filename)

		g := gen.NewGeneratedFile(filename, data.GoImportPath)
		if _, err := g.Write([]byte(contents)); err != nil {
			return fmt.Errorf("failed to write to file: %w", err)
		}
	}

	return nil
}
