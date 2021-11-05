package main

import (
	_ "github.com/shalb/cluster.dev/pkg/backend/azurerm"
	_ "github.com/shalb/cluster.dev/pkg/backend/do"
	_ "github.com/shalb/cluster.dev/pkg/backend/gcs"
	_ "github.com/shalb/cluster.dev/pkg/backend/s3"
	_ "github.com/shalb/cluster.dev/pkg/logging"
	_ "github.com/shalb/cluster.dev/pkg/modules/shell/common"
	_ "github.com/shalb/cluster.dev/pkg/modules/shell/kubectl"
	_ "github.com/shalb/cluster.dev/pkg/modules/shell/terraform/helm"
	_ "github.com/shalb/cluster.dev/pkg/modules/shell/terraform/kubernetes"
	_ "github.com/shalb/cluster.dev/pkg/modules/shell/terraform/module"
	_ "github.com/shalb/cluster.dev/pkg/modules/shell/terraform/printer"
	_ "github.com/shalb/cluster.dev/pkg/project"
	_ "github.com/shalb/cluster.dev/pkg/secrets/aws_secretmanager"
	_ "github.com/shalb/cluster.dev/pkg/secrets/sops"
)
