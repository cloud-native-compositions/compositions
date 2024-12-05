// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package crds

// TODO: barney-s: migrate use of apiextensions to extv1

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	"github.com/gobuffalo/flect"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsvalidation "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/validation"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	FacadeGroup = "facade.compositions.google.com"
)

var scheme = runtime.NewScheme()

func init() {
	if err := extv1.AddToScheme(scheme); err != nil {
		panic(err)
	}
	if err := apiextensions.AddToScheme(scheme); err != nil {
		panic(err)
	}
}

type CRDInfo struct {
	GVK            schema.GroupVersionKind
	Plural         string
	ShortNames     []string
	Categories     []string
	PrinterColumns []apiextensions.CustomResourceColumnDefinition
	Labels         map[string]string
	schema         *apiextensions.JSONSchemaProps
}

func NewFacadeCRDInfo(
	gvk schema.GroupVersionKind,
	plural string,
	shortNames []string,
	printerCols []apiextensions.CustomResourceColumnDefinition,
	labels map[string]string) *CRDInfo {

	if gvk.Group == "" {
		gvk.Group = FacadeGroup
	}
	if gvk.Version == "" {
		gvk.Version = "v1"
	}
	if plural == "" {
		plural = strings.ToLower(flect.Pluralize(gvk.Kind))
	}
	crd := CRDInfo{
		GVK:            gvk,
		Plural:         plural,
		ShortNames:     shortNames,
		Categories:     []string{"facade", "facades"},
		PrinterColumns: printerCols,
		schema:         nil,
	}

	crd.Labels = map[string]string{
		"compositions.google.com/facade": "yes",
	}
	for k, v := range labels {
		crd.Labels[k] = v
	}
	return &crd
}

func (c *CRDInfo) SetCRDSchema(schema *apiextensions.JSONSchemaProps) {
	c.schema = schema
}

func (c *CRDInfo) Name() string {
	return c.Plural + "." + c.GVK.Group
}

func (c *CRDInfo) String() string {
	return c.Name()
}

func (c *CRDInfo) SetSpec(specProperties *extv1.JSONSchemaProps) error {
	// Do we need this ?
	// Can we use extv1 only instead
	specUnversionedProperties := &apiextensions.JSONSchemaProps{}
	// Risk ? nil conversion.Scope passed
	if err := extv1.Convert_v1_JSONSchemaProps_To_apiextensions_JSONSchemaProps(
		specProperties, specUnversionedProperties, nil); err != nil {
		return err
	}
	statusProperties := map[string]apiextensions.JSONSchemaProps{
		"conditions": {
			Type: "array",
			Items: &apiextensions.JSONSchemaPropsOrArray{
				Schema: &apiextensions.JSONSchemaProps{
					Description: "",
					Required:    []string{"lastTransitionTime", "message", "reason", "status", "type"},
					Type:        "object",
					Properties: map[string]apiextensions.JSONSchemaProps{
						"lastTransitionTime": {
							Description: "",
							Format:      "date-time",
							Type:        "string",
						},
						"message": {
							Description: "human readable message",
							MaxLength:   ptr.To[int64](1024),
							Type:        "string",
						},
						"observedGeneration": {
							Description: "",
							Format:      "int64",
							Minimum:     ptr.To[float64](0),
							Type:        "integer",
						},
						"reason": {
							Description: "",
							MaxLength:   ptr.To[int64](256),
							MinLength:   ptr.To[int64](1),
							Pattern:     "^[A-Za-z]([A-Za-z0-9_,:]*[A-Za-z0-9_])?$",
							Type:        "string",
						},
						"status": {
							Description: "status of the condition, one of True, False, Unknown.",
							Enum:        []apiextensions.JSON{"True", "False", "Unknown"},
							Type:        "string",
						},
						"type": {
							Description: "type of condition in CamelCase or in foo.example.com/CamelCase." +
								" The regex it matches is (dns1123SubdomainFmt/)?(qualifiedNameFmt)",
							MaxLength: ptr.To[int64](316),
							Pattern: "^([a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*/)" +
								"?(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])$",
							Type: "string",
						},
					},
				},
			},
		},
	}

	crdSchema := &apiextensions.JSONSchemaProps{
		Type:        "object",
		Description: "TODO",
		Properties: map[string]apiextensions.JSONSchemaProps{
			"apiVersion": {Type: "string"},
			"kind":       {Type: "string"},
			"metadata":   {Type: "object"},
			"spec":       *specUnversionedProperties.DeepCopy(),
			"status": {
				Type:                   "object",
				Properties:             statusProperties,
				XPreserveUnknownFields: ptr.To[bool](true),
			},
		},
	}

	c.schema = crdSchema
	return nil
}

