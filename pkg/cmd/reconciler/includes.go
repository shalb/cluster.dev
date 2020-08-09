package reconciler

import (
	// Init logging.
	_ "github.com/shalb/cluster.dev/internal/logging"

	// Register AWS provider, modules and provisioners.
	_ "github.com/shalb/cluster.dev/pkg/provider/aws"
	_ "github.com/shalb/cluster.dev/pkg/provider/aws/module/addons"
	_ "github.com/shalb/cluster.dev/pkg/provider/aws/module/backend"
	_ "github.com/shalb/cluster.dev/pkg/provider/aws/module/eks"
	_ "github.com/shalb/cluster.dev/pkg/provider/aws/module/minikube"
	_ "github.com/shalb/cluster.dev/pkg/provider/aws/module/route53"
	_ "github.com/shalb/cluster.dev/pkg/provider/aws/module/vpc"
	_ "github.com/shalb/cluster.dev/pkg/provider/aws/provisioner/eks"
	_ "github.com/shalb/cluster.dev/pkg/provider/aws/provisioner/minikube"
)
