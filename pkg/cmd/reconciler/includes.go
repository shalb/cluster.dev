package reconciler

import (
	// Init logging.
	_ "github.com/shalb/cluster.dev/internal/logging"
	_ "github.com/shalb/cluster.dev/pkg/backend/do"
	_ "github.com/shalb/cluster.dev/pkg/backend/s3"
	_ "github.com/shalb/cluster.dev/pkg/modules/terraform"
	_ "github.com/shalb/cluster.dev/pkg/project"
)
