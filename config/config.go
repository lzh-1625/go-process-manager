package config

import (
	"log"
	"os"
	"reflect"
	"strconv"

	"github.com/lzh-1625/go_process_manager/utils"
	"go.yaml.in/yaml/v2"
)

var CF = new(Configuration)

func init() {
	err := LoadConfig()
	if err == nil {
		log.Println("config loaded from file")
		return
	}
	if err := ResetConfig(); err != nil {
		log.Println("load config default value failed, reset config failed")
		panic(err)
	}
	if err := DumpConfig(); err != nil {
		panic(err)
	}
	log.Println("config initialized successfully")
}

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
	SecretKey                 string  `default:"-"`
}

func LoadConfig() error {
	f, err := os.Open("config.yaml")
	if err != nil {
		return err
	}
	defer f.Close()
	err = yaml.NewDecoder(f).Decode(CF)
	if err != nil {
		return err
	}
	return nil
}

func DumpConfig() error {
	f, err := os.Create("config.yaml")
	if err != nil {
		return err
	}
	defer f.Close()
	err = yaml.NewEncoder(f).Encode(CF)
	if err != nil {
		return err
	}
	return nil
}

func ResetConfig() error {
	typeElem := reflect.TypeFor[Configuration]()
	valueElem := reflect.ValueOf(CF).Elem()
	for i := 0; i < typeElem.NumField(); i++ {
		typeField := typeElem.Field(i)
		valueField := valueElem.Field(i)
		var err error
		defaultValue := typeField.Tag.Get("default")
		if defaultValue == "-" {
			continue
		}
		switch typeField.Type.Kind() {
		case reflect.String:
			valueField.SetString(defaultValue)
		case reflect.Bool:
			value, errV := strconv.ParseBool(defaultValue)
			err = errV
			if err == nil {
				valueField.SetBool(value)
			}
		case reflect.Float64:
			value, errV := strconv.ParseFloat(defaultValue, 64)
			err = errV
			if err == nil {
				valueField.SetFloat(value)
			}
		case reflect.Int64, reflect.Int:
			value, errV := strconv.ParseInt(defaultValue, 10, 64)
			err = errV
			if err == nil {
				valueField.SetInt(value)
			}
		default:
			continue
		}
		if err != nil {
			return err
		}
	}
	CF.SecretKey = utils.RandString(32)
	return nil
}
