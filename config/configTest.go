package config

import (
	"fmt"
)

func TestReadProfile(){
	cfg := ParseProfile()
	fmt.Printf("cfg ready, version=%s", cfg.Version)
}
