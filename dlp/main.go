// Copyright 2018 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

// dlp is an example of using the DLP API.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	dlp "cloud.google.com/go/dlp/apiv2beta1"
	dlppb "google.golang.org/genproto/googleapis/privacy/dlp/v2beta1"
)

func inspect(w io.Writer, client *dlp.Client, s string) {
	rcr := &dlppb.InspectContentRequest{
		InspectConfig: &dlppb.InspectConfig{
			InfoTypes: []*dlppb.InfoType{
				{
					Name: "US_SOCIAL_SECURITY_NUMBER",
				},
			},
			MinLikelihood: dlppb.Likelihood_LIKELIHOOD_UNSPECIFIED,
		},
		Items: []*dlppb.ContentItem{
			{
				Type: "text/plain",
				DataItem: &dlppb.ContentItem_Data{
					Data: []byte(s),
				},
			},
		},
	}
	r, err := client.InspectContent(context.Background(), rcr)
	if err != nil {
		log.Fatal(err)
	}
	fs := r.GetResults()[0].GetFindings()
	for _, f := range fs {
		fmt.Fprintf(w, "%s\n", f.GetInfoType().GetName())
	}
}

func redact(w io.Writer, client *dlp.Client, s string) {
	rcr := &dlppb.RedactContentRequest{
		InspectConfig: &dlppb.InspectConfig{
			InfoTypes: []*dlppb.InfoType{
				{
					Name: "US_SOCIAL_SECURITY_NUMBER",
				},
			},
			MinLikelihood: dlppb.Likelihood_LIKELIHOOD_UNSPECIFIED,
		},
		ReplaceConfigs: []*dlppb.RedactContentRequest_ReplaceConfig{
			{
				InfoType:    &dlppb.InfoType{Name: "US_SOCIAL_SECURITY_NUMBER"},
				ReplaceWith: "[redacted]",
			},
		},
		Items: []*dlppb.ContentItem{
			{
				Type: "text/plain",
				DataItem: &dlppb.ContentItem_Data{
					Data: []byte(s),
				},
			},
		},
	}
	r, err := client.RedactContent(context.Background(), rcr)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(w, "%s\n", r.GetItems()[0].GetData())
}

func infoTypes(w io.Writer, client *dlp.Client, s string) {
	rcr := &dlppb.ListInfoTypesRequest{
		Category: s,
	}
	r, err := client.ListInfoTypes(context.Background(), rcr)
	if err != nil {
		log.Fatal(err)
	}
	for _, it := range r.GetInfoTypes() {
		fmt.Fprintf(w, "%s\n", it.GetName())
	}
}

func categories(w io.Writer, client *dlp.Client) {
	rcr := &dlppb.ListRootCategoriesRequest{}
	r, err := client.ListRootCategories(context.Background(), rcr)
	if err != nil {
		log.Fatal(err)
	}
	for _, c := range r.GetCategories() {
		fmt.Fprintf(w, "%s (%s)\n", c.GetName(), c.GetDisplayName())
	}
}

func mask(w io.Writer, client *dlp.Client, s string) {
	rcr := &dlppb.DeidentifyContentRequest{
		DeidentifyConfig: &dlppb.DeidentifyConfig{
			Transformation: &dlppb.DeidentifyConfig_InfoTypeTransformations{
				InfoTypeTransformations: &dlppb.InfoTypeTransformations{
					Transformations: []*dlppb.InfoTypeTransformations_InfoTypeTransformation{
						{
							InfoTypes: []*dlppb.InfoType{},
							PrimitiveTransformation: &dlppb.PrimitiveTransformation{
								Transformation: &dlppb.PrimitiveTransformation_CharacterMaskConfig{
									CharacterMaskConfig: &dlppb.CharacterMaskConfig{
										MaskingCharacter: "*",
									},
								},
							},
						},
					},
				},
			},
		},
		Items: []*dlppb.ContentItem{
			{
				Type: "text/plain",
				DataItem: &dlppb.ContentItem_Data{
					Data: []byte(s),
				},
			},
		},
	}
	r, err := client.DeidentifyContent(context.Background(), rcr)
	if err != nil {
		log.Fatal(err)
	}
	for _, i := range r.GetItems() {
		fmt.Fprintf(w, "%s\n", i.GetData())
	}
}

func deidentifyFPE(w io.Writer, client *dlp.Client, s, wrappedKey, cryptoKeyName string) {
	rcr := &dlppb.DeidentifyContentRequest{
		DeidentifyConfig: &dlppb.DeidentifyConfig{
			Transformation: &dlppb.DeidentifyConfig_InfoTypeTransformations{
				InfoTypeTransformations: &dlppb.InfoTypeTransformations{
					Transformations: []*dlppb.InfoTypeTransformations_InfoTypeTransformation{
						{
							InfoTypes: []*dlppb.InfoType{},
							PrimitiveTransformation: &dlppb.PrimitiveTransformation{
								Transformation: &dlppb.PrimitiveTransformation_CryptoReplaceFfxFpeConfig{
									CryptoReplaceFfxFpeConfig: &dlppb.CryptoReplaceFfxFpeConfig{
										CryptoKey: &dlppb.CryptoKey{
											Source: &dlppb.CryptoKey_KmsWrapped{
												KmsWrapped: &dlppb.KmsWrappedCryptoKey{
													WrappedKey:    []byte(wrappedKey),
													CryptoKeyName: cryptoKeyName,
												},
											},
										},
										Alphabet: &dlppb.CryptoReplaceFfxFpeConfig_CommonAlphabet{
											CommonAlphabet: dlppb.CryptoReplaceFfxFpeConfig_ALPHA_NUMERIC,
										},
									},
								},
							},
						},
					},
				},
			},
		},
		Items: []*dlppb.ContentItem{
			{
				Type: "text/plain",
				DataItem: &dlppb.ContentItem_Data{
					Data: []byte(s),
				},
			},
		},
	}
	r, err := client.DeidentifyContent(context.Background(), rcr)
	if err != nil {
		log.Fatal(err)
	}
	for _, i := range r.GetItems() {
		fmt.Fprintf(w, "%s\n", i.GetData())
	}
}

func main() {
	ctx := context.Background()
	client, err := dlp.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	flag.Parse()

	switch flag.Arg(0) {
	case "inspect":
		inspect(os.Stdout, client, flag.Arg(1))
	case "redact":
		redact(os.Stdout, client, flag.Arg(1))
	case "infoTypes":
		infoTypes(os.Stdout, client, flag.Arg(1))
	case "categories":
		categories(os.Stdout, client)
	case "mask":
		mask(os.Stdout, client, flag.Arg(1))
	case "deidfpe":
		deidentifyFPE(os.Stdout, client, flag.Arg(1), flag.Arg(2), flag.Arg(3))
	default:
		fmt.Fprintf(os.Stderr, `Usage: %s CMD "string"\n`, os.Args[0])
		os.Exit(1)
	}
}