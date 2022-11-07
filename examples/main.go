package main

import (
	"fmt"
	"github.com/smartwalle/nmail"
)

func main() {
	var client = nmail.NewClient(
		"smartwalle@gmail.com",
		"your password",
		"smtp.gmail.com",
		"587",
		nmail.WithMaxIdle(1),
		nmail.WithMaxActive(1),
	)

	var m = nmail.NewHTMLMessage("Title", "<a href='http://www.google.com'>Hello Google</a>")
	m.From = "Yang<webreservation@hoteldelins.com>"
	m.To = []string{"917996695@qq.com"}

	for i := 0; i < 10; i++ {
		go func(i int) {
			fmt.Println(i, client.Send(m))
		}(i)
	}

	select {}
}
