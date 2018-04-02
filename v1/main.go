package main

import (
"os"

	"fmt"
)

func main() {
	configPath := os.Args[1]
	fmt.Printf("Loading config from : %s \n",configPath )

	ad := NewMifloraAd(configPath)
	ad.Start()
	select{}

}


//func readConfig() {
//
//}


