package main

import (
	"fmt"
	"github.com/conplementag/cops-hq/v2/pkg/cli"
	"github.com/conplementag/cops-hq/v2/pkg/commands"
	"github.com/conplementag/cops-hq/v2/pkg/error_handling"
	copshq "github.com/conplementag/cops-hq/v2/pkg/hq"
	helmhq "github.com/conplementag/cops-hq/v2/pkg/recipes/helm"
	"github.com/conplementag/cops-vigilante/pkg/vigilante/helpers"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"time"
)

const VigilanteImageName = "cops-vigilante"

func main() {
	error_handling.PanicOnAnyError = true

	hq := copshq.New("vigilante-dev-cli", "1.0.0", "vigilante-dev-cli.log")

	createCommands(hq)
	hq.Run()
}

func createCommands(hq copshq.HQ) {
	buildCommand := hq.GetCli().AddBaseCommand("build", "Builds and pushes the Docker image.",
		"Builds and pushes the Docker image to the configured Docker registry.",
		func() {
			build(hq)
		})

	addAzureContainerRegistryParameter(buildCommand)

	deployCommand := hq.GetCli().AddBaseCommand("deploy", "Deploy using the Helm chart",
		"Deploys the application using the Helm chart.",
		func() {
			deploy(hq)
		})

	addAzureContainerRegistryParameter(deployCommand)
	addTagParameter(deployCommand)
	addNamespaceParameter(deployCommand)

	buildAndDeployCommand := hq.GetCli().AddBaseCommand("build-and-deploy", "Build, push and deploy using the Helm chart",
		"Builds, pushes and deploys the application using the Helm chart.",
		func() {
			buildAndDeploy(hq)
		})

	addAzureContainerRegistryParameter(buildAndDeployCommand)
	addNamespaceParameter(buildAndDeployCommand)

	cleanupCommand := hq.GetCli().AddBaseCommand("cleanup", "Removes the deployed resources.",
		"Removes the deployed resources from the cluster where the deployment took place.",
		func() {
			cleanup(hq)
		})

	addNamespaceParameter(cleanupCommand)
}

func build(hq copshq.HQ) (vigilanteAppVersion string) {
	acr := viper.GetString(ArgumentAzureContainerRegistry)

	executor := hq.GetExecutor()

	logrus.Info("================== Building and pushing the Vigilante app container ====================")

	logrus.Info("Logging into the configured Azure container registry...")
	executor.ExecuteTTY("az acr login -n " + acr)

	vigilanteDockerfile := filepath.Join(copshq.ProjectBasePath, "Dockerfile")
	vigilanteContext := copshq.ProjectBasePath

	// Locally built images are prefixed with local- to separate them from images built via CI
	vigilanteAppVersion = "local-" + time.Now().Format("20060102") + "-" + helpers.GenerateUniqueShortString(10)
	vigilanteAppTag := acr + "/" + VigilanteImageName + ":" + vigilanteAppVersion

	logrus.Info("Building and pushing the app container as " + vigilanteAppTag)
	executor.ExecuteTTY("docker build -f " + vigilanteDockerfile + " " + vigilanteContext + " -t " + vigilanteAppTag)
	executor.ExecuteTTY("docker push " + vigilanteAppTag)

	logrus.Info("================== Building app build and push completed for version " + vigilanteAppTag + ". ====================")
	return
}

func deploy(hq copshq.HQ) {
	executor := hq.GetExecutor()

	logrus.Info("================== Deploying the cops-vigilante ====================")
	logrus.Info("The deployment will take place in the currently configured cluster.")
	logrus.Info("Current context: ")
	executor.Execute("kubectl config current-context")

	// we need to create the namespace because cops-hq helm does not provide the functionality to do so
	namespace := viper.GetString(ArgumentNamespace)
	createNamespace(executor, namespace)

	helm := helmhq.New(hq.GetExecutor(), namespace, "cops-vigilante",
		filepath.Join(copshq.ProjectBasePath, "charts"))

	// we will take our local config and transfer it via helm to the application
	configFilePath := filepath.Join(copshq.ProjectBasePath, "config", "conf.yaml")
	configYaml, err := os.ReadFile(configFilePath)

	if err != nil {
		logrus.Error(err)
		panic("Could not parse the local config file in " + configFilePath)
	}

	configParsed := make(map[string]interface{})
	err = yaml.Unmarshal(configYaml, &configParsed)
	if err != nil {
		logrus.Error(err)
		panic("Unmarshalling of the yaml file failed. Config file location was in: " + configFilePath)
	}

	configuration := map[string]interface{}{
		"image": map[string]interface{}{
			"repository": viper.GetString(ArgumentAzureContainerRegistry) + "/" + VigilanteImageName,
			"tag":        viper.GetString(ArgumentVigilanteTag),
		},
		"config": configParsed,
	}

	helm.SetVariables(configuration)
	helm.Deploy()

	logrus.Info("==================       Deployment completed        ====================")
}

func buildAndDeploy(hq copshq.HQ) {
	vigilanteTag := build(hq)

	// run expects this parameter via command line, but we can also pass it like this
	viper.Set(ArgumentVigilanteTag, vigilanteTag)

	deploy(hq)
}

func cleanup(hq copshq.HQ) {
	logrus.Info("==================           Cleanup               ====================")
	namespace := viper.GetString(ArgumentNamespace)
	hq.GetExecutor().ExecuteLoud(fmt.Sprintf("helm delete cops-vigilante -n %s", namespace))
	hq.GetExecutor().ExecuteLoud(fmt.Sprintf("kubectl delete namespace %s", namespace))

	logrus.Info("==================        Cleanup completed        ====================")
}

func createNamespace(executor commands.Executor, namespace string) {
	namespaceYaml, err := executor.Execute(fmt.Sprintf("kubectl create namespace %s --save-config --dry-run=client -o yaml", namespace))
	if err != nil {
		panic(err)
	}

	writeToTemporaryFile("namespace.yaml", namespaceYaml)

	executor.Execute(fmt.Sprintf("kubectl apply -f %s", filepath.Join(".generated", "namespace.yaml")))
}

const ArgumentNamespace = "namespace"
const ArgumentAzureContainerRegistry = "azure-container-registry"
const ArgumentVigilanteTag = "vigilante-tag"

func addNamespaceParameter(command cli.Command) {
	command.AddParameterString(ArgumentNamespace, "", true, "n",
		"Kubernetes namespace which will be created and where the application will be deployed.")
}

func addAzureContainerRegistryParameter(command cli.Command) {
	command.AddParameterString(ArgumentAzureContainerRegistry, "", true, "r",
		"Azure container registry which will be used to push the build image.")
}

func addTagParameter(command cli.Command) {
	command.AddParameterString(ArgumentVigilanteTag, "", true, "t",
		"Vigilante image tag which will be used for deployment.")
}

func writeToTemporaryFile(fileName string, content string) {
	err := os.MkdirAll(".generated", os.ModePerm)
	if err != nil {
		panic(err)
	}

	file, err := os.Create(filepath.Join(".generated", fileName))
	if err != nil {
		panic(err)
	}

	_, err = file.WriteString(content)
	if err != nil {
		panic(err)
	}
}