// CRD takes a schema and converts it to a CRD.
func (c *CRDInfo) CRD() (*apiextensions.CustomResourceDefinition, error) {
	if c.schema == nil {
		return nil, fmt.Errorf("Schema is nil. Use SetSpec or SetCRDSchema first.")
	}
	crd := &apiextensions.CustomResourceDefinition{
		Spec: apiextensions.CustomResourceDefinitionSpec{
			PreserveUnknownFields: ptr.To[bool](false),
			Group:                 c.GVK.Group,
			Names: apiextensions.CustomResourceDefinitionNames{
				Kind:       c.GVK.Kind,
				ListKind:   c.GVK.Kind + "List",
				Plural:     strings.ToLower(c.Plural),
				Singular:   strings.ToLower(c.GVK.Kind),
				ShortNames: c.ShortNames,
				Categories: c.Categories,
			},
			Validation: &apiextensions.CustomResourceValidation{
				OpenAPIV3Schema: c.schema,
			},
			Scope:   apiextensions.NamespaceScoped,
			Version: c.GVK.Version,
			Subresources: &apiextensions.CustomResourceSubresources{
				Status: &apiextensions.CustomResourceSubresourceStatus{},
				Scale:  nil,
			},
			Versions: []apiextensions.CustomResourceDefinitionVersion{
				{
					Name:    c.GVK.Version,
					Storage: true,
					Served:  true,
				},
			},
			AdditionalPrinterColumns: c.PrinterColumns,
		},
	}

	// Defaulting functions are not found in versionless CRD package
	crdv1 := &extv1.CustomResourceDefinition{}
	if err := scheme.Convert(crd, crdv1, nil); err != nil {
		return nil, err
	}
	scheme.Default(crdv1)

	crd2 := &apiextensions.CustomResourceDefinition{}
	if err := scheme.Convert(crdv1, crd2, nil); err != nil {
		return nil, err
	}
	crd2.ObjectMeta.Name = c.Name()

	labels := c.Labels
	if labels == nil {
		labels = make(map[string]string)
	}
	crd2.ObjectMeta.Labels = labels

	return crd2, nil
}

func (c *CRDInfo) InstallCRD(ctx context.Context,
	logger logr.Logger,
	cc client.Client,
	rs *runtime.Scheme,
) error {

	var crd extv1.CustomResourceDefinition
	crdName := c.Name()

	logger.Info("Checking if CRD exists", "crd", crdName)
	err := cc.Get(ctx, types.NamespacedName{Name: crdName, Namespace: ""}, &crd)
	// CRD exists. Nothing to be done.
	if err == nil {
		logger.Info("CRD exists. Not creating.", "crd", crdName)
		return nil
	}

	// If we are unable to get it for some reason other than not found return
	if !apierrors.IsNotFound(err) {
		logger.Error(err, "failed to get an Facade CRD object")
		return err
	}

	// Construct Facade CRD from the openAPI Schema
	unversionedFacadeCRD, err := c.CRD()
	if err == nil {
		facadeCRD := &extv1.CustomResourceDefinition{}
		err = rs.Convert(unversionedFacadeCRD, facadeCRD, nil)
		if err == nil {
			err = cc.Create(ctx, facadeCRD)
			if err != nil {
				logger.Error(err, "failed to Create Facade CRD")
			}
		} else {
			logger.Error(err, "CRD conversion error")
		}
	} else {
		logger.Error(err, "Error getting unversioned CRD")
	}

	logger.Info("Created Facade CRD", "crd", crdName)
	return err
}

// ValidateCRD calls the CRD package's validation on an internal representation of the CRD.
func ValidateCRD(ctx context.Context, crd *apiextensions.CustomResourceDefinition) error {
	errs := apiextensionsvalidation.ValidateCustomResourceDefinition(ctx, crd)
	if len(errs) > 0 {
		return errs.ToAggregate()
	}
	return nil
}
