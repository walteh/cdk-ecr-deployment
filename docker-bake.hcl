group "default" {
  targets = ["cdk-ecr-deployment-go"]
}

target "cdk-ecr-deployment-go" {
  context = "dockerfiles"
  dockerfile = "Dockerfile.cdk-ecr-deployment-go"
  output = ["type=local,dest=./hack/generated"]
  args = {
	ROOT_GO_MOD_PATH = "github.com/walteh/cdk-ecr-deployment-go/hack/generated"
  }
}

