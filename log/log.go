package log

import (
	"git.cloudbro.net/michaelfigg/yallawebsites/log/hooks"
	"github.com/sirupsen/logrus"
	"os"
)

var defLogger *logrus.Logger

func SetupLog() {
	defLogger = logrus.New()
	defLogger.Out = os.Stdout
	defLogger.AddHook(hooks.NewHook())
}

func Log() *logrus.Logger {
	return defLogger
}
