# Local development 

Local development can be done by running the code on your machines, and connecting to a remote Kubernetes cluster. 
You current kubectl context will be used as the target. 

# Remote development / deployment

For a more realistic test, you can deploy everything to a namespace of your choice (will be created for you) in a 
Kubernetes cluster (current context configured for kubectl). Since we use Helm here, there is some "orchestration" to
be done, so we wrapped it into a set of command in cmd/dev directory. Navigate there and try out some of the available 
commands, such as:
- go run . build
- go run . deploy
- go run . build-and-deploy
- go run . cleanup

The config in config/conf.yaml will be given to Helm and passed to the application. You can modify the code in dev/main.go
to set different values for the Helm values.yaml file. 

# Release process

Our main branch should always contain the last stable release source code. We host GitHub Pages from our main branch, 
which serves as a Helm repository for our chart. 

Full development / release cycle looks like follows:
- you create your dev / release branch(es), and after all the PR reviews are done, you merge to main
- then you first release the new container using GitHub actions CI process
- and now you can create a new Helm chart:
  - you create a new branch for the Helm chart changes
  - you take the new container version, write it in  helm/values.yaml file as the new "default" version
  - you increment the helm chart and application versions in the Chart.yaml file
  - from the helm directory, run 'helm package .' and then 'helm repo index --url https://conplementag.github.io/cops-hq/ --merge index.yaml .'
  - merge to main (including the new .tgz files) - this automatically updates the GitHub pages, which expose the new index.yaml and makes the new chart available
