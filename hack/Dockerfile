FROM oven/bun:alpine AS bun


# install git
RUN apk add --no-cache git nodejs npm go curl

# install jq
RUN apk add --no-cache jq make

WORKDIR /app

# # Get latest release, download and extract it
# RUN latest_version=$(curl -s https://api.github.com/repos/cdklabs/cdk-ecr-deployment/releases/latest | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/') && \
# 	echo "$latest_version" > /tmp/version && \
# 	curl -L "https://github.com/cdklabs/cdk-ecr-deployment/archive/refs/tags/${latest_version}.tar.gz" -o cdk-ecr-deployment.tar.gz && \
# 	tar xzf cdk-ecr-deployment.tar.gz && \
# 	rm cdk-ecr-deployment.tar.gz && \
# 	mv cdk-ecr-deployment-* cdk-ecr-deployment

WORKDIR /app/cdk-ecr-deployment

COPY . .

RUN bun install

ARG ROOT_GO_MOD_PATH

RUN jq -r '.version = "'$(cat /tmp/version)'"' package.json > tmp.$$ && mv tmp.$$ package.json
RUN jq -r '.jsii.targets.go.moduleName = "'${ROOT_GO_MOD_PATH}'"' package.json > tmp.$$ && mv tmp.$$ package.json

ENV CGO_ENABLED=0 GOOS=linux GOARCH=arm64
RUN cd lambda && make lambda OUTPUT=./out/bootstrap BUILDTAGS="exclude_graphdriver_devicemapper exclude_graphdriver_btrfs containers_image_openpgp lambda.norpc" GO_DYN_FLAGS="-ldflags='-extldflags=-static'"

RUN bun run compile

RUN bun run package:go

RUN cd ./dist/go/cdkecrdeployment && go mod tidy

FROM scratch

COPY --from=bun /app/cdk-ecr-deployment/dist/go .

