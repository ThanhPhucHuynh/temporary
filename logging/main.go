package main

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/tphuc/go-logging/glog"
)

func main() {
	hook, err := glog.NewKafkaHook(
		"kh",
		[]logrus.Level{logrus.InfoLevel, logrus.WarnLevel, logrus.ErrorLevel},
		&logrus.JSONFormatter{},
		[]string{"localhost:9092"},
	)
	if err != nil {
		fmt.Println("dhasgj: ", err)
		return
	}
	l := glog.NewWithKafkaHook(hook)
	ll := l.WithField("topics", []string{"quickstart"})
	ll.Infof("hello s my log")
	ll.Errorf("hello s my log")
	for i := 0; i < 100; i++ {
		go func(i int) {
			ll.Infof("xhello dsahkdjs my log %d", i)
		}(i)
	}
	time.Sleep(time.Second * 9)

}
