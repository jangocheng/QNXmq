package config

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// CommentPrefix string
var commentPrefix = []string{"//", "#", ";", "/**", " * ", "##"}

type Configs struct {
	config map[string]interface{}
	node   string
}

var Conf *Configs

func SrvInit(LoadConfig string) {
	Conf = new(Configs)
	Conf.LoadConfig(LoadConfig)
}

func checkComment(line string) bool {
	line = strings.TrimSpace(line)
	for p := range commentPrefix {
		if strings.HasPrefix(line, commentPrefix[p]) {
			return true
		}
	}
	return false
}

func (conf *Configs) LoadConfig(path string) {
	conf.config = make(map[string]interface{})
	file, err := os.Open(path)
	scanner := bufio.NewScanner(file)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	buf := bufio.NewReader(file)
	for {
		lines, _, err := buf.ReadLine()
		line := strings.TrimSpace(string(lines))
		if err != nil {
			//last line read
			if err == io.EOF {
				break
			}
			fmt.Println(err)
			return
		}
		// checkComment
		if checkComment(scanner.Text()) {
			continue
		}
		//if is [xxx] commit
		n := strings.Index(line, "[")
		nl := strings.LastIndex(line, "]")

		if n > -1 && nl > -1 && nl > n+1 {
			conf.node = strings.TrimSpace(line[n+1 : nl])
			continue
		}
		if len(conf.node) == 0 || len(line) == 0 {
			continue
		}
		arr := strings.Split(line, "=")
		key := strings.TrimSpace(arr[0])
		value := strings.TrimSpace(arr[1])
		newKey := conf.node + key
		conf.config[newKey] = value
	}
}

type block struct {
	data interface{}
}

func mapjson(file map[string]interface{}) (*block, error) {
	if file == nil {
		return nil, errors.New("get config fail, config content is nil")
	}
	return &block{
		data: file,
	}, nil
}

//Get function returns the value of string key
func (conf *Configs) Get(node, key string) (value interface{}) {
	key = node + key
	if v, ok := conf.config[key]; !ok {
		return ""
	} else {
		return v
	}
}

//Getint function returns the value of int key
func (conf *Configs) Getint(node, key string) int {
	key = node + key
	value, _ := conf.config[key]
	if len(value.(string)) == 0 {
		return 0
	}
	i, err := strconv.Atoi(value.(string))
	if err != nil {
		return 0
	}
	return i
}

//GetDouble function returns the value of a float64
func (conf *Configs) GetDouble(node, key string) float64 {
	key = node + key
	value, _ := conf.config[key]
	if len(value.(string)) == 0 {
		return 0
	}
	i, err := strconv.ParseFloat(value.(string), 64)
	if err != nil {
		return 0
	}
	return i
}
