// Package main provides the Terraform provider entry point for Braintrust.
package main

import (
	"context"
	"flag"
	"log"

	"github.com/braintrustdata/terraform-provider-braintrustdata/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

// version will be set by the goreleaser build process
var version = "dev"

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/braintrustdata/braintrustdata",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), provider.New(version), opts)

	if err != nil {
		log.Fatal(err.Error())
	}
}
