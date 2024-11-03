package quickstart

import (
	_ "embed"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/tools/goctl/util/pathx"
)

var (
	//go:embed idl/greet.proto
	protocContent string
	//go:embed idl/rpc.yaml
	rpcEtcContent string
	zrpcWorkDir   string
	protoName     = "greet.proto"
)

type serviceImpl struct {
	starter func()
}

func (s serviceImpl) Start() {
	s.starter()
}

func (s serviceImpl) Stop() {}

func toUpperFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	us := []rune(varStringServiceName)
	us[0] = unicode.ToUpper(us[0])
	return string(us)
}

func initRPCProto() error {
	zrpcWorkDir = filepath.Join(projectDir, "rpc")
	if err := pathx.MkdirIfNotExist(zrpcWorkDir); err != nil {
		return err
	}
	if varStringServiceName != "" && varStringServiceName != "greet" {
		protoName = varStringServiceName + ".proto"
		protocContent = strings.ReplaceAll(protocContent, "Greet", toUpperFirst(varStringServiceName))
		rpcEtcContent = strings.ReplaceAll(rpcEtcContent, "greet", varStringServiceName)
	}
	protoFilename := filepath.Join(zrpcWorkDir, protoName)
	rpcBytes := []byte(protocContent)
	return os.WriteFile(protoFilename, rpcBytes, 0o666)
}

type micro struct{}

func newMicroService() micro {
	m := micro{}
	m.mustStartRPCProject()
	return m
}

func (m micro) mustStartRPCProject() {
	logx.Must(initRPCProto())
	log.Debug(">> Generating quickstart zRPC project...")
	arg := "goctl rpc protoc " + protoName + " --go_out=. --go-grpc_out=. --zrpc_out=. --verbose"
	execCommand(zrpcWorkDir, arg)
	etcFile := filepath.Join(zrpcWorkDir, "etc", varStringServiceName)
	logx.Must(os.WriteFile(etcFile, []byte(rpcEtcContent), 0o666))
}

func (m micro) start() {
	mono := newMonoService(true)
	goModTidy(projectDir)
	sg := service.NewServiceGroup()
	sg.Add(serviceImpl{func() {
		log.Debug(">> Ready to start a zRPC server...")
		goStart(zrpcWorkDir)
	}})
	sg.Add(serviceImpl{func() {
		mono.start()
	}})
	sg.Start()
}
