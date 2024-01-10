package bind_etcd_cfg

type ILogger interface {
	Infof(template string, args ...interface{})
	Warnf(template string, args ...interface{})
	Errorf(template string, args ...interface{})
}

var defaultLogger ILogger = &dummyLogger{}

// SetLogger ...
func SetLogger(l ILogger) {
	defaultLogger = l
}

type dummyLogger struct{}

func (d *dummyLogger) Infof(template string, args ...interface{}) {}

func (d *dummyLogger) Warnf(template string, args ...interface{}) {}

func (d *dummyLogger) Errorf(template string, args ...interface{}) {}
