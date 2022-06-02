package main

import (
	_ "github.com/shalb/cluster.dev/pkg/backend/azurerm"
	_ "github.com/shalb/cluster.dev/pkg/backend/do"
	_ "github.com/shalb/cluster.dev/pkg/backend/gcs"
	_ "github.com/shalb/cluster.dev/pkg/backend/local"
	_ "github.com/shalb/cluster.dev/pkg/backend/s3"
	_ "github.com/shalb/cluster.dev/pkg/logging"
	_ "github.com/shalb/cluster.dev/pkg/project"
	_ "github.com/shalb/cluster.dev/pkg/secrets/aws_secretmanager"
	_ "github.com/shalb/cluster.dev/pkg/secrets/sops"
	_ "github.com/shalb/cluster.dev/pkg/units/shell/common"
	_ "github.com/shalb/cluster.dev/pkg/units/shell/k8s_manifest"
	_ "github.com/shalb/cluster.dev/pkg/units/shell/terraform/helm"
	_ "github.com/shalb/cluster.dev/pkg/units/shell/terraform/kubernetes"
	_ "github.com/shalb/cluster.dev/pkg/units/shell/terraform/module"
	_ "github.com/shalb/cluster.dev/pkg/units/shell/terraform/printer"
)
