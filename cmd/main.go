package main

import (
	"context"
	"fmt"
	"github.com/bharat-rajani/cloudoreg/internal/cloudoreg"
)

func main() {

	fmt.Println(`
 ██████╗██╗      ██████╗ ██╗   ██╗██████╗  ██████╗ ██████╗ ███████╗ ██████╗ 
██╔════╝██║     ██╔═══██╗██║   ██║██╔══██╗██╔═══██╗██╔══██╗██╔════╝██╔════╝ 
██║     ██║     ██║   ██║██║   ██║██║  ██║██║   ██║██████╔╝█████╗  ██║  ███╗
██║     ██║     ██║   ██║██║   ██║██║  ██║██║   ██║██╔══██╗██╔══╝  ██║   ██║
╚██████╗███████╗╚██████╔╝╚██████╔╝██████╔╝╚██████╔╝██║  ██║███████╗╚██████╔╝
 ╚═════╝╚══════╝ ╚═════╝  ╚═════╝ ╚═════╝  ╚═════╝ ╚═╝  ╚═╝╚══════╝ ╚═════╝ 
                                                                            
`)
	ctx := context.Background()
	cloudoregApp, err := cloudoreg.NewCloudoreg(ctx, cloudoreg.DefaultCloudoregConfig())
	if err != nil {
		panic(err)
	}
	cloudoregApp.Run(ctx)
}
