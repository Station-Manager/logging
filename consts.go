package logging

import "github.com/Station-Manager/types"

const (
	ServiceName = types.LoggerServiceName //"logger"
	emptyString = ""
)

const (
	errMsgNilService       = "Logger service is nil."
	errMsgWorkingDirNotSet = "Working directory is not set."
	errMsgAppCfgNotSet     = "Application config is not set."
)
