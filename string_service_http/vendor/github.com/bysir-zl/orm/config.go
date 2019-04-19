package orm

import (
	"errors"
	"fmt"
)

type Config map[string]Connect

var config = Config{}

func (p *Config) writeConnect(connect string) (conn *Connect, err error) {
	m := map[string]Connect(*p)
	if c, ok := m[connect + "-write"]; ok {
		conn = &c
		return
	}
	if c, ok := m[connect]; ok {
		conn = &c
		return
	}
	err = errors.New("can't found connect: " + connect)
	return
}

func (p *Config) readConnect(connect string) (conn *Connect, err error) {
	m := map[string]Connect(*p)
	if c, ok := m[connect + "-read"]; ok {
		conn = &c
		return
	}
	if c, ok := m[connect]; ok {
		conn = &c
		return
	}
	err = errors.New("can't found connect: " + connect)
	return
}

type Connect struct {
	Driver string `json:"driver"`
	// USER:PWD@tcp(HOST:PORT)/DBNAME
	Url string `json:"url"`
}

func (p *Connect) String() string {
	return fmt.Sprintf("%s~%s", p.Url, p.Driver)
}

func (p *Connect) SqlString() string {
	return p.Url
}
