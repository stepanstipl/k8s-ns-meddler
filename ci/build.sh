#/usr/bin/env sh

set -x

APP="k8s-ns-meddler"
GLIDE_VERSION="v0.12.3"
PACKAGE="github.com/stepanstipl/${APP}"
PACKAGE_DIR="${GOPATH}/src/${PACKAGE}"

# Setup env
mkdir -p "${GOPATH}/src/github.com/stepanstipl"
ln -s "${WERCKER_SOURCE_DIR}" "${GOPATH}/src/${PACKAGE}"
cd "${PACKAGE_DIR}"
apk add --update openssl git

# Install Glide
wget -O- "https://github.com/Masterminds/glide/releases/download/${GLIDE_VERSION}/glide-${GLIDE_VERSION}-linux-amd64.tar.gz" | tar -xzO linux-amd64/glide > /usr/local/bin/glide
chmod +x /usr/local/bin/glide

# Install Dependencies
glide install

# Get git tag
RELEASE_TAG=$(git describe --tags --exact --match '*.*.*')
[ -z "${RELEASE_TAG}" ] && RELEASE_TAG="not set"
echo "${RELEASE_TAG}" > "${WERCKER_OUTPUT_DIR}/.release_tag"

# Build
CGO_ENABLED=0 go build  -ldflags '-s' -v "${APP}.go"

# Test
go test $(glide novendor)

# Copy binary to WERCKER_ROOT
cp "${APP}" "${WERCKER_OUTPUT_DIR}/"
