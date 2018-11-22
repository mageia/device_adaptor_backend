package points

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"log"
	"math"
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
	PointKey  string          `json:"point_key" yaml:"point_key" gorm:"not null"`
	Name      string          `json:"name" yaml:"name"`
	Label     string          `json:"label" yaml:"label"`
	Unit      string          `json:"unit" yaml:"unit"`
	Address   string          `json:"address" yaml:"address"`
	PointType PointType       `json:"point_type" yaml:"point_type"`
	Parameter float64         `json:"parameter,omitempty" yaml:"parameter"`
	Option    HashStringType  `json:"option,omitempty" yaml:"option" gorm:"type:text"`
	Control   HashStringType  `json:"control,omitempty" yaml:"control" gorm:"type:text"`
	Tags      ArrayStringType `json:"tags,omitempty" yaml:"tags" gorm:"type:text"`
	Extra     HashMapType     `json:"extra,omitempty" yaml:"extra" gorm:"type:text"`
}

const (
	_ PointType = iota
	PointAnalog
	PointDigital
	PointString
	PointUnknown = math.MaxUint8
)

func (hs *HashStringType) Scan(val interface{}) error {
	switch val := val.(type) {
	case string:
		return json.Unmarshal([]byte(val), hs)
	case []byte:
		return json.Unmarshal(val, hs)
	default:
		return errors.New("not support")
	}
	return nil
}
func (hs HashStringType) Value() (driver.Value, error) {
	bytes, err := json.Marshal(hs)
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
		return json.Unmarshal([]byte(val), hs)
	case []byte:
		return json.Unmarshal(val, hs)
	default:
		return errors.New("not support")
	}
	return nil
}
func (hs HashMapType) Value() (driver.Value, error) {
	bytes, err := json.Marshal(hs)
	return string(bytes), err
}

func init() {
	var err error
	dbPath := "point_map.db"
	if runtime.GOOS == "linux" {
		dbPath = "/var/run/deviceAdaptor/point_map.db"
	}

	SqliteDB, err = gorm.Open("sqlite3", dbPath)
	if err != nil {
		log.Printf("failed to connect database: %v", err)
		//panic("failed to connect database")
	}

	SqliteDB.LogMode(true)
	SqliteDB.AutoMigrate(&PointDefine{})
}
