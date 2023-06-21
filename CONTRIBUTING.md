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