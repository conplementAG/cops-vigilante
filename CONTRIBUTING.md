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

Our main branch should always contain the last stable release source code.

Process is as follows:
(1) Update all the versions to the next version (search of vX.X.X of the latest release in the source code, e.g. README.md,
Helm Chart values.yaml and chart.yaml etc.) in a release / feature branch. Do not manually edit the charts/index.yaml file, 
this should still contain older releases!
- Update the Helm repo via running these commands in that same branch:

```
cd ./charts
helm package .
helm repo index --url https://conplementag.github.io/cops-vigilante/charts --merge index.yaml .
```

(2) Commit the new *.tgz file and the changes to index.yaml. These file, once merged to main branch, will be published via GitHub pages
automatically.
(3) Merge via PR. Once the index.yaml is updated in main, the new Helm chart become available. 
(4) Tag the main branch with the vX.X.X version as in step (1), and push the tag. This will publish the Docker container to the 
Docker hub.

We attempted to automate this process via https://github.com/helm/chart-releaser-action at one point, but for some reason it 
never worked for us. 