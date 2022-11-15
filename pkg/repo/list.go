package repo

import (
	"fmt"

	"github.com/jatalocks/opsilon/internal/config"
)

func List() {
	fmt.Println(config.GetConfig())
}
