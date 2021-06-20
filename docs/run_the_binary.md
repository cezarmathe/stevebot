# Run the binary

- Download archive with the binary for your os (`linux`, `darwin` or `windows`)
and the architecture for your os (`amd64` or `arm64`).

`curl -fsSLOJ https://github.com/cezarmathe/stevebot/releases/download/${version}/stevebot-${version}-${os}-${arch}.tar.gz`

- Download SHA512 message digest.

`curl -fsSLOJ https://github.com/cezarmathe/stevebot/releases/download/${version}/sha512sums.txt`

- Download signature of the SHA512 message digest.

`curl -fsSLOJ https://github.com/cezarmathe/stevebot/releases/download/${version}/sha512sums.txt.minisig`

- Write public key to a file in the current directory

`printf "%s\n%s\n" "untrusted comment: minisign public key 4F5AD150363013BA" "RWS6EzA2UNFaTxOCmOarJIwPNVoEmsVe6/mUU1g27SXErPDjpEwhgbhy" | tee minisign.pub >/dev/null`

- Verify SHA512 message digest signature.

`minisign -V -m sha512sums.txt`

- Verify SHA512 message digest of the archive

`sha512sum --check --ignore-missing sha512sums.txt`

- Extract binary from archive

`tar xvf stevebot-${version}-${os}-${arch}.tar.gz`

- Create a `.env` file (see the [sample](../sample.env))

- Run the binary

