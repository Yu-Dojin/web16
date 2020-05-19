package main

import (
	"log"
	"net/http"

	"github.com/Yu-Dojin/web16/app"
)

func main() {
	m := app.MakeHandler("./test.db") //코드에 밖혀있더라도 이와같이 패키지 안쪽에 있는것보다 최대한 바깥쪽에 있는것이 좋다. 또는 flag 패키지를 이용해 Arg 와 같은 실행이자를 가져오게 할 수 있다.
	defer m.Close()

	log.Println("Started App")
	err := http.ListenAndServe(":3000", m)
	if err != nil {
		panic(err)
	}
}
