package config

var CF = new(Configuration)

// 只支持 float64、int、int64、bool、string类型
type Configuration struct {
	LogLevel                  string  `default:"debug"  describe:"log level [debug,info]"`
	Listen                    string  `default:":8797" describe:"listen port"`
	StorgeType                string  `default:"sqlite" describe:"storage engine [sqlite,es,bleve]"`
	EsUrl                     string  `default:"" describe:"Elasticsearch url"`
	EsIndex                   string  `default:"server_log_v1" describe:"Elasticsearch index"`
	EsUsername                string  `default:"" describe:"Elasticsearch username"`
	EsPassword                string  `default:"" describe:"Elasticsearch password"`
	EsWindowLimit             bool    `default:"true" describe:"Es pagination 10000 limit"`
	FileSizeLimit             float64 `default:"10.0" describe:"file size limit (MB)"`
	ProcessRestartsLimit      int     `default:"2" describe:"process restart limit"`
	ProcessMsgCacheLinesLimit int     `default:"50" describe:"std process cache message lines limit"`
	ProcessMsgCacheBufLimit   int     `default:"4096" describe:"pty process cache message bytes limit"`
	ProcessExpireTime         int64   `default:"60" describe:"process control timeout (seconds)"`
	PerformanceInfoListLength int     `default:"30" describe:"performance info storage length"`
	PerformanceInfoInterval   int     `default:"60" describe:"monitor interval time (seconds)"`
	TerminalConnectTimeout    int     `default:"10" describe:"terminal connect timeout (minutes)"`
	UserPassWordMinLength     int     `default:"4" describe:"user password min length"`
	LogMinLenth               int     `default:"0" describe:"filter log min length"`
	LogHandlerPoolSize        int     `default:"10" describe:"log handler parallel number"`
	PprofEnable               bool    `default:"true" describe:"enable pprof analysis tool"`
	KillWaitTime              int     `default:"5" describe:"kill signal wait time (seconds)"`
	TaskTimeout               int     `default:"60" describe:"task execution timeout (seconds)"`
	TokenExpirationTime       int64   `default:"720" describe:"token expiration time (hours)"`
	WsHealthCheckInterval     int     `default:"3" describe:"ws health check interval (seconds)"`
	CgroupPeriod              int64   `default:"100000" describe:"CgroupPeriod"`
	CgroupSwapLimit           bool    `default:"false" describe:"cgroup swap limit"`
	CondWaitTime              int     `default:"30" describe:"long polling wait time (seconds)"`
	EventStorageTime          int     `default:"1" describe:"event storage time (days)"`
	GZipEnable                bool    `default:"false" describe:"enable gzip compression"`
	Tui                       bool    `default:"-"`
}
