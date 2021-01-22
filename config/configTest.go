package config

import (
	"fmt"
)

func main(){
	cfg := ParseProfile()
	fmt.Printf("cfg ready, version=%s", cfg.Version)
}
