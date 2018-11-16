package points

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"log"
	"math"
	"strings"
)

var SqliteDB *gorm.DB

type HashStringType map[string]string
type HashMapType map[string]interface{}
type ArrayStringType []string
type PointType uint8

type PointDefine struct {
	gorm.Model `json:"-"`
	InputName  string          `json:"-" gorm:"not null"`
	Name       string          `json:"name" yaml:"name"`
	Label      string          `json:"label" yaml:"label"`
	Unit       string          `json:"unit" yaml:"unit"`
	Address    string          `json:"address" yaml:"address"`
	PointType  PointType       `json:"point_type" yaml:"point_type"`
	Parameter  float64         `json:"parameter,omitempty" yaml:"parameter"`
	Option     HashStringType  `json:"option,omitempty" yaml:"option" gorm:"type:text,default:'{}'"`
	Control    HashStringType  `json:"control,omitempty" yaml:"control" gorm:"type:text,default:'{}'"`
	Tags       ArrayStringType `json:"tags,omitempty" yaml:"tags" gorm:"type:text,default:''"`
	Extra      HashMapType     `json:"extra,omitempty" yaml:"extra" gorm:"type:text,default:'{}'"`
}

const (
	_ PointType = iota
	PointAnalog
	PointDigital
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
	SqliteDB, err = gorm.Open("sqlite3", "point_map.db")
	if err != nil {
		log.Println(err)
		panic("failed to connect database")
	}

	//SqliteDB.LogMode(true)
	SqliteDB.AutoMigrate(&PointDefine{})
}
