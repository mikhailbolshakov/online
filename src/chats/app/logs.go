package app

import (
	"encoding/json"
	"log"
)

const (
	llDebug = "DEBUG"
	llInfo  = "INFO"
	llError = "ERROR"
)

type LogHandler struct {
	Level string
}

func initLogs() *LogHandler {
	return &LogHandler{ Level: llDebug }
}

//func (log *LogHandler) Scan() {
//	reader := bufio.NewReader(os.Stdin)
//	for {
//		text, err := reader.ReadString('\n')
//		if err != nil {
//			break
//		}
//		text = strings.Replace(text, "\n", "", -1)
//
//		if strings.Compare("api", text) == 0 {
//			fmt.Print("this!") //	TODO
//			log.Print = !log.Print
//		}
//	}
//}

func (l *LogHandler) Info(m...interface{}){
	if l.Level == llInfo || l.Level == llDebug {
		log.Println(m)
	}
}

func (l *LogHandler) Infof(m string, a...interface{}){
	if l.Level == llInfo || l.Level == llDebug {
		log.Printf(m + "\n", a)
	}
}

func (l *LogHandler) Debug(m...interface{}){
	if l.Level == llDebug {
		log.Println(m)
	}
}

func (l *LogHandler) Debugf(m string, a...interface{}){

	var params []interface{}

	for _, i := range a {
		s, _ := json.MarshalIndent(i, "", "\t")
		params = append(params, s)
	}

	if l.Level == llDebug {
		log.Printf(m + "\n", params)
	}
}

func (l *LogHandler) Error(m...interface{}){
	log.Println(m)
}

func (l *LogHandler) Errorf(m string, a...interface{}){
	log.Printf(m, a)
}
