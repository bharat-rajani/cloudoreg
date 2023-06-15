package main

import (
	"context"
	"github.com/bharat-rajani/cloudoreg/internal/cloudoreg"
)

func main() {

	ctx := context.Background()
	cloudoregApp, err := cloudoreg.NewCloudoreg(ctx, cloudoreg.DefaultCloudoregConfig())
	if err != nil {
		panic(err)
	}
	cloudoregApp.Run(ctx)
}
