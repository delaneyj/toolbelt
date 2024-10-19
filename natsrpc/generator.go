package natsrpc

import (
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/delaneyj/toolbelt"
	ext "github.com/delaneyj/toolbelt/natsrpc/protos/natsrpc"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/durationpb"
)

var isFirst = true

func Generate(gen *protogen.Plugin, file *protogen.File) error {

	pkgData, err := optsToPackageData(file)
	if err != nil {
		return fmt.Errorf("failed to convert options to data: %w", err)
	}

	if pkgData == nil {
		return nil
	}

	if isFirst {
		isFirst = false
		sharedFilepath := filepath.Join(filepath.Dir(pkgData.FileBasepath), "natsrpc_shared.go")
		log.Printf("Writing to file %s", sharedFilepath)

		sharedContent := goSharedTypesTemplate(pkgData)
		g := gen.NewGeneratedFile(sharedFilepath, pkgData.GoImportPath)
		if _, err := g.Write([]byte(sharedContent)); err != nil {
			return fmt.Errorf("failed to write to file: %w", err)
		}
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

type kvTemplData struct {
	PackageName      toolbelt.CasedString
	Name             toolbelt.CasedString
	Bucket           string
	IsClientReadonly bool
	TTL              time.Duration
	ID               toolbelt.CasedString
	IdIsString       bool
	HistoryCount     uint32
}

type packageTmplData struct {
	GoImportPath protogen.GoImportPath
	FileBasepath string
	PackageName  toolbelt.CasedString
	Services     []*serviceTmplData
	KeyValues    []*kvTemplData
}

func optsToPackageData(file *protogen.File) (*packageTmplData, error) {
	if len(file.Services) == 0 {
		return nil, nil
	}

	// log.Printf("Generating package %+v", file)
	data := &packageTmplData{
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

	for _, msg := range file.Messages {
		kvBucket, ok := proto.GetExtension(msg.Desc.Options(), ext.E_KvBucket).(string)
		if !ok || kvBucket == "" {
			continue
		}

		isReadonly := proto.GetExtension(msg.Desc.Options(), ext.E_KvClientReadonly).(bool)
		ttl := proto.GetExtension(msg.Desc.Options(), ext.E_KvTtl).(*durationpb.Duration)
		historyCount := proto.GetExtension(msg.Desc.Options(), ext.E_KvHistoryCount).(uint32)

		var idField *protogen.Field
		for _, f := range msg.Fields {
			isID := proto.GetExtension(f.Desc.Options(), ext.E_KvId).(bool)
			if isID {
				idField = f
				break
			}
		}

		if idField == nil {
			for _, f := range msg.Fields {
				if f.Desc.Name() == "id" {
					idField = f
					break
				}
			}
		}

		if idField == nil {
			return nil, fmt.Errorf("no id field found in message %s", msg.Desc.Name())
		}

		kvData := &kvTemplData{
			PackageName:      data.PackageName,
			Name:             toolbelt.ToCasedString(string(msg.Desc.Name())),
			Bucket:           kvBucket,
			IsClientReadonly: isReadonly,
			TTL:              ttl.AsDuration(),
			ID:               toolbelt.ToCasedString(string(idField.Desc.Name())),
			IdIsString:       idField.Desc.Kind() == protoreflect.StringKind,
			HistoryCount:     historyCount,
		}

		data.KeyValues = append(data.KeyValues, kvData)
	}

	return data, nil
}

func generateGoFile(gen *protogen.Plugin, data *packageTmplData) error {
	// log.Printf("Generating package %+v", data)

	files := map[string]string{
		data.FileBasepath + "_server.go": goServerTemplate(data),
		data.FileBasepath + "_client.go": goClientTemplate(data),
	}

	if len(data.KeyValues) > 0 {
		files[data.FileBasepath+"_kv.go"] = goKVTemplate(data)
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
