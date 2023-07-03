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
- Update all the versions to the next version (search of vX.X.X of the latest release in the source code, e.g. README.md,
Helm Chart values.yaml and chart.yaml etc.) in a release / feature branch
- Merge via PR (CI will run as a validation gate, and also after merging to main, but nothing will be pushed or released yet)
- Tag the main branch with the next vX.X.X version, and push the tag. This will trigger both CI & CD Workflows to release
both the Docker container and the Helm chart (new version will be available via our GitHub Pages Helm repo)