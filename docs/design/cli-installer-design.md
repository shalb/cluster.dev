# Deploy Tool Dialogs Design

We need to create a tool which will simplify installation process for most of users.

The main features:
  1. Create or reuse infrastructure repo within git hosting provider.
  2. Select and create cloud user and required permissions.
  3. Populate repo with sample files.
  4. Edit view for the first cluster config.
  5. Commit all code and install the cluster.
  6. Display credentials with output.

Design requirements:
  - cli should be ran as dialog by default
  - as an option cli could receive parameters as command line options
  - delivered as binary and as a docker container
  - ability to run on CI/CD systems
  - if possible re-usable functions to be used in Web version

### Sample Workflow

```
$ ./cluster-dev install
```
CD1: Hi, we gonna install create an infrastructure for you.  
```
! if command spawned inside existing infrastructure repo, inform about it and propose
  to skip CD1-CD4 steps, ex: delete all existing configuration [yes/NO]?
```
CD2: As this is a GitOps approach we need to start with the repo, please select your Git hosting:
```
 > GitHub
 > GitLab
 > BitBucket
! on selection credentials checked and required setting requested as user/password

```
CD3: Now we need to create Github/GitLab/Bitbucket repo, or use existing:
```
 [infrastructure] <-editable
! repo created by terraform or selected existing and tested access to it
```
CD4: So, I'm creating/using [infrastructure] repo and clone it locally:
```
! show commands:
   git clone github.com/voatsap/testrepo.git
```
CD5: Ok we need to select the Cloud Provider for your cluster:
```
 > AWS
 > Google
 > DO
! credentials checked and required setting installed with profile selection.
```
CD6: Please select existing or we would create a separate user and role for your cluster with limited permissions:
```
 [cluster-dev-user] <- editable
 ! run script to create/use user and or grant required permissions by role
```
CD7: Now I'm populating sample files to your repo:
```
!  show commands:
  cd /tmp/ &&  git clone github.com/shalb/cluster-dev-samples.git
  cp -R /tmp/cluster-dev-samples/ ~/testrepo/
```
CD8: Please enter the name for your cluster:
```
[develop] <- editable
```
CD9: Lets edit the the cluster manifest
```
! EDITOR ~/testrepo/cluster.yaml
  with predefined cloud/provider/cluster.name from previous steps
```
CD10: Can we push it or open an PR for review?
```
> Commit/Push
> Create a PR
> Exit, I'll do this manually
```
CD11: Execute and show outputs

```
Cluster is ready!
Visit: https//cluster-name-gitorg.cluster.dev for details.
```
