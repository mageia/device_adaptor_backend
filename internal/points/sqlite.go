package points

import (
	"database/sql/driver"
	"errors"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/json-iterator/go"
	"github.com/rs/zerolog/log"
	"math"
	"os"
	"path"
	"runtime"
	"strings"
	"time"
)

var SqliteDB *gorm.DB

type HashStringType map[string]string
type HashMapType map[string]interface{}
type ArrayStringType []string
type PointType uint8

type PointDefine struct {
	ID        uint            `gorm:"primary_key" json:"-"`
	CreatedAt time.Time       `json:"-"`
	UpdatedAt time.Time       `json:"-"`
	InputName string          `json:"-" gorm:"not null"`
	PointKey  string          `json:"-" gorm:"not null"` //纯ASCII码命名，一般作为点表主key
	Name      string          `json:"name" yaml:"name"`  //命名任意，唯一即可，类似于short description
	Label     string          `json:"label,omitempty" yaml:"label"`
	Unit      string          `json:"unit,omitempty" yaml:"unit"`
	Address   string          `json:"address" yaml:"address"`
	PointType PointType       `json:"point_type" yaml:"point_type"`
	Parameter float64         `json:"parameter,omitempty" yaml:"parameter"`
	Option    HashStringType  `json:"option,omitempty" yaml:"option" gorm:"type:text"`
	Control   HashStringType  `json:"control,omitempty" yaml:"control" gorm:"type:text"`
	Tags      ArrayStringType `json:"tags,omitempty" yaml:"tags" gorm:"type:text"`
	Extra     HashMapType     `json:"extra,omitempty" yaml:"extra" gorm:"type:text"` //网关在存储extra时可能会序列化，使用时请注意
}

const (
	PointAnalog  PointType = iota
	PointDigital
	PointInteger
	PointString
	PointUnknown = math.MaxUint8
)

func (hs *HashStringType) Scan(val interface{}) error {
	switch val := val.(type) {
	case string:
		return jsoniter.Unmarshal([]byte(val), hs)
	case []byte:
		return jsoniter.Unmarshal(val, hs)
	default:
		return errors.New("not support")
	}
	return nil
}
func (hs HashStringType) Value() (driver.Value, error) {
	bytes, err := jsoniter.Marshal(hs)
	return string(bytes), err
}

func (as *ArrayStringType) Scan(val interface{}) error {
	switch val := val.(type) {
	case string:
		*as = ArrayStringType(strings.Split(val, ","))
		return nil
	case []byte:
		*as = ArrayStringType(strings.Split(string(val), ","))
		return nil
	default:
		return errors.New("not support")
	}
	return nil
}
func (as ArrayStringType) Value() (driver.Value, error) {
	return []byte(strings.Join(as, ",")), nil
}

func (hs *HashMapType) Scan(val interface{}) error {
	switch val := val.(type) {
	case string:
		return jsoniter.Unmarshal([]byte(val), hs)
	case []byte:
		return jsoniter.Unmarshal(val, hs)
	default:
		return errors.New("not support")
	}
	return nil
}
func (hs HashMapType) Value() (driver.Value, error) {
	return jsoniter.Marshal(hs)
}

func init() {
	var err error
	dbPath := "point_map.db"
	
	if runtime.GOOS == "linux" {
		if runtime.GOARCH == "arm" {
			dbPath = "./"
		} else {
			dbPath = "/var/device_adaptor/"
		}
		if _, e := os.Stat(dbPath); e != nil {
			if os.IsNotExist(err) {
				os.Mkdir(dbPath, 0777)
			}
		}
		dbPath = path.Join(dbPath, "point_map.db")
	}

	SqliteDB, err = gorm.Open("sqlite3", dbPath)
	if err != nil {
		log.Error().Err(err).Msg("Connect database")
	}

	SqliteDB.LogMode(false)
	SqliteDB.AutoMigrate(&PointDefine{})
}
