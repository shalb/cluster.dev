# Cluster.dev vs. Pulumi and Crossplane

Pulumi and Crossplane are modern alternatives to Terraform. 

These are great tools and we admire alternative views on infrastructure management. 

What makes Cluster.dev different is its purpose and limitations. 
Tools like Pulumi, Crossplane, and Terraform are aimed to manage clouds - creating new instances or clusters, cloud resources like databases, and others. 
Cluster.dev however is designed to manage the whole infrastructure, including those tools as units. That means you can run Terraform, then run Pulumi, or Bash, or Ansible with variables received from Terraform, and then run Crossplane or something else. Cluster.dev is created to connect and manage all infrastructure tools. 

With infrastructure tools, users are often restricted with one-tool usage that has a specific language or DSL, whereas Cluster.dev allows to have a limitless number of units and workflow combinations between tools. 

For now, Cluster.dev has major support for Terraform only, mainly because we want to provide the best experience for the majority of users. Moreover, Terraform is a de-facto industry standard and already has many modules created by the community. 
To read more on the subject please refer to the [Cluster.dev vs. Terraform](https://docs.cluster.dev/cdev-vs-terraform/) section.

If you or your company would like to use Pulumi or Crossplane with Cluster.dev, please feel free to [contact us](https://calendly.com/clusterdev/30min?month=2021-11). 
