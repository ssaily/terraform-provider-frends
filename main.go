package main

import (
	"context"
	"flag"
	"log"

	"github.com/frends/terraform-provider-frends/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

// Version is set during build via ldflags.
var Version = "dev"

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/frends/frends",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), provider.New(Version), opts)
	if err != nil {
		log.Fatal(err.Error())
	}
}
