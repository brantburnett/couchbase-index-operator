OPERATOR_SDK_VERSION=${1:-"0.11.0"}
OPERATOR_SDK_KEY_SERVER=${2:-"keyserver.ubuntu.com"}
OPERATOR_SDK_GPG_KEY=${3:-"052996E2A20B5C7E"}

# Download the operator-sdk and signature and checksum
OPERATOR_SDK_DL_URL=https://github.com/operator-framework/operator-sdk/releases/download/v1.11.0
curl -LO ${OPERATOR_SDK_DL_URL}/operator-sdk_linux_amd64
curl -LO ${OPERATOR_SDK_DL_URL}/checksums.txt
curl -LO ${OPERATOR_SDK_DL_URL}/checksums.txt.asc

# Validate the signature and checksum
gpg --keyserver ${OPERATOR_SDK_KEY_SERVER} --recv-keys ${OPERATOR_SDK_GPG_KEY}
gpgv --keyring ~/.gnupg/pubring.kbx checksums.txt.asc checksums.txt
grep operator-sdk_linux_amd64 checksums.txt | sha256sum -c -
rm -f checksums.txt checksums.txt.asc

# Move to bin
chmod +x operator-sdk_linux_amd64
mv operator-sdk_linux_amd64 /usr/local/bin/operator-sdk
