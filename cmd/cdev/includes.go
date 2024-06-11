package main

import (
	_ "github.com/shalb/cluster.dev/internal/backend/azurerm"
	_ "github.com/shalb/cluster.dev/internal/backend/gcs"
	_ "github.com/shalb/cluster.dev/internal/backend/local"
	_ "github.com/shalb/cluster.dev/internal/backend/s3"
	_ "github.com/shalb/cluster.dev/internal/project"
	_ "github.com/shalb/cluster.dev/internal/secrets/aws_secretmanager"
	_ "github.com/shalb/cluster.dev/internal/secrets/sops"
	_ "github.com/shalb/cluster.dev/internal/units/shell/common"
	_ "github.com/shalb/cluster.dev/internal/units/shell/k8s_manifest"
	_ "github.com/shalb/cluster.dev/internal/units/shell/terraform/helm"
	_ "github.com/shalb/cluster.dev/internal/units/shell/terraform/kubernetes"
	_ "github.com/shalb/cluster.dev/internal/units/shell/terraform/module"
	_ "github.com/shalb/cluster.dev/internal/units/shell/terraform/printer"
	_ "github.com/shalb/cluster.dev/pkg/logging"
)
