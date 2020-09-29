# Cleanup

To shutdown the cluster and remove all associated resources:

1. Open `.cluster.dev/` directory in your repo.
2. In each manifest set `cluster.installed` to `false`.
3. Commit and push changes.
4. Open GitHub Action output to see the removal status.

After successful removal, you can safely delete cluster manifest file from `.cluster.dev/` directory.
