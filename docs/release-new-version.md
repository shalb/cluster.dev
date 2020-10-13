# Release new cluster.dev version

1. Create new branch or make up-to-date working branch with main branch.
2. Replace all needed mentions of previous version by new one using your code editor.  
**Required:** change `image` version in [action.yaml](/action.yml) file. It will be used by Github Action.
3. Add version bumping commit.
4. After merging PR, [create new release](https://github.com/shalb/cluster.dev/releases/new) from the main branch.  
In the release, describe all changes made from the previous release.
5. Check that docker image successfully builded [here](https://github.com/shalb/cluster.dev/actions?query=workflow%3ADocker).
