group "default" {
  targets = ["cdk-ecr-deployment-go"]
}

target "cdk-ecr-deployment-go" {
  context = "."
  dockerfile = "hack/Dockerfile"
  output = ["type=local,dest=./hack/generated"]
  args = {
	ROOT_GO_MOD_PATH = "github.com/walteh/cdk-ecr-deployment-go/hack/generated"
  }
}

